package main

import (
	"context"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

var routeMatrixURI string
var routeMatrixInterval time.Duration

func main() {

	service := common.CreateService("analytics")

	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001")
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 10 * time.Second)

	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval)

	ProcessRelayStats(service.Context)

	ProcessBilling(service.Context)

	ProcessMatchData(service.Context)

	service.StartWebServer()

	service.WaitForShutdown()
}

func ProcessRelayStats(ctx context.Context) {

	ticker := time.NewTicker(routeMatrixInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			core.Debug("get route matrix")
		}
	}
}

func ProcessBilling(ctx context.Context) {

	// todo: create google pubsub consumer

	/*
	for {
		select {

		case <-service.Context.Done():
			return

		case message := <-consumer.MessageChannel:
			// todo: process message
			core.Debug("received billing message")
			_ = message
		}
	}
	*/
}

func ProcessMatchData(ctx context.Context) {

	// todo: create google pubsub consumer

	/*
	for {
		select {

		case <-service.Context.Done():
			return

		case message := <-consumer.MessageChannel:
			// todo: process message
			core.Debug("received match data message")
			_ = message
		}
	}
	*/
}

/*
import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "analytics"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	logger := log.NewNopLogger()

	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("could not get metrics handler: %v", err)
		return 1
	}

	env := backend.GetEnv()

	if gcpOK {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("could not initialize stackdriver profiler: %v", err)
			return 1
		}
	}

	analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create analytics metrics: %v", err)
		return 1
	}

	// Create an error chan for exiting from goroutines
	errChan := make(chan error, 1)

	var pingStatsWriter analytics.PingStatsWriter = &analytics.NoOpPingStatsWriter{}
	var relayStatsWriter analytics.RelayStatsWriter = &analytics.NoOpRelayStatsWriter{}

	if gcpOK {
		// Google BigQuery
		{
			pingStatsDataset := envvar.GetString("GOOGLE_BIGQUERY_DATASET_PING_STATS", "")
			if pingStatsDataset == "" {
				core.Error("envvar GOOGLE_BIGQUERY_DATASET_PING_STATS not set")
				return 1
			}

			pingStatsTableName := envvar.GetString("GOOGLE_BIGQUERY_TABLE_PING_STATS", "")
			if pingStatsTableName == "" {
				core.Error("envvar GOOGLE_BIGQUERY_TABLE_PING_STATS not set")
				return 1
			}

			pingStatsToPublishAtOnce := envvar.GetInt("PING_STATS_TO_PUBLISH_AT_ONCE", 10000)

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				core.Error("could not create ping stats bigquery client: %v", err)
				return 1
			}

			b := analytics.NewGoogleBigQueryPingStatsWriter(bqClient, &analyticsMetrics.PingStatsMetrics, pingStatsDataset, pingStatsTableName, pingStatsToPublishAtOnce)
			pingStatsWriter = &b

			go b.WriteLoop(wg)
		}

		{
			relayStatsDataset := envvar.GetString("GOOGLE_BIGQUERY_DATASET_RELAY_STATS", "")
			if relayStatsDataset == "" {
				core.Error("envvar GOOGLE_BIGQUERY_DATASET_RELAY_STATS not set")
				return 1
			}

			relayStatsTableName := envvar.GetString("GOOGLE_BIGQUERY_TABLE_RELAY_STATS", "")
			if relayStatsTableName == "" {
				core.Error("envvar GOOGLE_BIGQUERY_TABLE_RELAY_STATS not set")
				return 1
			}

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				core.Error("could not create relay stats bigquery client: %v", err)
				return 1
			}

			b := analytics.NewGoogleBigQueryRelayStatsWriter(bqClient, &analyticsMetrics.RelayStatsMetrics, relayStatsDataset, relayStatsTableName)
			relayStatsWriter = &b

			go b.WriteLoop(wg)
		}
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")

	if gcpOK || pubsubEmulatorOK {

		if pubsubEmulatorOK {
			// Prefer to use the emulator instead of actual Google pubsub
			gcpProjectID = "local"

			// Use local ping stats and relay stats writer
			pingStatsWriter = &analytics.LocalPingStatsWriter{Metrics: &analyticsMetrics.PingStatsMetrics}
			relayStatsWriter = &analytics.LocalRelayStatsWriter{Metrics: &analyticsMetrics.RelayStatsMetrics}

			core.Debug("Detected pubsub emulator")
		}

		// Google pubsub forwarder
		{
			topicName := envvar.GetString("PING_STATS_TOPIC_NAME", "ping_stats")
			subscriptionName := envvar.GetString("PING_STATS_SUBSCRIPTION_NAME", "ping_stats")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewPingStatsPubSubForwarder(pubsubCtx, pingStatsWriter, &analyticsMetrics.PingStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				core.Error("could not create ping stats pub sub forwarder: %v", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}

		{
			topicName := envvar.GetString("RELAY_STATS_TOPIC_NAME", "relay_stats")
			subscriptionName := envvar.GetString("RELAY_STATS_SUBSCRIPTION_NAME", "relay_stats")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewRelayStatsPubSubForwarder(pubsubCtx, relayStatsWriter, &analyticsMetrics.RelayStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				core.Error("could not create relay stats pub sub forwarder: %v", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}
	}

	// Setup the status handler info
	statusData := &metrics.AnalyticsStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				analyticsMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.AnalyticsStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(analyticsMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = analyticsMetrics.MemoryAllocated.Value()
				newStatusData.PingStatsEntriesReceived = int(analyticsMetrics.PingStatsMetrics.EntriesReceived.Value())
				newStatusData.PingStatsEntriesSubmitted = int(analyticsMetrics.PingStatsMetrics.EntriesSubmitted.Value())
				newStatusData.PingStatsEntriesQueued = int(analyticsMetrics.PingStatsMetrics.EntriesQueued.Value())
				newStatusData.PingStatsEntriesFlushed = int(analyticsMetrics.PingStatsMetrics.EntriesFlushed.Value())
				newStatusData.RelayStatsEntriesReceived = int(analyticsMetrics.RelayStatsMetrics.EntriesReceived.Value())
				newStatusData.RelayStatsEntriesSubmitted = int(analyticsMetrics.RelayStatsMetrics.EntriesSubmitted.Value())
				newStatusData.RelayStatsEntriesQueued = int(analyticsMetrics.RelayStatsMetrics.EntriesQueued.Value())
				newStatusData.RelayStatsEntriesFlushed = int(analyticsMetrics.RelayStatsMetrics.EntriesFlushed.Value())

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
		port := envvar.GetString("PORT", "41001")
		if port == "" {
			core.Error("envvar PORT not set: %v", err)
			return 1
		}

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		fmt.Printf("starting http server on port %s\n", port)

		go func() {
			err = http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		ctxCancelFunc()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}
*/

