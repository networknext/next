package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v36/github"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"golang.org/x/oauth2"
	"gopkg.in/auth0.v4/management"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/backend" // todo: not a good name for a module
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/notifications"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
	keys          middleware.JWKS
)

const (
	MAILCHIMP_SERVER_PREFIX = "us20"
	MAILCHIMP_LIST_ID       = "553903bc6f"
)

func main() {
	fmt.Printf("portal: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Stackdriver Logging
	{
		var enableSDLogging bool
		enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
		if ok {
			var err error
			enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
			if err != nil {
				level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
		}

		if enableSDLogging {
			if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
				loggingClient, err := gcplogging.NewClient(ctx, projectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}

				logger = logging.NewStackdriverLogger(loggingClient, "portal")
			}
		}
	}

	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	// Get redis connections
	redisHostname := envvar.Get("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.Get("REDIS_PASSWORD", "")
	redisMaxIdleConns, err := envvar.GetInt("REDIS_MAX_IDLE_CONNS", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	redisMaxActiveConns, err := envvar.GetInt("REDIS_MAX_ACTIVE_CONNS", 64)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	redisPoolTopSessions := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolTopSessions); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_TOP_SESSIONS", "err", err)
		os.Exit(1)
	}

	redisPoolSessionMap := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionMap); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_MAP", "err", err)
		os.Exit(1)
	}

	redisPoolSessionMeta := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionMeta); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_META", "err", err)
		os.Exit(1)
	}

	redisPoolSessionSlices := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionSlices); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_SLICES", "err", err)
		os.Exit(1)
	}

	auth0Issuer := os.Getenv("AUTH0_ISSUER")
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	auth0ClientID := os.Getenv("AUTH0_CLIENTID")
	manager, err := management.New(
		auth0Domain,
		auth0ClientID,
		os.Getenv("AUTH0_CLIENTSECRET"),
	)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	var jobManager storage.JobManager = manager.Job
	var roleManager storage.RoleManager = manager.Role
	var userManager storage.UserManager = manager.User

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")

	db, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	// switch db.(type) {
	// case *storage.SQL:
	// 	if env == "local" {
	// 		err = storage.SeedSQLStorage(ctx, db)
	// 	}
	// }

	// Setup feature config for bigtable
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BIGTABLE",
			Enum:        config.FEATURE_BIGTABLE,
			Value:       false,
			Description: "Bigtable integration for historic session data",
		},
	})
	featureConfig = envVarConfig

	// Setup Bigtable

	btEmulatorOK := envvar.Exists("BIGTABLE_EMULATOR_HOST")
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected bigtable emulator host")
	}

	useBigtable := featureConfig.FeatureEnabled(config.FEATURE_BIGTABLE) && (gcpOK || btEmulatorOK)

	var btClient *storage.BigTable
	var btCfName string
	if useBigtable {
		// Get Bigtable instance ID
		btInstanceID := envvar.Get("BIGTABLE_INSTANCE_ID", "localhost:8086")

		// Get the table name
		btTableName := envvar.Get("BIGTABLE_TABLE_NAME", "")

		// Get the column family
		btCfName = envvar.Get("BIGTABLE_CF_NAME", "")

		// Create a bigtable admin for setup
		btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Check if the table exists in the instance
		tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		if !tableExists {
			level.Error(logger).Log("table", btTableName, "err", "Table does not exist in Bigtable instance. Create the table before starting the portal")
			os.Exit(1)
		}

		// Close the admin client
		if err = btAdmin.Close(); err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a standard client for writing to the table
		btClient, err = storage.NewBigTable(ctx, gcpProjectID, btInstanceID, btTableName, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	btMetrics, err := metrics.NewBigTableMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create bigtable metrics", "err", err)
		os.Exit(1)
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpOK {
		// Stackdriver Profiler
		{
			var enableSDProfiler bool
			enableSDProfilerString, ok := os.LookupEnv("ENABLE_STACKDRIVER_PROFILER")
			if ok {
				enableSDProfiler, err = strconv.ParseBool(enableSDProfilerString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_PROFILER", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "portal",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to initialze StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	// if env == "local" {
	// 	if err = storage.SeedStorage(ctx, db, relayPublicKey, customerID, customerPublicKey); err != nil {
	// 		level.Error(logger).Log("err", err)
	// 		os.Exit(1)
	// 	}
	// }
	// We're not using the route matrix in the portal anymore, because RouteSelection()
	// is commented out in the ops service.

	// routeMatrix := &routing.RouteMatrix{}
	// var routeMatrixMutex sync.RWMutex

	// {
	// 	if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
	// 		rmsyncinterval := os.Getenv("ROUTE_MATRIX_SYNC_INTERVAL")
	// 		syncInterval, err := time.ParseDuration(rmsyncinterval)
	// 		if err != nil {
	// 			level.Error(logger).Log("envvar", "ROUTE_MATRIX_SYNC_INTERVAL", "value", rmsyncinterval, "err", err)
	// 			os.Exit(1)
	// 		}

	// 		go func() {
	// 			for {
	// 				newRouteMatrix := &routing.RouteMatrix{}
	// 				var matrixReader io.Reader

	// 				// Default to reading route matrix from file
	// 				if f, err := os.Open(uri); err == nil {
	// 					matrixReader = f
	// 				}

	// 				// Prefer to get it remotely if possible
	// 				if r, err := http.Get(uri); err == nil {
	// 					matrixReader = r.Body
	// 				}

	// 				// Don't swap route matrix if we fail to read
	// 				_, err := newRouteMatrix.ReadFrom(matrixReader)
	// 				if err != nil {
	// 					level.Warn(logger).Log("matrix", "route", "op", "read", "envvar", "ROUTE_MATRIX_URI", "value", uri, "err", err, "msg", "forcing empty route matrix to avoid stale routes")
	// 					time.Sleep(syncInterval)
	// 					continue
	// 				}

	// 				// Swap the route matrix pointer to the new one
	// 				// This double buffered route matrix approach makes the route matrix lockless
	// 				routeMatrixMutex.Lock()
	// 				routeMatrix = newRouteMatrix
	// 				routeMatrixMutex.Unlock()

	// 				time.Sleep(syncInterval)
	// 			}
	// 		}()
	// 	}
	// }

	serviceMetrics, err := metrics.NewBuyerEndpointMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create service metrics", "err", err)
		os.Exit(1)
	}

	lookerSecret, ok := os.LookupEnv("LOOKER_SECRET")
	if !ok {
		level.Error(logger).Log("err", "env var LOOKER_SECRET must be set")
		os.Exit(1)
	}

	githubAccessToken, ok := os.LookupEnv("GITHUB_ACCESS_TOKEN")
	if !ok {
		level.Error(logger).Log("err", "env var GITHUB_ACCESS_TOKEN must be set")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient := github.NewClient(tc)

	webHookUrl := envvar.Get("SLACK_WEBHOOK_URL", "")
	if webHookUrl == "" {
		level.Error(logger).Log("err", "env var SLACK_WEBHOOK_URL must be set")
		os.Exit(1)
	}

	channel := envvar.Get("SLACK_CHANNEL", "")
	if channel == "" {
		level.Error(logger).Log("err", "env var SLACK_CHANNEL must be set")
		os.Exit(1)
	}

	slackClient := notifications.SlackClient{
		WebHookUrl: webHookUrl,
		UserName:   "PortalBot",
		Channel:    channel,
	}

	// If the hubspot API key isn't set, hubspot functionality will be turned off
	hubspotAPIKey := envvar.Get("HUBSPOT_API_KEY", "")
	hubspotClient, err := notifications.NewHubSpotClient(hubspotAPIKey, 10*time.Second)

	authenticationClient, err := notifications.NewAuth0AuthClient(auth0ClientID, auth0Domain)
	if err != nil {
		level.Error(logger).Log("err", "failed to create authentication client")
		os.Exit(1)
	}

	authservice := &jsonrpc.AuthService{
		AuthenticationClient: authenticationClient,
		HubSpotClient:        hubspotClient,
		MailChimpManager: notifications.MailChimpHandler{
			HTTPHandler: *http.DefaultClient,
			MembersURI:  fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members", MAILCHIMP_SERVER_PREFIX, MAILCHIMP_LIST_ID),
		},
		Logger:      logger,
		JobManager:  jobManager,
		RoleManager: roleManager,
		UserManager: userManager,
		SlackClient: slackClient,
		Storage:     db,
	}

	// Generate Sessions Map Points periodically
	buyerService := jsonrpc.BuyersService{
		UseBigtable:            useBigtable,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        btMetrics,
		Logger:                 logger,
		RedisPoolTopSessions:   redisPoolTopSessions,
		RedisPoolSessionMeta:   redisPoolSessionMeta,
		RedisPoolSessionSlices: redisPoolSessionSlices,
		RedisPoolSessionMap:    redisPoolSessionMap,
		Storage:                db,
		Env:                    env,
		Metrics:                serviceMetrics,
		LookerSecret:           lookerSecret,
		GithubClient:           githubClient,
		SlackClient:            slackClient,
	}

	configService := jsonrpc.ConfigService{
		Logger:  logger,
		Storage: db,
	}

	newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
	if err != nil {
		level.Error(logger).Log("msg", "error fetching auth0 cert", "err", err)
		os.Exit(1)
	}
	keys = newKeys

	go func() {
		fetchAuthCertInterval, err := envvar.GetDuration("AUTH_CERT_INTERVAL", time.Minute*10)
		if err != nil {
			level.Error(logger).Log("envvar", "AUTH_CERT_INTERVAL", "value", fetchAuthCertInterval, "err", err)
			os.Exit(1)
		}

		for {
			newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
			if err != nil {
				continue
			}
			keys = newKeys
			time.Sleep(fetchAuthCertInterval)
		}
	}()

	go func() {
		for {
			err := authservice.RefreshAuthRolesCache()
			if err != nil {
				continue
			}
			time.Sleep(time.Hour)
		}
	}()

	go func() {
		mapGenInterval, err := envvar.GetDuration("SESSION_MAP_INTERVAL", time.Second*1)
		if err != nil {
			level.Error(logger).Log("envvar", "SESSION_MAP_INTERVAL", "value", mapGenInterval, "err", err)
			os.Exit(1)
		}

		for {
			if err := buyerService.GenerateMapPointsPerBuyer(ctx); err != nil {
				level.Error(logger).Log("msg", "error generating sessions map points", "err", err)
				os.Exit(1)
			}
			time.Sleep(mapGenInterval)
		}
	}()

	go func() {
		fetchReleaseNotesInterval, err := envvar.GetDuration("RELEASE_NOTES_INTERVAL", time.Second*30)
		if err != nil {
			level.Error(logger).Log("envvar", "RELEASE_NOTES_INTERVAL", "value", fetchReleaseNotesInterval, "err", err)
			os.Exit(1)
		}

		for {
			if err := buyerService.FetchReleaseNotes(ctx); err != nil {
				level.Error(logger).Log("msg", "error fetching today's release notes", "err", err)
			}
			time.Sleep(fetchReleaseNotesInterval)
		}
	}()

	uiDir := os.Getenv("UI_DIR")
	if uiDir == "" {
		level.Error(logger).Log("err", "env var UI_DIR must be set")
		os.Exit(1)
	}

	relayMap := jsonrpc.NewRelayStatsMap()

	// TODO: b0rked, needs to process a csv file from /relays and this GET
	//       needs to be auth'd...
	// go func() {
	// 	relayStatsURL := os.Getenv("RELAY_STATS_URI")
	// 	fmt.Printf("RELAY_STATS_URI: %s\n", relayStatsURL)

	// 	sleepInterval := time.Second
	// 	if siStr, ok := os.LookupEnv("RELAY_STATS_SYNC_SLEEP_INTERVAL"); ok {
	// 		if si, err := time.ParseDuration(siStr); err == nil {
	// 			sleepInterval = si
	// 		} else {
	// 			level.Error(logger).Log("msg", "could not parse stats sync sleep interval", "err", err)
	// 		}
	// 	}

	// 	for {
	// 		time.Sleep(sleepInterval)

	// 		res, err := http.Get(relayStatsURL)
	// 		if err != nil {
	// 			level.Error(logger).Log("msg", "unable to get relay stats", "err", err)
	// 			continue
	// 		}

	// 		if res.StatusCode != http.StatusOK {
	// 			level.Error(logger).Log("msg", "bad relay_stats request")
	// 			continue
	// 		}

	// 		if res.ContentLength == -1 {
	// 			level.Error(logger).Log("msg", fmt.Sprintf("relay_stats content length invalid: %d\n", res.ContentLength))
	// 			res.Body.Close()
	// 			continue
	// 		}

	// 		data, err := ioutil.ReadAll(res.Body)
	// 		if err != nil {
	// 			level.Error(logger).Log("msg", "unable to read response body", "err", err)
	// 			res.Body.Close()
	// 			continue
	// 		}
	// 		res.Body.Close()

	// 		if err := relayMap.ReadAndSwap(data); err != nil {
	// 			level.Error(logger).Log("msg", "unable to read relay stats map", "err", err)
	// 		}
	// 	}
	// }()

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("msg", fmt.Sprintf("Starting portal on port %s", port))

		s := rpc.NewServer()
		s.RegisterInterceptFunc(func(i *rpc.RequestInfo) *http.Request {
			user := i.Request.Context().Value(middleware.Keys.UserKey)
			if user != nil {
				claims := user.(*jwt.Token).Claims.(jwt.MapClaims)
				if requestData, ok := claims["https://networknext.com/userData"]; ok {
					var userRoles []string
					if roles, ok := requestData.(map[string]interface{})["roles"]; ok {
						rolesInterface := roles.([]interface{})
						userRoles = make([]string, len(rolesInterface))
						for i, v := range rolesInterface {
							userRoles[i] = v.(string)
						}
					}
					var companyCode string
					if companyCodeInterface, ok := requestData.(map[string]interface{})["company_code"]; ok {
						companyCode = companyCodeInterface.(string)
					}

					var newsletterConsent bool
					if consent, ok := requestData.(map[string]interface{})["newsletter"]; ok {
						newsletterConsent = consent.(bool)
					}
					return middleware.AddTokenContext(i.Request, userRoles, companyCode, newsletterConsent)
				}
			}
			return middleware.SetIsAnonymous(i.Request, i.Request.Header.Get("Authorization") == "")
		})
		s.RegisterCodec(json2.NewCodec(), "application/json")
		s.RegisterService(&jsonrpc.OpsService{
			Logger:    logger,
			Release:   tag,
			BuildTime: buildtime,
			Storage:   db,
			RelayMap:  &relayMap,
		}, "")
		s.RegisterService(&buyerService, "")
		s.RegisterService(&configService, "")
		s.RegisterService(authservice, "")

		relayFrontEnd, ok := os.LookupEnv("RELAY_FRONTEND")
		if !ok {
			level.Error(logger).Log("err", "RELAY_FRONTEND environment variable not set")
			os.Exit(1)
		}

		relayGateway, ok := os.LookupEnv("RELAY_GATEWAY")
		if !ok {
			level.Error(logger).Log("err", "RELAY_GATEWAY environment variable not set")
			os.Exit(1)
		}

		relayForwarder, ok := os.LookupEnv("RELAY_FORWARDER")
		if !ok {
			level.Error(logger).Log("err", "RELAY_FORWARDER environment variable not set")
			os.Exit(1)
		}

		env, ok := os.LookupEnv("ENV")
		if !ok {
			level.Error(logger).Log("err", "ENV environment variable not set")
			os.Exit(1)
		}

		mondayApiKey, ok := os.LookupEnv("MONDAY_API_KEY")
		if !ok {
			level.Error(logger).Log("err", "MONDAY_API_KEY environment variable not set")
			os.Exit(1)
		}

		s.RegisterService(&jsonrpc.RelayFleetService{
			RelayFrontendURI:  relayFrontEnd,
			RelayGatewayURI:   relayGateway,
			RelayForwarderURI: relayForwarder,
			Logger:            logger,
			Storage:           db,
			Env:               env,
			MondayApiKey:      mondayApiKey,
		}, "")

		s.RegisterService(&jsonrpc.LiveServerService{
			Logger: logger,
		}, "")

		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		httpTimeout, err := envvar.GetDuration("HTTP_TIMEOUT", time.Second*40)
		if err != nil {
			level.Error(logger).Log("envvar", "HTTP_TIMEOUT", "value", httpTimeout, "err", err)
			os.Exit(1)
		}

		r := mux.NewRouter()

		r.Handle("/rpc", middleware.JSONRPCMiddleware(keys, os.Getenv("JWT_AUDIENCE"), http.TimeoutHandler(s, httpTimeout, "Connection Timed Out!"), strings.Split(allowedOrigins, ","), auth0Issuer))
		r.HandleFunc("/health", transport.HealthHandlerFunc())
		r.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, strings.Split(allowedOrigins, ",")))

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
		if enablePProf {
			r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		level.Info(logger).Log("addr", ":"+port)

		// If the port is set to 443 then build the certificates and run a TLS-enabled HTTP server
		if port == "443" {
			cert, err := tls.X509KeyPair(transport.TLSCertificate, transport.TLSPrivateKey)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			server := &http.Server{
				Addr:      ":" + port,
				TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
				Handler:   r,
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}

		// Fall through to running on any other port defined with TLS disabled
		err = http.ListenAndServe(":"+port, r)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
