/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	// FUCK this logging system. FUCK IT. Marked for death!!!
	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/pubsub"

	"golang.org/x/sys/unix"

	"cloud.google.com/go/compute/metadata"
	googlepubsub "cloud.google.com/go/pubsub"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
	keys          middleware.JWKS
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "server_backend"
	fmt.Printf("\n%s\n\n", serviceName)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	isDebug, err := envvar.GetBool("NEXT_DEBUG", false)
	if err != nil {
		core.Error("could not parse NEXT_DEBUG: %v", err)
		isDebug = false
	}

	if isDebug {
		core.Debug("running as debug")
	}

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("could not get env: %v", err)
		return 1
	}

	// FUCK THIS LOGGING SYSTEM!!!
	logger := log.NewNopLogger()

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("could not get metrics handler: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("could not initialize stackdriver profiler: %v", err)
			return 1
		}
	}

	backendMetrics, err := metrics.NewServerBackendMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("could not create backend metrics: %v", err)
		return 1
	}

	maxmindSyncMetrics, err := metrics.NewMaxmindSyncMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("could not max mind sync metrics: %v", err)
		return 1
	}

	if !envvar.Exists("SERVER_BACKEND_PRIVATE_KEY") {
		core.Error("SERVER_BACKEND_PRIVATE_KEY not set")
		return 1
	}

	privateKey, err := envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", nil)
	if err != nil {
		core.Error("invalid SERVER_BACKEND_PRIVATE_KEY: %v", err)
		return 1
	}

	if !envvar.Exists("RELAY_ROUTER_PRIVATE_KEY") {
		core.Error("RELAY_ROUTER_PRIVATE_KEY not set")
		return 1
	}

	routerPrivateKeySlice, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	if err != nil {
		core.Error("invalid RELAY_ROUTER_PRIVATE_KEY: %v", err)
		return 1
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	maxmindCityFile := envvar.Get("MAXMIND_CITY_DB_FILE", "")
	if maxmindCityFile == "" {
		core.Error("could not get maxmind city file")
		return 1
	}

	maxmindISPFile := envvar.Get("MAXMIND_ISP_DB_FILE", "")
	if maxmindISPFile == "" {
		core.Error("could not get maxmind isp file")
		return 1
	}

	// function to get mmdb under mutex

	mmdb := &routing.MaxmindDB{
		CityFile:  maxmindCityFile,
		IspFile:   maxmindISPFile,
		IsStaging: env == "staging",
	}

	if err := mmdb.Sync(ctx, maxmindSyncMetrics); err != nil {
		core.Error("could not open maxmind city/isp files: %v", err)
		return 1
	}

	var mmdbMutex sync.RWMutex

	getIPLocator := func() *routing.MaxmindDB {
		mmdbMutex.RLock()
		mmdbRet := mmdb
		mmdbMutex.RUnlock()
		return mmdbRet
	}

	// Sync mmdb
	{
		maxmindSyncInterval, err := envvar.GetDuration("MAXMIND_SYNC_DB_INTERVAL", time.Minute*1)
		if err != nil {
			maxmindSyncInterval = time.Minute * 1
		}

		go func() {
			ticker := time.NewTicker(maxmindSyncInterval)

			for {
				select {
				case <-ticker.C:
					// Load the new MMDB
					newMMDB := &routing.MaxmindDB{
						CityFile:  maxmindCityFile,
						IspFile:   maxmindISPFile,
						IsStaging: env == "staging",
					}

					if err := newMMDB.Sync(ctx, maxmindSyncMetrics); err != nil {
						core.Error("could not update maxmind db: %v", err)
						continue
					}

					// Pointer swap under mutex
					mmdbMutex.Lock()
					mmdb.CloseCity()
					mmdb.CloseISP()
					mmdb = newMMDB
					mmdbMutex.Unlock()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	staleDuration, err := envvar.GetDuration("MATRIX_STALE_DURATION", 20*time.Second)
	if err != nil {
		core.Error("invalid MATRIX_STALE_DURATION: %v", err)
	}

	// function to get the route matrix pointer under mutex

	routeMatrix := &routing.RouteMatrix{}

	var routeMatrixMutex sync.RWMutex

	getRouteMatrix := func() *routing.RouteMatrix {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
	}

	// function to get the database under mutex

	database := routing.CreateEmptyDatabaseBinWrapper()

	var databaseMutex sync.RWMutex

	getDatabase := func() *routing.DatabaseBinWrapper {
		databaseMutex.RLock()
		db := database
		databaseMutex.RUnlock()
		return db
	}

	// function to clear route matrix and database atomically

	clearEverything := func() {
		routeMatrixMutex.RLock()
		databaseMutex.RLock()
		database = routing.CreateEmptyDatabaseBinWrapper()
		routeMatrix = &routing.RouteMatrix{}
		databaseMutex.RUnlock()
		routeMatrixMutex.RUnlock()
	}

	// Sync route matrix
	{
		uri := envvar.Get("ROUTE_MATRIX_URI", "")

		if uri == "" {
			core.Error("ROUTE_MATRIX_URI not set")
			return 1
		}

		syncInterval, err := envvar.GetDuration("ROUTE_MATRIX_SYNC_INTERVAL", time.Second)
		if err != nil {
			core.Error("ROUTE_MATRIX_SYNC_INTERVAL not set")
			return 1
		}

		go func() {
			httpClient := &http.Client{
				Timeout: time.Second * 4,
			}

			ticker := time.NewTicker(syncInterval)

			for {
				select {
				case <-ticker.C:

					var buffer []byte
					start := time.Now()

					var routeMatrixReader io.ReadCloser

					if f, err := os.Open(uri); err == nil {
						routeMatrixReader = f
					}

					if r, err := httpClient.Get(uri); err == nil {
						routeMatrixReader = r.Body
					}

					if routeMatrixReader == nil {
						clearEverything()
						backendMetrics.ErrorMetrics.RouteMatrixReaderNil.Add(1)
						continue
					}

					buffer, err = ioutil.ReadAll(routeMatrixReader)

					routeMatrixReader.Close()

					if err != nil {
						core.Error("faired to read route matrix data: %v", err)
						clearEverything()
						backendMetrics.ErrorMetrics.RouteMatrixReadFailure.Add(1)
						continue
					}

					if len(buffer) == 0 {
						core.Debug("route matrix buffer is empty")
						clearEverything()
						backendMetrics.ErrorMetrics.RouteMatrixBufferEmpty.Add(1)
						continue
					}

					var newRouteMatrix routing.RouteMatrix
					readStream := encoding.CreateReadStream(buffer)
					if err := newRouteMatrix.Serialize(readStream); err != nil {
						core.Error("failed to serialize route matrix: %v", err)
						clearEverything()
						backendMetrics.ErrorMetrics.RouteMatrixSerializeFailure.Add(1)
						continue
					}

					if newRouteMatrix.CreatedAt+uint64(staleDuration.Seconds()) < uint64(time.Now().Unix()) {
						core.Error("route matrix is stale")
						backendMetrics.ErrorMetrics.StaleRouteMatrix.Add(1)
						continue
					}

					routeEntriesTime := time.Since(start)
					duration := float64(routeEntriesTime.Milliseconds())
					backendMetrics.RouteMatrixUpdateDuration.Set(duration)
					if duration > 250 {
						core.Error("long route matrix duration %dms", int(duration))
						backendMetrics.RouteMatrixUpdateLongDuration.Add(1)
					}

					// update some statistics from the route matrix

					numRoutes := int32(0)
					for i := range newRouteMatrix.RouteEntries {
						numRoutes += newRouteMatrix.RouteEntries[i].NumRoutes
					}
					backendMetrics.RouteMatrixNumRoutes.Set(float64(numRoutes))
					backendMetrics.RouteMatrixBytes.Set(float64(len(buffer)))

					// decode the database in the route matrix

					var newDatabase routing.DatabaseBinWrapper

					databaseBuffer := bytes.NewBuffer(newRouteMatrix.BinFileData)
					decoder := gob.NewDecoder(databaseBuffer)
					err := decoder.Decode(&newDatabase)
					if err == io.EOF {
						core.Error("database.bin is empty")
						clearEverything()
						backendMetrics.ErrorMetrics.BinWrapperEmpty.Add(1)
						continue
					}
					if err != nil {
						core.Error("failed to read database.bin: %v", err)
						clearEverything()
						backendMetrics.ErrorMetrics.BinWrapperFailure.Add(1)
						continue
					}

					// pointer swap route matrix and database atomically

					routeMatrixMutex.Lock()
					databaseMutex.Lock()
					routeMatrix = &newRouteMatrix
					database = &newDatabase
					databaseMutex.Unlock()
					routeMatrixMutex.Unlock()
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Setup feature config for billing and vanity metrics
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BILLING2",
			Enum:        config.FEATURE_BILLING2,
			Value:       true,
			Description: "Inserts BillingEntry2 types to Google Pub/Sub",
		},
		{
			Name:        "FEATURE_VANITY_METRIC",
			Enum:        config.FEATURE_VANITY_METRIC,
			Value:       false,
			Description: "Vanity metrics for fast aggregate statistic lookup",
		},
	})
	featureConfig = envVarConfig

	featureBilling2 := featureConfig.FeatureEnabled(config.FEATURE_BILLING2)

	// Create local biller
	var biller2 billing.Biller = &billing.LocalBiller{
		Metrics: backendMetrics.BillingMetrics,
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if gcpProjectID != "" || pubsubEmulatorOK {

		pubsubCtx := ctx
		if pubsubEmulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			core.Debug("detected pubsub emulator")
		}

		// Google Pubsub
		{
			clientCount, err := envvar.GetInt("BILLING_CLIENT_COUNT", 1)
			if err != nil {
				core.Error("invalid BILLING_CLIENT_COUNT: %v", err)
				return 1
			}

			countThreshold, err := envvar.GetInt("BILLING_BATCHED_MESSAGE_COUNT", 100)
			if err != nil {
				core.Error("invalid BILLING_BATCHED_MESSAGE_COUNT: %v", err)
				return 1
			}

			byteThreshold, err := envvar.GetInt("BILLING_BATCHED_MESSAGE_MIN_BYTES", 1024)
			if err != nil {
				core.Error("invalid BILLING_BATCHED_MESSAGE_MIN_BYTES: %v", err)
				return 1
			}

			// todo: why don't we remove our batching, and just use theirs instead? less code = less problems...

			// We do our own batching so don't stack the library's batching on top of ours
			// Specifically, don't stack the message count thresholds
			settings := googlepubsub.DefaultPublishSettings
			settings.CountThreshold = 1
			settings.ByteThreshold = byteThreshold
			settings.NumGoroutines = runtime.GOMAXPROCS(0)

			if featureBilling2 {
				billing2TopicID := envvar.Get("FEATURE_BILLING2_TOPIC_NAME", "billing2")

				pubsub, err := billing.NewGooglePubSubBiller(pubsubCtx, backendMetrics.BillingMetrics, gcpProjectID, billing2TopicID, clientCount, countThreshold, byteThreshold, &settings)
				if err != nil {
					core.Error("could not create pubsub biller2: %v", err)
					return 1
				}

				biller2 = pubsub
			}
		}
	}

	// Start portal cruncher publisher
	portalPublishers := make([]pubsub.Publisher, 0)
	{
		portalCruncherHosts := envvar.GetList("PORTAL_CRUNCHER_HOSTS", []string{"tcp://127.0.0.1:5555"})

		postSessionPortalSendBufferSize, err := envvar.GetInt("POST_SESSION_PORTAL_SEND_BUFFER_SIZE", 1000000)
		if err != nil {
			core.Error("invalid POST_SESSION_PORTAL_SEND_BUFFER_SIZE: %v", err)
			return 1
		}

		for _, host := range portalCruncherHosts {
			portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(host, postSessionPortalSendBufferSize)
			if err != nil {
				core.Error("could not create portal cruncher publisher: %v", err)
				return 1
			}

			portalPublishers = append(portalPublishers, portalCruncherPublisher)
		}
	}

	numPostSessionGoroutines, err := envvar.GetInt("POST_SESSION_THREAD_COUNT", 1000)
	if err != nil {
		core.Error("invalid POST_SESSION_THREAD_COUNT: %v", err)
		return 1
	}

	postSessionBufferSize, err := envvar.GetInt("POST_SESSION_BUFFER_SIZE", 1000000)
	if err != nil {
		core.Error("invalid POST_SESSION_BUFFER_SIZE: %v", err)
		return 1
	}

	postSessionPortalMaxRetries, err := envvar.GetInt("POST_SESSION_PORTAL_MAX_RETRIES", 10)
	if err != nil {
		core.Error("invalid POST_SESSION_PORTAL_MAX_RETRIES: %v", err)
		return 1
	}

	// Determine if should use vanity metrics
	useVanityMetrics := featureConfig.FeatureEnabled(config.FEATURE_VANITY_METRIC)

	// Start vanity metrics publisher
	vanityPublishers := make([]pubsub.Publisher, 0)
	{
		vanityMetricHosts := envvar.GetList("FEATURE_VANITY_METRIC_HOSTS", []string{"tcp://127.0.0.1:6666"})

		postVanityMetricSendBufferSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_POST_SEND_BUFFER_SIZE", 1000000)
		if err != nil {
			core.Error("invalid FEATURE_VANITY_METRIC_POST_SEND_BUFFER_SIZE: %v", err)
			return 1
		}

		for _, host := range vanityMetricHosts {
			vanityPublisher, err := pubsub.NewVanityMetricPublisher(host, postVanityMetricSendBufferSize)
			if err != nil {
				core.Error("could not create vanity metric publisher: %v", err)
				return 1
			}

			vanityPublishers = append(vanityPublishers, vanityPublisher)
		}
	}

	postVanityMetricMaxRetries, err := envvar.GetInt("FEATURE_VANITY_METRIC_POST_MAX_RETRIES", 10)
	if err != nil {
		core.Error("invalid FEATURE_VANITY_METRIC_POST_MAX_RETRIES: %v", err)
		return 1
	}

	// Create a post session handler to handle the post process of session updates.
	// This way, we can quickly return from the session update handler and not spawn a
	// ton of goroutines if things get backed up.
	var wgPostSession sync.WaitGroup
	postSessionHandler := transport.NewPostSessionHandler(numPostSessionGoroutines, postSessionBufferSize, portalPublishers, postSessionPortalMaxRetries, vanityPublishers, postVanityMetricMaxRetries, useVanityMetrics, biller2, featureBilling2, backendMetrics.PostSessionMetrics)
	go postSessionHandler.StartProcessing(ctx, &wgPostSession)

	// Create a server tracker to keep track of which servers are sending updates to this backend
	serverTracker := storage.NewServerTracker()

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	if err != nil {
		core.Error("could not create local multipath veto handler: %v", err)
		return 1
	}
	var multipathVetoHandler storage.MultipathVetoHandler = localMultiPathVetoHandler

	redisMultipathVetoHost := envvar.Get("REDIS_HOST_MULTIPATH_VETO", "")
	if redisMultipathVetoHost != "" {
		redisMultipathVetoPassword := envvar.Get("REDIS_PASSWORD_MULTIPATH_VETO", "")
		redisMultipathVetoMaxIdleConns, err := envvar.GetInt("REDIS_MAX_IDLE_CONNS_MULTIPATH_VETO", 5)
		if err != nil {
			core.Error("invalid REDIS_MAX_IDLE_CONNS_MULTIPATH_VETO: %v", err)
			return 1
		}
		redisMultipathVetoMaxActiveConns, err := envvar.GetInt("REDIS_MAX_ACTIVE_CONNS_MULTIPATH_VETO", 64)
		if err != nil {
			core.Error("invalid REDIS_MAX_ACTIVE_CONNS_MULTIPATH_VETO: %v", err)
			return 1
		}

		// Create the multipath veto handler to handle syncing multipath vetoes to and from redis
		multipathVetoSyncFrequency, err := envvar.GetDuration("MULTIPATH_VETO_SYNC_FREQUENCY", time.Second*10)
		if err != nil {
			core.Error("invalid MULTIPATH_VETO_SYNC_FREQUENCY: %v", err)
			return 1
		}

		multipathVetoHandler, err = storage.NewRedisMultipathVetoHandler(redisMultipathVetoHost, redisMultipathVetoPassword, redisMultipathVetoMaxIdleConns, redisMultipathVetoMaxActiveConns, getDatabase)
		if err != nil {
			core.Error("could not create redis multipath veto handler: %v", err)
			return 1
		}

		if err := multipathVetoHandler.Sync(); err != nil {
			core.Error("faild to sync multipath veto handler: %v", err)
			return 1
		}

		// Start a routine to sync multipath vetoed users from redis to this instance
		{
			ticker := time.NewTicker(multipathVetoSyncFrequency)
			go func(ctx context.Context) {
				for {
					select {
					case <-ticker.C:
						if err := multipathVetoHandler.Sync(); err != nil {
							core.Error("faild to sync multipath veto handler: %v", err)
						}
					case <-ctx.Done():
						return
					}
				}
			}(ctx)
		}
	}

	// Fetch the Auth0 Cert and refresh occasionally
	newKeys, err := middleware.FetchAuth0Cert()
	if err != nil {
		core.Error("failed to fetch auth0 cert: %v", err)
		return 1
	}
	keys = newKeys

	fetchAuthCertInterval, err := envvar.GetDuration("AUTH_CERT_INTERVAL", time.Minute*10)
	if err != nil {
		core.Error("invalid AUTH_CERT_INTERVAL: %v", err)
		return 1
	}

	go func() {
		ticker := time.NewTicker(fetchAuthCertInterval)
		for {
			select {
			case <-ticker.C:
				newKeys, err := middleware.FetchAuth0Cert()
				if err != nil {
					continue
				}
				keys = newKeys
			case <-ctx.Done():
				return
			}

		}
	}()

	// Setup the status handler info

	statusData := &metrics.ServerBackendStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				backendMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				backendMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.ServerBackendStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(backendMetrics.ServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = backendMetrics.ServiceMetrics.MemoryAllocated.Value()

				// Server Init Metrics
				newStatusData.ServerInitInvocations = int(backendMetrics.ServerInitMetrics.HandlerMetrics.Invocations.Value())
				newStatusData.ServerInitReadPacketFailure = int(backendMetrics.ServerInitMetrics.ReadPacketFailure.Value())
				newStatusData.ServerInitBuyerNotFound = int(backendMetrics.ServerInitMetrics.BuyerNotFound.Value())
				newStatusData.ServerInitBuyerNotActive = int(backendMetrics.ServerInitMetrics.BuyerNotActive.Value())
				newStatusData.ServerInitSignatureCheckFailed = int(backendMetrics.ServerInitMetrics.SignatureCheckFailed.Value())
				newStatusData.ServerInitSDKTooOld = int(backendMetrics.ServerInitMetrics.SDKTooOld.Value())
				newStatusData.ServerInitDatacenterMapNotFound = int(backendMetrics.ServerInitMetrics.DatacenterMapNotFound.Value())
				newStatusData.ServerInitDatacenterNotFound = int(backendMetrics.ServerInitMetrics.DatacenterNotFound.Value())
				newStatusData.ServerInitWriteResponseFailure = int(backendMetrics.ServerInitMetrics.WriteResponseFailure.Value())

				// Server Update Metrics
				newStatusData.ServerUpdateInvocations = int(backendMetrics.ServerUpdateMetrics.HandlerMetrics.Invocations.Value())
				newStatusData.ServerUpdateReadPacketFailure = int(backendMetrics.ServerUpdateMetrics.ReadPacketFailure.Value())
				newStatusData.ServerUpdateBuyerNotFound = int(backendMetrics.ServerUpdateMetrics.BuyerNotFound.Value())
				newStatusData.ServerUpdateBuyerNotLive = int(backendMetrics.ServerUpdateMetrics.BuyerNotLive.Value())
				newStatusData.ServerUpdateSignatureCheckFailed = int(backendMetrics.ServerUpdateMetrics.SignatureCheckFailed.Value())
				newStatusData.ServerUpdateSDKTooOld = int(backendMetrics.ServerUpdateMetrics.SDKTooOld.Value())
				newStatusData.ServerUpdateDatacenterMapNotFound = int(backendMetrics.ServerUpdateMetrics.DatacenterMapNotFound.Value())
				newStatusData.ServerUpdateDatacenterNotFound = int(backendMetrics.ServerUpdateMetrics.DatacenterNotFound.Value())

				// Session Update Metrics
				newStatusData.SessionUpdateInvocations = int(backendMetrics.SessionUpdateMetrics.HandlerMetrics.Invocations.Value())
				newStatusData.SessionUpdateDirectSlices = int(backendMetrics.SessionUpdateMetrics.DirectSlices.Value())
				newStatusData.SessionUpdateNextSlices = int(backendMetrics.SessionUpdateMetrics.NextSlices.Value())
				newStatusData.SessionUpdateReadPacketFailure = int(backendMetrics.SessionUpdateMetrics.ReadPacketFailure.Value())
				newStatusData.SessionUpdateFallbackToDirectUnknownReason = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectUnknownReason.Value())
				newStatusData.SessionUpdateFallbackToDirectBadRouteToken = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectBadRouteToken.Value())
				newStatusData.SessionUpdateFallbackToDirectNoNextRouteToContinue = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectNoNextRouteToContinue.Value())
				newStatusData.SessionUpdateFallbackToDirectPreviousUpdateStillPending = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectPreviousUpdateStillPending.Value())
				newStatusData.SessionUpdateFallbackToDirectBadContinueToken = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectBadContinueToken.Value())
				newStatusData.SessionUpdateFallbackToDirectRouteExpired = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectRouteExpired.Value())
				newStatusData.SessionUpdateFallbackToDirectRouteRequestTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectRouteRequestTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectContinueRequestTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectContinueRequestTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectClientTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectClientTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectUpgradeResponseTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectUpgradeResponseTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectRouteUpdateTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectRouteUpdateTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectPongTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectDirectPongTimedOut.Value())
				newStatusData.SessionUpdateFallbackToDirectNextPongTimedOut = int(backendMetrics.SessionUpdateMetrics.FallbackToDirectNextPongTimedOut.Value())
				newStatusData.SessionUpdateBuyerNotFound = int(backendMetrics.SessionUpdateMetrics.BuyerNotFound.Value())
				newStatusData.SessionUpdateSignatureCheckFailed = int(backendMetrics.SessionUpdateMetrics.SignatureCheckFailed.Value())
				newStatusData.SessionUpdateClientLocateFailure = int(backendMetrics.SessionUpdateMetrics.ClientLocateFailure.Value())
				newStatusData.SessionUpdateReadSessionDataFailure = int(backendMetrics.SessionUpdateMetrics.ReadSessionDataFailure.Value())
				newStatusData.SessionUpdateBadSessionID = int(backendMetrics.SessionUpdateMetrics.BadSessionID.Value())
				newStatusData.SessionUpdateBadSliceNumber = int(backendMetrics.SessionUpdateMetrics.BadSliceNumber.Value())
				newStatusData.SessionUpdateBuyerNotLive = int(backendMetrics.SessionUpdateMetrics.BuyerNotLive.Value())
				newStatusData.SessionUpdateClientPingTimedOut = int(backendMetrics.SessionUpdateMetrics.ClientPingTimedOut.Value())
				newStatusData.SessionUpdateDatacenterMapNotFound = int(backendMetrics.SessionUpdateMetrics.DatacenterMapNotFound.Value())
				newStatusData.SessionUpdateDatacenterNotFound = int(backendMetrics.SessionUpdateMetrics.DatacenterNotFound.Value())
				newStatusData.SessionUpdateDatacenterNotEnabled = int(backendMetrics.SessionUpdateMetrics.DatacenterNotEnabled.Value())
				newStatusData.SessionUpdateNearRelaysLocateFailure = int(backendMetrics.SessionUpdateMetrics.NearRelaysLocateFailure.Value())
				newStatusData.SessionUpdateNearRelaysChanged = int(backendMetrics.SessionUpdateMetrics.NearRelaysChanged.Value())
				newStatusData.SessionUpdateNoRelaysInDatacenter = int(backendMetrics.SessionUpdateMetrics.NoRelaysInDatacenter.Value())
				newStatusData.SessionUpdateRouteDoesNotExist = int(backendMetrics.SessionUpdateMetrics.RouteDoesNotExist.Value())
				newStatusData.SessionUpdateRouteSwitched = int(backendMetrics.SessionUpdateMetrics.RouteSwitched.Value())
				newStatusData.SessionUpdateNextWithoutRouteRelays = int(backendMetrics.SessionUpdateMetrics.NextWithoutRouteRelays.Value())
				newStatusData.SessionUpdateSDKAborted = int(backendMetrics.SessionUpdateMetrics.SDKAborted.Value())
				newStatusData.SessionUpdateNoRoute = int(backendMetrics.SessionUpdateMetrics.NoRoute.Value())
				newStatusData.SessionUpdateMultipathOverload = int(backendMetrics.SessionUpdateMetrics.MultipathOverload.Value())
				newStatusData.SessionUpdateLatencyWorse = int(backendMetrics.SessionUpdateMetrics.LatencyWorse.Value())
				newStatusData.SessionUpdateMispredictVeto = int(backendMetrics.SessionUpdateMetrics.MispredictVeto.Value())
				newStatusData.SessionUpdateWriteResponseFailure = int(backendMetrics.SessionUpdateMetrics.WriteResponseFailure.Value())
				newStatusData.SessionUpdateStaleRouteMatrix = int(backendMetrics.SessionUpdateMetrics.StaleRouteMatrix.Value())

				// Post Session Metrics
				newStatusData.PostSessionBillingEntries2Sent = int(backendMetrics.PostSessionMetrics.BillingEntries2Sent.Value())
				newStatusData.PostSessionBillingEntries2Finished = int(backendMetrics.PostSessionMetrics.BillingEntries2Finished.Value())
				newStatusData.PostSessionBilling2BufferFull = int(backendMetrics.PostSessionMetrics.Billing2BufferFull.Value())
				newStatusData.PostSessionPortalEntriesSent = int(backendMetrics.PostSessionMetrics.PortalEntriesSent.Value())
				newStatusData.PostSessionPortalEntriesFinished = int(backendMetrics.PostSessionMetrics.PortalEntriesFinished.Value())
				newStatusData.PostSessionPortalBufferFull = int(backendMetrics.PostSessionMetrics.PortalBufferFull.Value())
				newStatusData.PostSessionVanityMetricsSent = int(backendMetrics.PostSessionMetrics.VanityMetricsSent.Value())
				newStatusData.PostSessionVanityMetricsFinished = int(backendMetrics.PostSessionMetrics.VanityMetricsFinished.Value())
				newStatusData.PostSessionVanityBufferFull = int(backendMetrics.PostSessionMetrics.VanityBufferFull.Value())
				newStatusData.PostSessionBilling2Failure = int(backendMetrics.PostSessionMetrics.Billing2Failure.Value())
				newStatusData.PostSessionPortalFailure = int(backendMetrics.PostSessionMetrics.PortalFailure.Value())
				newStatusData.PostSessionVanityMarshalFailure = int(backendMetrics.PostSessionMetrics.VanityMarshalFailure.Value())
				newStatusData.PostSessionVanityTransmitFailure = int(backendMetrics.PostSessionMetrics.VanityTransmitFailure.Value())

				// Billing Metrics
				newStatusData.BillingEntries2Submitted = int(backendMetrics.BillingMetrics.Entries2Submitted.Value())
				newStatusData.BillingEntries2Queued = int(backendMetrics.BillingMetrics.Entries2Queued.Value())
				newStatusData.BillingEntries2Flushed = int(backendMetrics.BillingMetrics.Entries2Flushed.Value())
				newStatusData.Billing2PublishFailure = int(backendMetrics.BillingMetrics.ErrorMetrics.Billing2PublishFailure.Value())

				// Route Matrix Metrics
				newStatusData.RouteMatrixNumRoutes = int(backendMetrics.RouteMatrixNumRoutes.Value())
				newStatusData.RouteMatrixBytes = int(backendMetrics.RouteMatrixBytes.Value())

				// Error Metrics
				newStatusData.RouteMatrixReaderNil = int(backendMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
				newStatusData.RouteMatrixReadFailure = int(backendMetrics.ErrorMetrics.RouteMatrixReadFailure.Value())
				newStatusData.RouteMatrixBufferEmpty = int(backendMetrics.ErrorMetrics.RouteMatrixBufferEmpty.Value())
				newStatusData.RouteMatrixSerializeFailure = int(backendMetrics.ErrorMetrics.RouteMatrixSerializeFailure.Value())
				newStatusData.BinWrapperEmpty = int(backendMetrics.ErrorMetrics.BinWrapperEmpty.Value())
				newStatusData.BinWrapperFailure = int(backendMetrics.ErrorMetrics.BinWrapperFailure.Value())
				newStatusData.StaleRouteMatrix = int(backendMetrics.ErrorMetrics.StaleRouteMatrix.Value())

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

	// Start HTTP server
	{
		allowedOrigins := envvar.Get("ALLOWED_ORIGINS", "")
		if allowedOrigins == "" {
			core.Debug("unable to parse ALLOWED_ORIGINS environment variable")
		}

		audience := envvar.Get("JWT_AUDIENCE", "")
		if audience == "" {
			core.Error("unable to parse JWT_AUDIENCE environment variable")
		}

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.Handle("/debug/vars", expvar.Handler())
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")

		serverTrackerHandler := http.HandlerFunc(transport.ServerTrackerHandlerFunc(serverTracker))
		router.Handle("/servers", middleware.PlainHttpAuthMiddleware(keys, audience, serverTrackerHandler, strings.Split(allowedOrigins, ",")))

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("invalid FEATURE_ENABLE_PPROF: %v", err)
		}
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		httpPort := envvar.Get("HTTP_PORT", "40001")

		srv := &http.Server{
			Addr:    ":" + httpPort,
			Handler: router,
		}

		go func() {
			fmt.Printf("started http server on port %s\n\n", httpPort)
			err := srv.ListenAndServe()
			if err != nil {
				core.Error("failed to start http server: %v", err)
				return
			}
		}()

		if gcpProjectID != "" {
			metadataSyncInterval, err := envvar.GetDuration("METADATA_SYNC_INTERVAL", time.Minute*1)
			if err != nil {
				core.Error("invalid METADATA_SYNC_INTERVAL: %v", err)
				return 1
			}
			connectionDrainMetadata := envvar.Get("CONNECTION_DRAIN_METADATA_FIELD", "connection-drain")

			// Start a goroutine to shutdown the HTTP server when the metadata changes
			go func() {
				for {
					ticker := time.NewTicker(metadataSyncInterval)
					select {
					case <-ticker.C:
						// Get metadata value for connection drain
						val, err := metadata.InstanceAttributeValue(connectionDrainMetadata)
						if err != nil {
							core.Error("failed to get instance attribute value for connection drain metadata field %s: %v", connectionDrainMetadata, err)
						}

						if val == "true" {
							core.Debug("connection drain metadata field %s is true, shutting down HTTP server", connectionDrainMetadata)
							// Shutdown the HTTP server
							ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
							defer cancel()
							srv.Shutdown(ctxTimeout)
						}
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	numThreads, err := envvar.GetInt("NUM_THREADS", 1)
	if err != nil {
		core.Error("invalid NUM_THREADS: %v", err)
		return 1
	}

	readBuffer, err := envvar.GetInt("READ_BUFFER", 100000)
	if err != nil {
		core.Error("invalid READ_BUFFER: %v", err)
		return 1
	}

	writeBuffer, err := envvar.GetInt("WRITE_BUFFER", 100000)
	if err != nil {
		core.Error("invalid WRITE_BUFFER: %v", err)
		return 1
	}

	udpPort := envvar.Get("UDP_PORT", "40000")

	var wg sync.WaitGroup

	wg.Add(numThreads)

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse address socket option: %v", err))
				}

				err = unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})

			return err
		},
	}

	serverInitHandler := transport.ServerInitHandlerFunc(getDatabase, serverTracker, backendMetrics.ServerInitMetrics)
	serverUpdateHandler := transport.ServerUpdateHandlerFunc(getDatabase, postSessionHandler, serverTracker, backendMetrics.ServerUpdateMetrics)
	sessionUpdateHandler := transport.SessionUpdateHandlerFunc(getIPLocator, getRouteMatrix, multipathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, backendMetrics.SessionUpdateMetrics, staleDuration)

	for i := 0; i < numThreads; i++ {
		go func(thread int) {

			lp, err := lc.ListenPacket(ctx, "udp", "0.0.0.0:"+udpPort)
			if err != nil {
				panic(fmt.Sprintf("could not bind socket: %v", err))
			}

			conn := lp.(*net.UDPConn)
			defer conn.Close()

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection read buffer size: %v", err))
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection write buffer size: %v", err))
			}

			dataArray := [transport.DefaultMaxPacketSize]byte{}
			for {
				data := dataArray[:]
				size, fromAddr, err := conn.ReadFromUDP(data)
				if err != nil {
					core.Error("failed to read udp packet: %v", err)
					break
				}

				if size <= 0 {
					continue
				}

				data = data[:size]

				// Check the packet hash is legit and remove the hash from the beginning of the packet
				// to continue processing the packet as normal
				if !crypto.IsNetworkNextPacket(crypto.PacketHashKey, data) {
					continue
				}

				packetType := data[0]
				data = data[crypto.PacketHashSize+1 : size]

				var buffer bytes.Buffer
				packet := transport.UDPPacket{From: *fromAddr, Data: data}

				switch packetType {
				case transport.PacketTypeServerInitRequest:
					serverInitHandler(&buffer, &packet)
				case transport.PacketTypeServerUpdate:
					serverUpdateHandler(&buffer, &packet)
				case transport.PacketTypeSessionUpdate:
					sessionUpdateHandler(&buffer, &packet)
				}

				if buffer.Len() > 0 {
					response := buffer.Bytes()

					// Sign and hash the response
					response = crypto.SignPacket(privateKey, response)
					crypto.HashPacket(crypto.PacketHashKey, response)

					if _, err := conn.WriteToUDP(response, fromAddr); err != nil {
						core.Error("failed to write udp response packet: %v", err)
					}
				}
			}

			wg.Done()
		}(i)
	}

	fmt.Printf("started udp server on port %s\n\n", udpPort)

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	fmt.Println("Received shutdown signal.")

	ctxCancelFunc()
	// Wait for essential post session goroutines to finish up
	wgPostSession.Wait()

	fmt.Println("Successfully shutdown.")

	return 0
}