/*
// analytics pusher, gets the route matrix... generates relay stats and ping stats

package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/analytics"
	pusher "github.com/networknext/backend/modules/analytics_pusher"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "analytics_pusher"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("error getting logger: %v", err)
		return 1
	}

	env := backend.GetEnv()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create analytics pusher metrics
	analyticsPusherMetrics, err := metrics.NewAnalyticsPusherMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		core.Error("failed to create analytics pusher metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Determine how frequently we should publish ping and relay stats
	pingStatsPublishInterval := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", 1*time.Minute)
	relayStatsPublishInterval := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", 10*time.Second)

	// Get HTTP timeout for route matrix
	httpTimeout := envvar.GetDuration("HTTP_TIMEOUT", 4*time.Second)

	// Get route matrix URI
	routeMatrixURI := envvar.GetString("ROUTE_MATRIX_URI", "")
	if routeMatrixURI == "" {
		core.Error("ROUTE_MATRIX_URI not set")
		return 1
	}

	// Get route matrix stale duration
	routeMatrixStaleDuration := envvar.GetDuration("ROUTE_MATRIX_STALE_DURATION", 20*time.Second)

	// Setup ping stats and relay stats publishers
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				core.Debug("Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pingPubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &analyticsPusherMetrics.PingStatsMetrics, gcpProjectID, "ping_stats", settings)
				if err != nil {
					core.Error("could not create ping stats analytics pubsub publisher: %v", err)
					return 1
				}

				pingStatsPublisher = pingPubsub

				relayPubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &analyticsPusherMetrics.RelayStatsMetrics, gcpProjectID, "relay_stats", settings)
				if err != nil {
					core.Error("could not create relay stats analytics pubsub publisher: %v", err)
					return 1
				}

				relayStatsPublisher = relayPubsub
			}
		}
	}

	// Create an error chan for exiting from goroutines
	errChan := make(chan error, 1)

	// Create a waitgroup for goroutines
	var wg sync.WaitGroup

	// Start the relay stats and ping stats goroutines
	analyticsPusher, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, relayStatsPublishInterval, pingStatsPublishInterval, httpTimeout, routeMatrixURI, routeMatrixStaleDuration, analyticsPusherMetrics)
	if err != nil {
		core.Error("could not create analytics pusher: %v", err)
		return 1
	}

	wg.Add(1)
	go analyticsPusher.Start(ctx, &wg, errChan)

	// Setup the status handler info
	statusData := &metrics.AnalyticsPusherStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				analyticsPusherMetrics.AnalyticsPusherServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsPusherMetrics.AnalyticsPusherServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.AnalyticsPusherStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(analyticsPusherMetrics.AnalyticsPusherServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = analyticsPusherMetrics.AnalyticsPusherServiceMetrics.MemoryAllocated.Value()
				newStatusData.RouteMatrixInvocations = int(analyticsPusherMetrics.RouteMatrixInvocations.Value())
				newStatusData.RouteMatrixSuccesses = int(analyticsPusherMetrics.RouteMatrixSuccess.Value())
				newStatusData.RouteMatrixDuration = int(analyticsPusherMetrics.RouteMatrixUpdateDuration.Value())
				newStatusData.RouteMatrixLongDurations = int(analyticsPusherMetrics.RouteMatrixUpdateLongDuration.Value())
				newStatusData.PingStatsEntriesReceived = int(analyticsPusherMetrics.PingStatsMetrics.EntriesReceived.Value())
				newStatusData.PingStatsEntriesSubmitted = int(analyticsPusherMetrics.PingStatsMetrics.EntriesSubmitted.Value())
				newStatusData.PingStatsEntriesFlushed = int(analyticsPusherMetrics.PingStatsMetrics.EntriesFlushed.Value())
				newStatusData.RelayStatsEntriesReceived = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesReceived.Value())
				newStatusData.RelayStatsEntriesSubmitted = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesSubmitted.Value())
				newStatusData.RelayStatsEntriesFlushed = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesFlushed.Value())
				newStatusData.PingStatusPublishFailures = int(analyticsPusherMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure.Value())
				newStatusData.RelayStatsPublishFailures = int(analyticsPusherMetrics.RelayStatsMetrics.ErrorMetrics.PublishFailure.Value())
				newStatusData.RouteMatrixReaderNilErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
				newStatusData.RouteMatrixReadErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
				newStatusData.RouteMatrixBufferEmptyErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixBufferEmpty.Value())
				newStatusData.RouteMatrixSerializeErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixSerializeFailure.Value())
				newStatusData.RouteMatrixStaleErrors = int(analyticsPusherMetrics.ErrorMetrics.StaleRouteMatrix.Value())

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

	// Start HTTP Server
	{
		port := envvar.GetString("PORT", "41002")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		return 1
	}
}
*/

