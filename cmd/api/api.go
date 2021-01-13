package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/vanity"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
	gcpProjectID  string
	vanityMetrics *vanity.VanityMetricHandler
	sd            *metrics.StackDriverHandler
	shortNameMap  map[string]string
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	fmt.Printf("api: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	logger := log.NewLogfmtLogger(os.Stdout)

	env, err := backend.GetEnv()
	if err != nil {
		level.Error(logger).Log("err", "ENV not set")
		return 1
	}

	// Set the map for buyer short name to buyerID
	shortNameMap = GetShortNameBuyerIDMap(env)

	gcpProjectID = backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""
	if gcpOK {
		level.Info(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "Found GOOGLE_PROJECT_ID")
	} else {
		level.Error(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "GOOGLE_PROJECT_ID not set. Cannot get vanity metrics.")
		return 1
	}

	// StackDriver Logging
	logger, err = backend.GetLogger(ctx, gcpProjectID, "api")
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then

	// StackDriver Metrics
	var enableSDMetrics bool
	enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
	if ok {
		enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
		if err != nil {
			level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "could not parse", "err", err)
			return 1
		}
	} else {
		level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "ENABLE_STACKDRIVER_METRICS not set. Cannot get vanity metrics.")
		return 1
	}

	if enableSDMetrics {
		// Set up StackDriver metrics
		sdHandler := metrics.StackDriverHandler{
			ProjectID:          gcpProjectID,
			OverwriteFrequency: time.Second,
			OverwriteTimeout:   10 * time.Second,
		}

		if err := sdHandler.Open(ctx); err != nil {
			level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
			return 1
		}

		sd = &sdHandler
	}

	// Create metric handler for tracking performance of api service
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create vanity metric metrics", "err", err)
		return 1
	}

	// StackDriver Profiler
	if err = backend.InitStackDriverProfiler(gcpProjectID, "api", env); err != nil {
		level.Error(logger).Log("msg", "failed to initialize StackDriver profiler", "err", err)
		return 1
	}

	vanityMetrics, err = vanity.NewVanityMetricHandler(sd, vanityServiceMetrics, 0, nil, "", 5, 5, time.Minute*5, "", env, logger)
	if err != nil {
		level.Error(logger).Log("msg", "error creating vanity metric handler", "err", err)
		return 1
	}

	errChan := make(chan error, 1)
	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/vanity", VanityMetricHandlerFunc())
			router.HandleFunc("/health", HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))

			port, ok := os.LookupEnv("PORT")
			if !ok {
				level.Error(logger).Log("err", "env var PORT must be set")
				errChan <- err
				return
			}

			serviceName := ""
			level.Info(logger).Log("addr", serviceName+":"+port)

			err := http.ListenAndServe(serviceName+":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}

		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}

func VanityMetricHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rawShortName, ok := r.URL.Query()["id"]
		if !ok {
			http.Error(w, "id is missing", http.StatusBadRequest)
			return
		}
		shortName := rawShortName[0]
		buyerID, exists := shortNameMap[shortName]
		if !exists {
			http.Error(w, "id is not valid", http.StatusBadRequest)
			return
		}

		// Get start time
		rawStartTime, ok := r.URL.Query()["start"]
		if !ok {
			http.Error(w, "start is missing", http.StatusBadRequest)
			return
		}
		// Parse the start time
		startTime, err := time.Parse(time.RFC3339, rawStartTime[0])
		if err != nil {
			errStr := fmt.Sprintf("could not parse start=%s as RFC3339 format (i.e. 2020-12-16T15:04:05-07:00 or 2020-12-16T23:04:05Z): %v", rawStartTime[0], err)
			http.Error(w, errStr, http.StatusBadRequest)
			return
		}

		// Get end time
		var endTime time.Time
		rawEndTime, ok := r.URL.Query()["end"]
		if !ok {
			// No end time provided, use time.Now()
			endTime = time.Now()
		} else {
			// Parse the end time
			endTime, err = time.Parse(time.RFC3339, rawEndTime[0])
			if err != nil {
				errStr := fmt.Sprintf("could not parse end=%s as RFC3339 format (i.e. 2020-12-16T15:04:05-07:00 or 2020-12-16T23:04:05Z): %v", rawEndTime[0], err)
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}
		}

		// Check if end time is before start time
		if endTime.Before(startTime) {
			errStr := fmt.Sprintf("end time %v is before start time %v", endTime, startTime)
			http.Error(w, errStr, http.StatusBadRequest)
			return
		}

		data, err := vanityMetrics.GetVanityMetricJSON(context.Background(), sd, gcpProjectID, buyerID, startTime, endTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}
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

func GetShortNameBuyerIDMap(env string) map[string]string {
	switch env {
	case "prod":
		return map[string]string{
			"gryadki":   "f43f9358918c145f", // GryadkiTeam
			"liquid":    "f0d3cc73dca0bc89", // Liquid Bit
			"velan":     "e5cee310ae3e26bc", // Velan Studios
			"orionark":  "e2cd4671821abeb2", // Orionark
			"Next":      "cd8d3fe954852686", // Network Next
			"blue":      "c46cf0f66b4f1ac7", // Blue Mammoth Games
			"turtle":    "c0ed3f02466425fb", // Turtle Rock Studios
			"psyonix":   "b8e4f84ca63b2021", // Psyonix
			"project":   "a955f6f111256ab4", // Project Games
			"gregyp":    "9ccf980c401aa166", // gregyp
			"hangzhou":  "838d793ef34d9bd3", // Hangzhou 24 Entertainment Network Technology Co. Ltd.
			"dylon":     "2f51620acca790b1", // Dylon
			"raspberry": "2b9c891211588152", // Raspberry
			"esl":       "1e4e8804454033c8", // ESL Gaming Online, Inc.
			"ghost":     "0000000000000000", // Ghost Army
			"global":    "global_prod",      // Global prod metrics
		}
	case "staging":
		return map[string]string{
			"next":   "bdbebdbf0f7be395", // Network Next
			"global": "global_staging",   // Global staging metrics
		}
	case "dev":
		return map[string]string{
			"Next":      "d25bfa9554e11583", // Network Next
			"next":      "bdbebdbf0f7be395", // networknext
			"wolfjaw":   "395a327867345157", // Wolfjaw Studios
			"raspberry": "2b9c891211588152", // Raspberry
			"valve":     "22edfbe14c08f419", // Valve
			"ghost":     "0000000000000000", // Ghost Army
			"global":    "global_dev",       // Global dev metrics
		}
	default:
		return map[string]string{
			"local":     "bdbebdbf0f7be395", // Local testing
			"raspberry": "2b9c891211588152", // Raspberry
			"global":    "global_local",     // Global local metrics
		}
	}
}
