/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

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

    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/envvar"

    "github.com/networknext/backend/modules-old/backend"
    "github.com/networknext/backend/modules-old/config"
    "github.com/networknext/backend/modules-old/metrics"
    portalcruncher "github.com/networknext/backend/modules-old/portal_cruncher"
    "github.com/networknext/backend/modules-old/transport"
    "github.com/networknext/backend/modules-old/transport/pubsub"

    "github.com/gorilla/mux"
)

var (
    buildTime     string
    commitMessage string
    commitHash    string
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
    os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
    serviceName := "portal_cruncher"
    fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

    est, _ := time.LoadLocation("EST")
    startTime := time.Now().In(est)

    // Setup the service
    ctx, cancel := context.WithCancel(context.Background())

    gcpProjectID := backend.GetGCPProjectID()

    backend.GetLogger(ctx, gcpProjectID, serviceName)

    logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
    if err != nil {
        core.Error("failed to get logger: %v", err)
        return 1
    }

    env := backend.GetEnv()

    metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
    if err != nil {
        core.Error("failed to get metrics handler: %v", err)
        return 1
    }

    if gcpProjectID != "" {
        if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
            core.Error("failed to initialize stackdriver profiler: %v", err)
            return 1
        }
    }

    portalCruncherMetrics, err := metrics.NewPortalCruncherMetrics(ctx, metricsHandler)
    if err != nil {
        core.Error("failed to create portal cruncher metrics: %v", err)
        return 1
    }

    btMetrics, err := metrics.NewBigTableMetrics(ctx, metricsHandler)
    if err != nil {
        core.Error("failed to create bigtable metrics: %v", err)
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

    // Start portal cruncher subscriber
    var portalSubscriber pubsub.Subscriber
    {
        /*
           cruncherPort := envvar.GetString("CRUNCHER_PORT", "5555")
           if cruncherPort == "" {
               core.Error("CRUNCHER_PORT not set")
               return 1
           }

           receiveBufferSize := envvar.GetInt("CRUNCHER_RECEIVE_BUFFER_SIZE", 1000000)

           portalCruncherSubscriber, err := pubsub.NewPortalCruncherSubscriber(cruncherPort, int(receiveBufferSize))
           if err != nil {
               core.Error("failed to create portal cruncher subscriber: %v", err)
               return 1
           }

           if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData); err != nil {
               core.Error("failed to subscribe to portal cruncher session data topic: %v", err)
               return 1
           }

           if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts); err != nil {
               core.Error("failed to subscribe to portal cruncher session counts topic: %v", err)
               return 1
           }

           portalSubscriber = portalCruncherSubscriber
        */
    }

    redisPingFrequency := envvar.GetDuration("CRUNCHER_REDIS_PING_FREQUENCY", time.Second*30)

    redisFlushFrequency := envvar.GetDuration("CRUNCHER_REDIS_FLUSH_FREQUENCY", time.Second)

    redisFlushCount := envvar.GetInt("PORTAL_CRUNCHER_REDIS_FLUSH_COUNT", 1000)

    redisGoroutineCount := envvar.GetInt("CRUNCHER_REDIS_GOROUTINE_COUNT", 5)

    messageChanSize := envvar.GetInt("CRUNCHER_MESSAGE_CHANNEL_SIZE", 10000000)

    redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
    redisPassword := envvar.GetString("REDIS_PASSWORD", "")
    redisMaxIdleConns := envvar.GetInt("REDIS_MAX_IDLE_CONNS", 10)
    redisMaxActiveConns := envvar.GetInt("REDIS_MAX_ACTIVE_CONNS", 64)

    // Determine if should insert into Bigtable
    useBigtable := featureConfig.FeatureEnabled(config.FEATURE_BIGTABLE)

    // Get Bigtable instance ID
    btInstanceID := envvar.GetString("BIGTABLE_INSTANCE_ID", "localhost:8086")
    // Get the table name
    btTableName := envvar.GetString("BIGTABLE_TABLE_NAME", "")
    // Get the column family name
    btCfName := envvar.GetString("BIGTABLE_CF_NAME", "")

    btGoroutineCount := envvar.GetInt("BIGTABLE_CRUNCHER_GOROUTINE_COUNT", 1)

    btEmulatorOK := envvar.Exists("BIGTABLE_EMULATOR_HOST")
    btHistoricalPath := envvar.GetString("BIGTABLE_HISTORICAL_TXT", "./testdata/bigtable_historical.txt")

    portalCruncher, err := portalcruncher.NewPortalCruncher(ctx,
        portalSubscriber,
        redisHostname,
        redisPassword,
        redisMaxIdleConns,
        redisMaxActiveConns,
        useBigtable,
        gcpProjectID,
        btInstanceID,
        btTableName,
        btCfName,
        btEmulatorOK,
        btHistoricalPath,
        messageChanSize,
        portalCruncherMetrics,
        btMetrics)
    if err != nil {
        core.Error("failed to create portal cruncher: %v", err)
        return 1
    }

    if err := portalCruncher.PingRedis(); err != nil {
        core.Error("unable to ping redis: %v", err)
        return 1
    }

    // Create an error channel for goroutines
    errChan := make(chan error, 1)

    // Create a waitgroup to manage clean shutdown
    var wg sync.WaitGroup

    // Start the portal cruncher
    go portalCruncher.Start(ctx, redisGoroutineCount, btGoroutineCount, redisPingFrequency, redisFlushFrequency, redisFlushCount, env, errChan, &wg)

    // Setup the status handler info

    statusData := &metrics.PortalCruncherStatus{}
    var statusMutex sync.RWMutex

    {
        memoryUsed := func() float64 {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            return float64(m.Alloc) / (1000.0 * 1000.0)
        }

        go func() {
            for {
                portalCruncherMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
                portalCruncherMetrics.MemoryAllocated.Set(memoryUsed())

                newStatusData := &metrics.PortalCruncherStatus{}

                // Service Information
                newStatusData.ServiceName = serviceName
                newStatusData.GitHash = commitMessage
                newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
                newStatusData.Uptime = time.Since(startTime).String()

                // Service Metrics
                newStatusData.Goroutines = int(portalCruncherMetrics.Goroutines.Value())
                newStatusData.MemoryAllocated = portalCruncherMetrics.MemoryAllocated.Value()

                // Cruncher Counts
                newStatusData.ReceivedMessageCount = int(portalCruncherMetrics.ReceivedMessageCount.Value())

                // Bigtable Counts
                newStatusData.WriteMetaSuccessCount = int(btMetrics.WriteMetaSuccessCount.Value())
                newStatusData.WriteSliceSuccessCount = int(btMetrics.WriteSliceSuccessCount.Value())

                // Bigtable Errors
                newStatusData.WriteMetaFailureCount = int(btMetrics.WriteMetaFailureCount.Value())
                newStatusData.WriteSliceFailureCount = int(btMetrics.WriteSliceFailureCount.Value())

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
        port := envvar.GetString("HTTP_PORT", "30000")
        if port == "" {
            core.Error("HTTP_PORT not set")
            return 1
        }
        fmt.Printf("starting http server on :%s\n", port)

        router := mux.NewRouter()
        router.HandleFunc("/health", transport.HealthHandlerFunc())
        router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
        router.HandleFunc("/status", serveStatusFunc).Methods("GET")
        router.Handle("/debug/vars", expvar.Handler())

        enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
        if err != nil {
            core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
        }
        if enablePProf {
            router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
        }

        go func() {
            err := http.ListenAndServe(":"+port, router)
            if err != nil {
                core.Error("failed to start http server: %v", err)
                errChan <- err
            }
        }()
    }

    // Wait for shutdown signal
    termChan := make(chan os.Signal, 1)
    signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

    select {
    case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
        fmt.Println("Received shutdown signal.")

        cancel()
        // Wait for essential goroutines to finish up
        wg.Wait()

        // Close the redis pool connection
        portalCruncher.CloseRedisPool()
        // Close Bigtable client
        portalCruncher.CloseBigTable()

        fmt.Println("Successfully shutdown.")
        return 0
    case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
        // Still let essential goroutines finish even though we got an error
        cancel()
        // Wait for essential goroutines to finish up
        wg.Wait()

        // Close the redis pool connection
        portalCruncher.CloseRedisPool()
        // Close Bigtable client
        portalCruncher.CloseBigTable()

        return 1
    }
}