/*
package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "billing"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()
	wg := &sync.WaitGroup{}

	// Setup the service
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	env := backend.GetEnv()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create billing metrics
	billingServiceMetrics, err := metrics.NewBillingServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create billing service metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Get billing feature config
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BILLING2",
			Enum:        config.FEATURE_BILLING2,
			Value:       true,
			Description: "Receives BillingEntry2 types from Google Pub/Sub and writes them to BigQuery",
		},
	})
	featureConfig = envVarConfig
	featureBilling2 := featureConfig.FeatureEnabled(config.FEATURE_BILLING2)

	// Create no-op biller
	var biller2 billing.Biller = &billing.NoOpBiller{}

	if gcpOK {
		// Google BigQuery
		if featureBilling2 {
			billingDataset := envvar.GetString("GOOGLE_BIGQUERY_DATASET_BILLING", "")
			if billingDataset == "" {
				core.Error("GOOGLE_BIGQUERY_DATASET_BILLING not set")
				return 1
			}

			batchSize := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", billing.DefaultBigQueryBatchSize)

			summaryBatchSize := envvar.GetInt("GOOGLE_BIGQUERY_SUMMARY_BATCH_SIZE", int(billing.DefaultBigQueryBatchSize/10))

			batchSizePercent := envvar.GetFloat("FEATURE_BILLING2_BATCH_SIZE_PERCENT", 0.80)

			billing2TableName := envvar.GetString("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING", "billing2")

			billing2SummaryTableName := envvar.GetString("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING_SUMMARY", "billing2_session_summary")

			// Pass context without cancel to ensure writing continues even past reception of shutdown signal
			bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
			if err != nil {
				core.Error("could not create BigQuery client: %v", err)
				return 1
			}

			b := billing.GoogleBigQueryClient{
				Metrics:              &billingServiceMetrics.BillingMetrics,
				TableInserter:        bqClient.Dataset(billingDataset).Table(billing2TableName).Inserter(),
				SummaryTableInserter: bqClient.Dataset(billingDataset).Table(billing2SummaryTableName).Inserter(),
				BatchSize:            batchSize,
				SummaryBatchSize:     summaryBatchSize,
				BatchSizePercent:     batchSizePercent,
				FeatureBilling2:      featureBilling2,
			}

			// Set the Biller to BigQuery
			biller2 = &b

			// Start the background WriteLoop and SummaryWriteLoop to batch write to BigQuery
			wg.Add(1)
			go b.WriteLoop2(ctx, wg)

			wg.Add(1)
			go b.SummaryWriteLoop2(ctx, wg)
		}
	}

	errChan := make(chan error, 1)

	emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local biller
		biller2 = &billing.LocalBiller{
			Metrics: &billingServiceMetrics.BillingMetrics,
		}

		core.Debug("detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			numRecvGoroutines := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)

			if featureBilling2 {

				core.Debug("Billing2 enabled")

				entryVeto := envvar.GetBool("BILLING_ENTRY_VETO", false)

				maxRetries := envvar.GetInt("FEATURE_BILLING2_MAX_RETRIES", 25)

				retryTime := envvar.GetDuration("FEATURE_BILLING2_RETRY_TIME", time.Second*1)

				topicName := envvar.GetString("FEATURE_BILLING2_TOPIC_NAME", "billing2")
				subscriptionName := envvar.GetString("FEATURE_BILLING2_SUBSCRIPTION_NAME", "billing2")

				pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
				defer cancelFunc()

				pubsubForwarder, err := billing.NewPubSubForwarder(pubsubCtx, biller2, entryVeto, maxRetries, retryTime, &billingServiceMetrics.BillingMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
				if err != nil {
					core.Error("could not create pubsub forwarder: %v", err)
					return 1
				}

				// Start the pubsub forwarder
				wg.Add(1)
				go pubsubForwarder.Forward2(ctx, wg, errChan)
			}
		}
	}

	// Setup the status handler info
	statusData := &metrics.BillingStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				billingServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				billingServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.BillingStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(billingServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = billingServiceMetrics.MemoryAllocated.Value()
				newStatusData.Billing2EntriesReceived = int(billingServiceMetrics.BillingMetrics.Entries2Received.Value())
				newStatusData.Billing2EntriesSubmitted = int(billingServiceMetrics.BillingMetrics.Entries2Submitted.Value())
				newStatusData.Billing2EntriesQueued = int(billingServiceMetrics.BillingMetrics.Entries2Queued.Value())
				newStatusData.Billing2EntriesFlushed = int(billingServiceMetrics.BillingMetrics.Entries2Flushed.Value())
				newStatusData.Billing2SummaryEntriesSubmitted = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Submitted.Value())
				newStatusData.Billing2SummaryEntriesQueued = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Queued.Value())
				newStatusData.Billing2SummaryEntriesFlushed = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Flushed.Value())
				newStatusData.Billing2EntriesWithNaN = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2EntriesWithNaN.Value())
				newStatusData.Billing2InvalidEntries = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2InvalidEntries.Value())
				newStatusData.Billing2ReadFailures = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2ReadFailure.Value())
				newStatusData.Billing2WriteFailures = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2WriteFailure.Value())

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
		port := envvar.GetString("PORT", "41000")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		ctxCancelFunc()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}
*/

