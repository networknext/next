package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v36/github"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"golang.org/x/oauth2"
	"gopkg.in/auth0.v4/management"

	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
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
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "portal"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()

	// Setup the service
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		return 1
	}

	// Get redis connections
	redisHostname := envvar.Get("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.Get("REDIS_PASSWORD", "")
	redisMaxIdleConns, err := envvar.GetInt("REDIS_MAX_IDLE_CONNS", 5)
	if err != nil {
		core.Error("failed to parse REDIS_MAX_IDLE_CONNS: %v", err)
		return 1
	}
	redisMaxActiveConns, err := envvar.GetInt("REDIS_MAX_ACTIVE_CONNS", 64)
	if err != nil {
		core.Error("failed to parse REDIS_MAX_ACTIVE_CONNS: %v", err)
		return 1
	}

	redisPoolTopSessions := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolTopSessions); err != nil {
		core.Error("failed to validate redis pool for top sessions: %v", err)
		return 1
	}

	redisPoolSessionMap := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionMap); err != nil {
		core.Error("failed to validate redis pool for session map: %v", err)
		return 1
	}

	redisPoolSessionMeta := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionMeta); err != nil {
		core.Error("failed to validate redis pool for session meta: %v", err)
		return 1
	}

	redisPoolSessionSlices := storage.NewRedisPool(redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns)
	if err := storage.ValidateRedisPool(redisPoolSessionSlices); err != nil {
		core.Error("failed to validate redis pool for session slices: %v", err)
		return 1
	}

	// Get Auth0 config
	auth0Issuer := envvar.Get("AUTH0_ISSUER", "")
	if auth0Issuer == "" {
		core.Error("AUTH0_ISSUER not set")
		return 1
	}

	auth0Domain := envvar.Get("AUTH0_DOMAIN", "")
	if auth0Domain == "" {
		core.Error("AUTH0_DOMAIN not set")
		return 1
	}

	auth0ClientID := envvar.Get("AUTH0_CLIENTID", "")
	if auth0ClientID == "" {
		core.Error("AUTH0_CLIENTID not set")
		return 1
	}

	auth0ClientSecret := envvar.Get("AUTH0_CLIENTSECRET", "")

	manager, err := management.New(
		auth0Domain,
		auth0ClientID,
		auth0ClientSecret,
	)
	if err != nil {
		core.Error("failed to create Auth0 manager: %v", err)
		return 1
	}

	var jobManager storage.JobManager = manager.Job
	var roleManager storage.RoleManager = manager.Role
	var userManager storage.UserManager = manager.User

	db, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		core.Error("failed to create storer: %v", err)
		return 1
	}

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
		core.Debug("detected bigtable emulator host")
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
		btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
		if err != nil {
			core.Error("failed to create bigtable admin: %v", err)
			return 1
		}

		// Check if the table exists in the instance
		tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
		if err != nil {
			core.Error("failed to verify if bigtable table %s exists: %v", btTableName, err)
			return 1
		}

		if !tableExists {
			core.Error("Table %s does not exist in Bigtable instance. Create the table before starting the portal", btTableName)
			return 1
		}

		// Close the admin client
		if err = btAdmin.Close(); err != nil {
			core.Error("failed to close the bigtable admin: %v", err)
			return 1
		}

		// Create a standard client for writing to the table
		btClient, err = storage.NewBigTable(ctx, gcpProjectID, btInstanceID, btTableName)
		if err != nil {
			core.Error("failed to create bigtable client: %v", err)
			return 1
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	btMetrics, err := metrics.NewBigTableMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create bigtable metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// if env == "local" {
	// 	if err = storage.SeedStorage(ctx, db, relayPublicKey, customerID, customerPublicKey); err != nil {
	// 		level.Error(logger).Log("err", err)
	// 		return 1
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
	// 			return 1
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
		core.Error("failed to create service metrics: %v", err)
		return 1
	}

	lookerSecret := envvar.Get("LOOKER_SECRET", "")
	if lookerSecret == "" {
		core.Error("LOOKER_SECRET not set")
		return 1
	}

	githubAccessToken := envvar.Get("GITHUB_ACCESS_TOKEN", "")
	if githubAccessToken == "" {
		core.Error("GITHUB_ACCESS_TOKEN not set")
		return 1
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient := github.NewClient(tc)

	webHookUrl := envvar.Get("SLACK_WEBHOOK_URL", "")
	if webHookUrl == "" {
		core.Error("SLACK_WEBHOOK_URL not set")
		return 1
	}

	channel := envvar.Get("SLACK_CHANNEL", "")
	if channel == "" {
		core.Error("SLACK_CHANNEL not set")
		return 1
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
		core.Error("failed to create authentication client: %v", err)
		return 1
	}

	// Create auth service
	authservice := &jsonrpc.AuthService{
		AuthenticationClient: authenticationClient,
		HubSpotClient:        hubspotClient,
		MailChimpManager: notifications.MailChimpHandler{
			HTTPHandler: *http.DefaultClient,
			MembersURI:  fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members", MAILCHIMP_SERVER_PREFIX, MAILCHIMP_LIST_ID),
		},
		JobManager:  jobManager,
		RoleManager: roleManager,
		UserManager: userManager,
		SlackClient: slackClient,
		Storage:     db,
	}

	// Create buyer service
	buyerService := jsonrpc.BuyersService{
		UseBigtable:            useBigtable,
		BigTableCfName:         btCfName,
		BigTable:               btClient,
		BigTableMetrics:        btMetrics,
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

	// Setup error channel with wait group to exit from goroutines
	errChan := make(chan error, 1)
	wg := &sync.WaitGroup{}

	newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
	if err != nil {
		core.Error("failed to fetch auth0 cert: %v", err)
		return 1
	}
	keys = newKeys

	fetchAuthCertInterval, err := envvar.GetDuration("AUTH0_CERT_INTERVAL", time.Minute*10)
	if err != nil {
		core.Error("failed to parse AUTH0_CERT_INTERVAL: %v", err)
		return 1
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(fetchAuthCertInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
				if err != nil {
					core.Error("failed to fetch auth0 cert: %v", err)
					continue
				}
				keys = newKeys
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Hour)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := authservice.RefreshAuthRolesCache()
				if err != nil {
					core.Error("failed to refresh auth roles cache: %v", err)
					continue
				}
			}
		}
	}()

	mapGenInterval, err := envvar.GetDuration("SESSION_MAP_INTERVAL", time.Second*1)
	if err != nil {
		core.Error("failed to parse SESSION_MAP_INTERVAL: %v", err)
		return 1
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(mapGenInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := buyerService.GenerateMapPointsPerBuyer(ctx); err != nil {
					core.Error("failed to generate session map points")
					errChan <- err
					return
				}
			}
		}
	}()

	fetchReleaseNotesInterval, err := envvar.GetDuration("RELEASE_NOTES_INTERVAL", time.Second*30)
	if err != nil {
		core.Error("failed to parse RELEASE_NOTES_INTERVAL: %v", err)
		return 1
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(fetchReleaseNotesInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := buyerService.FetchReleaseNotes(ctx); err != nil {
					core.Error("failed to fetch today's release notes: %v", err)
				}
			}
		}
	}()

	// Setup the status handler info
	statusData := &metrics.PortalStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				newStatusData := &metrics.PortalStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = runtime.NumGoroutine()
				newStatusData.MemoryAllocated = memoryUsed()

				// Bigtable Counts
				newStatusData.ReadMetaSuccessCount = int(btMetrics.ReadMetaSuccessCount.Value())
				newStatusData.ReadSliceSuccessCount = int(btMetrics.ReadSliceSuccessCount.Value())

				// Bigtable Errors
				newStatusData.ReadMetaFailureCount = int(btMetrics.ReadMetaFailureCount.Value())
				newStatusData.ReadSliceFailureCount = int(btMetrics.ReadSliceFailureCount.Value())

				// BuyerEndpoint Errors
				newStatusData.NoSlicesFailure = int(serviceMetrics.NoSlicesFailure.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			core.Error("could not write status data to json: %v\n%+v", err, data)
			w.WriteHeader(http.StatusInternalServerError)
		}
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

	port := envvar.Get("PORT", "")
	if port == "" {
		core.Error("PORT not set")
		return 1
	}

	// Start HTTP Server
	{
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

					var verified bool
					if emailVerified, ok := requestData.(map[string]interface{})["verified"]; ok {
						verified = emailVerified.(bool)
					}
					return middleware.AddTokenContext(i.Request, userRoles, companyCode, newsletterConsent, verified)
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

		relayFrontEnd := envvar.Get("RELAY_FRONTEND", "")
		if relayFrontEnd == "" {
			core.Error("RELAY_FRONTEND not set")
			return 1
		}

		relayGateway := envvar.Get("RELAY_GATEWAY", "")
		if relayGateway == "" {
			core.Error("RELAY_GATEWAY not set")
			return 1
		}

		relayForwarder := envvar.Get("RELAY_FORWARDER", "")
		if relayForwarder == "" {
			core.Error("RELAY_FORWARDER not set")
		}

		mondayApiKey := envvar.Get("MONDAY_API_KEY", "")
		if mondayApiKey == "" {
			core.Error("MONDAY_API_KEY not set")
			return 1
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

		allowedOrigins := envvar.Get("ALLOWED_ORIGINS", "")

		httpTimeout, err := envvar.GetDuration("HTTP_TIMEOUT", time.Second*40)
		if err != nil {
			core.Error("failed to parse HTTP_TIMEOUT: %v", err)
			return 1
		}

		r := mux.NewRouter()

		r.Handle("/rpc", middleware.HTTPAuthMiddleware(keys, envvar.GetList("JWT_AUDIENCES", []string{}), http.TimeoutHandler(s, httpTimeout, "Connection Timed Out!"), strings.Split(allowedOrigins, ","), auth0Issuer, true))
		r.HandleFunc("/health", transport.HealthHandlerFunc())
		r.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, strings.Split(allowedOrigins, ",")))
		r.HandleFunc("/status", serveStatusFunc).Methods("GET")

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("failed to parse FEATURE_ENABLE_PPROF: %v", err)
		}
		if enablePProf {
			r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		fmt.Printf("starting http server on port %s\n", port)

		go func() {
			// If the port is set to 443 then build the certificates and run a TLS-enabled HTTP server
			if port == "443" {
				cert, err := tls.X509KeyPair(transport.TLSCertificate, transport.TLSPrivateKey)
				if err != nil {
					core.Error("failed to create TSL cert: %v", err)
					errChan <- err
				}

				server := &http.Server{
					Addr:      ":" + port,
					TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
					Handler:   r,
				}

				err = server.ListenAndServeTLS("", "")
				if err != nil {
					core.Error("failed to start TLS-enabled HTTP server: %v", err)
					errChan <- err
				}
			} else {
				// Fall through to running on any other port defined with TLS disabled
				err = http.ListenAndServe(":"+port, r)
				if err != nil {
					core.Error("failed to start HTTP server: %v", err)
					errChan <- err
				}
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		fmt.Println("received shutdown signal")
		ctxCancelFunc()

		// Wait for essential goroutines to finish up
		wg.Wait()

		fmt.Println("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}
