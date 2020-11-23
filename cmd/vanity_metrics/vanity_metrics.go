package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	// "runtime"
	"strconv"
	"time"

	// gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {
	fmt.Printf("analytics: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	logger := log.NewLogfmtLogger(os.Stdout)

	// var metricsHandler metrics.Handler = &metrics.NoOpHandler{}

	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	if gcpOK {
		level.Info(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "Found GOOGLE_PROJECT_ID")
		// loggingClient, err := gcplogging.NewClient(ctx, projectID)
		// if err != nil {
		// 	level.Error(logger).Log("msg", "failed to create GCP logging client", "err", err)
		// 	os.Exit(1)
		// }

		// logger = logging.NewStackdriverLogger(loggingClient, "analytics")
	} else {
		level.Error(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "GOOGLE_PROJECT_ID not set. Cannot get vanity metrics.")
		os.Exit(1)
	}

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

	// env, ok := os.LookupEnv("ENV")
	// if !ok {
	// 	level.Error(logger).Log("err", "ENV not set")
	// 	os.Exit(1)
	// }

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	
	// StackDriver Metrics
	var enableSDMetrics bool
	var err error
	enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
	if ok {
		enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
		if err != nil {
			level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	} else {
		level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "ENABLE_STACKDRIVER_METRICS not set. Cannot get vanity metrics.")
		os.Exit(1)
	}

	var sd metrics.StackDriverHandler
	if enableSDMetrics {
		// Set up StackDriver metrics
		sd = metrics.StackDriverHandler{
			ProjectID:          gcpProjectID,
			OverwriteFrequency: time.Second,
			OverwriteTimeout:   10 * time.Second,
		}

		if err := sd.Open(ctx); err != nil {
			level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
			os.Exit(1)
		}



		// metricsHandler = &sd

		// sdwriteinterval := os.Getenv("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL")
		// writeInterval, err := time.ParseDuration(sdwriteinterval)
		// if err != nil {
		// 	level.Error(logger).Log("envvar", "GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", "value", sdwriteinterval, "err", err)
		// 	os.Exit(1)
		// }
		// go metricsHandler.WriteLoop(ctx, logger, writeInterval, 200)
	}


	// StackDriver Profiler

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
			Service:        "vanity_metrics",
			ServiceVersion: env,
			ProjectID:      gcpProjectID,
			MutexProfiling: true,
		}); err != nil {
			level.Error(logger).Log("msg", "failed to initialize StackDriver profiler", "err", err)
			os.Exit(1)
		}
	}

	filter := `metric.type = "custom.googleapis.com/server_backend/session_update.latency_worse"`
	// Name: "projects/" + sd.ProjectID,
	startTime := timestamppb.New(time.Now().Add(-10 * time.Minute))
	req := &monitoringpb.ListTimeSeriesRequest{
		Name: "projects/network-next-v3-dev/metricDescriptors/custom.googleapis.com/server_backend/session_update.latency_worse",
		Filter: filter,
		Interval: &monitoringpb.TimeInterval{EndTime: timestamppb.Now(), StartTime: startTime},
		View: monitoringpb.ListTimeSeriesRequest_TimeSeriesView(0),

	}
	it := sd.Client.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			fmt.Println("done with iterator")
			break
		}
		if err != nil {
			level.Error(logger).Log("err", err)
		}
		fmt.Printf("My resp is %v\n", resp)
		break
	}


	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))

			port, ok := os.LookupEnv("PORT")
			if !ok {
				level.Error(logger).Log("err", "env var PORT must be set")
				os.Exit(1)
			}

			level.Info(logger).Log("addr", ":"+port)

			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

func HealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		statusCode := http.StatusOK

		w.WriteHeader(statusCode)
		w.Write([]byte(http.StatusText(statusCode)))
	}
}
