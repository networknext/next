package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
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
	env           string

	buyerCodeMap   map[string]routing.Buyer
	buyerCodeMutex sync.RWMutex

	binCreator      string
	binCreationTime string
)

func init() {
	database := routing.CreateEmptyDatabaseBinWrapper()

	buyerCodeMap = make(map[string]routing.Buyer)

	filePath := envvar.Get("BIN_PATH", "./database.bin")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("could not load database binary: %s\n", filePath)
		os.Exit(1)
	}
	defer file.Close()

	if err = backend.DecodeBinWrapper(file, database); err != nil {
		fmt.Printf("DecodeBinWrapper() error: %v\n", err)
		os.Exit(1)
	}

	// Store a map of company code to buyer
	for _, buyer := range database.BuyerMap {
		buyerCodeMap[buyer.CompanyCode] = buyer
	}

	// Store the creator and creation time from the database
	binCreator = database.Creator
	binCreationTime = database.CreationTime
}

func main() {
	os.Exit(mainReturnWithCode())
}

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "api"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()

	// Get env and set the variable
	{
		envName, err := backend.GetEnv()
		if err != nil {
			core.Error("ENV not set: %v", err)
			return 1
		}
		env = envName
	}

	gcpProjectID = backend.GetGCPProjectID()
	if gcpProjectID != "" {
		core.Debug("Found GOOGLE_PROJECT_ID: %s", gcpProjectID)
	} else {
		core.Error("GOOGLE_PROJECT_ID not set. Cannot get vanity metrics")
		return 1
	}

	// StackDriver Logging
	// TODO: remove this once vanity metrics no longer requires a gokit logger
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("could not get logger: %v", err)
		return 1
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then

	// StackDriver Metrics
	var enableSDMetrics bool
	enableSDMetricsString := envvar.Get("ENABLE_STACKDRIVER_METRICS", "")
	if enableSDMetricsString != "" {
		enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
		if err != nil {
			core.Error("could not parse ENABLE_STACKDRIVER_METRICS: %v", err)
			return 1
		}
	} else {
		core.Error("ENABLE_STACKDRIVER_METRICS not set. Cannot get vanity metrics")
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
			core.Error("failed to create StackDriver metrics client: %v", err)
			return 1
		}

		sd = &sdHandler
	}

	redisHostname := envvar.Get("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.Get("REDIS_PASSWORD", "")

	// Create metric handler for tracking performance of api service
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create vanity metric metrics: %v", err)
		return 1
	}

	// StackDriver Profiler
	if err = backend.InitStackDriverProfiler(gcpProjectID, "api", env); err != nil {
		core.Error("failed to initialize StackDriver profiler: %v", err)
		return 1
	}

	vanityMetrics, err = vanity.NewVanityMetricHandler(sd, vanityServiceMetrics, 0, nil, redisHostname, redisPassword, 5, 5, time.Minute*5, "", env, logger)
	if err != nil {
		core.Error("failed to create vanity metric handler: %v", err)
		return 1
	}

	errChan := make(chan error, 1)

	// Setup file watchman on database.bin
	{
		// Get absolute path of database.bin
		databaseFilePath := envvar.Get("BIN_PATH", "./database.bin")
		absPath, err := filepath.Abs(databaseFilePath)
		if err != nil {
			core.Error("error getting absolute path %s: %v", databaseFilePath, err)
			return 1
		}

		// Check if file exists
		if _, err := os.Stat(absPath); err != nil {
			core.Error("%s does not exist: %v", absPath, err)
			return 1
		}

		// Get the directory of the database.bin
		// Used to watch over file creation and modification
		directoryPath := filepath.Dir(absPath)

		binSyncInterval, err := envvar.GetDuration("BIN_SYNC_INTERVAL", time.Minute*1)
		if err != nil {
			core.Error("failed to get BIN_SYNC_INTERVAL: %v", err)
			return 1
		}

		ticker := time.NewTicker(binSyncInterval)

		// Setup goroutine to watch for replaced file and update buyerCodeMap
		go func() {
			core.Debug("started watchman on %s", directoryPath)
			for {
				select {
				case <-ticker.C:
					// File has changed
					file, err := os.Open(absPath)
					if err != nil {
						core.Error("could not load database binary at %s: %v", absPath, err)
						continue
					}

					// Setup new buyer code map to read into
					databaseNew := routing.CreateEmptyDatabaseBinWrapper()

					if err = backend.DecodeBinWrapper(file, databaseNew); err == io.EOF {
						// Sometimes we receive an EOF error since the file is still being replaced
						// so early out here and proceed on the next notification
						file.Close()
						core.Debug("DecodeBinWrapper() EOF error, will wait for next notification")
						continue
					} else if err != nil {
						file.Close()
						core.Debug("DecodeBinWrapper() error: %v", err)
						continue
					}

					// Close the file since it is no longer needed
					file.Close()

					if databaseNew.IsEmpty() {
						// Don't want to use an empty bin wrapper
						// so early out here and use existing values
						core.Error("new database file is empty, keeping previous values")
						continue
					}

					// Store the creator and creation time from the database
					binCreator = databaseNew.Creator
					binCreationTime = databaseNew.CreationTime

					// Store the latest buyer info from the database
					buyerCodeMapNew := make(map[string]routing.Buyer)

					for _, buyer := range databaseNew.BuyerMap {
						buyerCodeMapNew[buyer.CompanyCode] = buyer
					}

					// Pointer swap the buyer code map
					buyerCodeMutex.Lock()
					buyerCodeMap = buyerCodeMapNew
					buyerCodeMutex.Unlock()
				}
			}
		}()
	}

	// Setup the status handler info
	type APIStatus struct {
		// Service Information
		ServiceName string `json:"serviceName"`
		GitHash     string `json:"gitHash"`
		Started     string `json:"started"`
		Uptime      string `json:"uptime"`

		// Metrics
		Goroutines      int     `json:"goroutines"`
		MemoryAllocated float64 `json:"mb_allocated"`
	}

	statusData := &APIStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				newStatusData := &APIStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()
				newStatusData.Goroutines = runtime.NumGoroutine()
				newStatusData.MemoryAllocated = memoryUsed()

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
		port := envvar.Get("PORT", "30005")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/vanity", VanityMetricHandlerFunc())
			router.HandleFunc("/health", transport.HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
			router.HandleFunc("/database_version", transport.DatabaseBinVersionFunc(&binCreator, &binCreationTime, &env))
			router.HandleFunc("/status", serveStatusFunc).Methods("GET")

			enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
			if err != nil {
				core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
			}
			if enablePProf {
				router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
			}

			err = http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
				return
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

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}

func VanityMetricHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var buyerID string
		ctx := context.Background()
		rawCompanyCode, ok := r.URL.Query()["id"]
		if !ok {
			// id was not provided, assume buyerID is global
			buyerID = fmt.Sprintf("global_%s", env)
		} else {
			companyCode := rawCompanyCode[0]

			// Vanity metrics for specific buyer
			buyerCodeMutex.RLock()
			buyer, ok := buyerCodeMap[companyCode]
			buyerCodeMutex.RUnlock()
			if !ok {
				errStr := fmt.Sprintf("id %s is not valid", companyCode)
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}

			// Check if vanity metrics enabled for this buyer
			if !buyer.InternalConfig.EnableVanityMetrics {
				// Vanity metrics are not enabled for this buyer
				errStr := fmt.Sprintf("vanity metrics are not enabled for buyer %s", companyCode)
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}
			buyerID = fmt.Sprintf("%016x", buyer.ID)
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

		data, err := vanityMetrics.GetVanityMetricJSON(ctx, sd, gcpProjectID, buyerID, startTime, endTime)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
		return
	}
}