/*
// match data

package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	md "github.com/networknext/backend/modules/match_data"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "match_data"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()
	wg := &sync.WaitGroup{}

	// Setup the service
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	env := backend.GetEnv()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create match data service metrics
	matchDataServiceMetrics, err := metrics.NewMatchDataServiceMetrics(ctx, metricsHandler, serviceName, "match_data", "Match Data", "match data entry")
	if err != nil {
		core.Error("failed to create match data service metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Create no-op matcher
	var matcher md.Matcher = &md.NoOpMatcher{}

	if gcpOK {
		// Google BigQuery
		matchDataDataset := envvar.GetString("GOOGLE_BIGQUERY_DATASET_MATCH_DATA", "")
		if matchDataDataset == "" {
			core.Error("GOOGLE_BIGQUERY_DATASET_MATCH_DATA not set")
			return 1
		}

		batchSize := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", md.DefaultBigQueryBatchSize)

		batchSizePercent := envvar.GetFloat("GOOGLE_BIGQUERY_BATCH_SIZE_PERCENT", 0.80)
		if err != nil {
			core.Error("could not parse GOOGLE_BIGQUERY_BATCH_SIZE_PERCENT: %v", err)
			return 1
		}

		matchDataTableName := envvar.GetString("GOOGLE_BIGQUERY_TABLE_MATCH_DATA", "match_data")

		// Pass context without cancel to ensure writing continues even past reception of shutdown signal
		bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
		if err != nil {
			core.Error("could not create BigQuery client: %v", err)
			return 1
		}

		b := md.GoogleBigQueryClient{
			Metrics:          matchDataServiceMetrics.MatchDataMetrics,
			TableInserter:    bqClient.Dataset(matchDataDataset).Table(matchDataTableName).Inserter(),
			BatchSize:        batchSize,
			BatchSizePercent: batchSizePercent,
		}

		// Set the Matcher to BigQuery
		matcher = &b

		// Start the background WriteLoop to batch write to BigQuery
		wg.Add(1)
		go b.WriteLoop(ctx, wg)
	}

	errChan := make(chan error, 1)

	emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local matcher
		matcher = &md.LocalMatcher{
			Metrics: matchDataServiceMetrics.MatchDataMetrics,
		}

		core.Debug("detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			numRecvGoroutines := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)
			entryVeto := envvar.GetBool("MATCH_DATA_ENTRY_VETO", false)
			maxRetries := envvar.GetInt("MATCH_DATA_MAX_RETRIES", 25)
			retryTime := envvar.GetDuration("MATCH_DATA_RETRY_TIME", time.Second*1)
			topicName := envvar.GetString("MATCH_DATA_TOPIC_NAME", "match_data")
			subscriptionName := envvar.GetString("MATCH_DATA_SUBSCRIPTION_NAME", "match_data")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := md.NewPubSubForwarder(pubsubCtx, matcher, entryVeto, maxRetries, retryTime, matchDataServiceMetrics.MatchDataMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
			if err != nil {
				core.Error("could not create pubsub forwarder: %v", err)
				return 1
			}

			// Start the pubsub forwarder
			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg, errChan)
		}
	}

	// Setup the status handler info
	statusData := &metrics.MatchDataStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				matchDataServiceMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				matchDataServiceMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.MatchDataStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(matchDataServiceMetrics.ServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = matchDataServiceMetrics.ServiceMetrics.MemoryAllocated.Value()
				newStatusData.MatchDataEntriesReceived = int(matchDataServiceMetrics.MatchDataMetrics.EntriesReceived.Value())
				newStatusData.MatchDataEntriesSubmitted = int(matchDataServiceMetrics.MatchDataMetrics.EntriesSubmitted.Value())
				newStatusData.MatchDataEntriesQueued = int(matchDataServiceMetrics.MatchDataMetrics.EntriesQueued.Value())
				newStatusData.MatchDataEntriesFlushed = int(matchDataServiceMetrics.MatchDataMetrics.EntriesFlushed.Value())
				newStatusData.MatchDataEntriesWithNaN = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataEntriesWithNaN.Value())
				newStatusData.MatchDataInvalidEntries = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataInvalidEntries.Value())
				newStatusData.MatchDataReadFailures = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataReadFailure.Value())
				newStatusData.MatchDataWriteFailures = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataWriteFailure.Value())

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
		port := envvar.GetString("PORT", "41003")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		ctxCancelFunc()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}
*/
