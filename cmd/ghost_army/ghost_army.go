package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	ghostarmy "github.com/networknext/backend/modules/ghost_army"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"
)

const (
	SecondsInDay = 86400
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	ctx := context.Background()

	serviceName := "ghost_army"
	fmt.Printf("\n%s\n\n", serviceName)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	isDebug := envvar.GetBool("NEXT_DEBUG", false)
	if isDebug {
		core.Debug("running as debug")
	}

	gcpProjectID := backend.GetGCPProjectID()

	env := backend.GetEnv()

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

	ghostArmyMetrics, err := metrics.NewGhostArmyMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("could not create backend metrics: %v", err)
		return 1
	}

	// Setup the status handler info

	statusData := &metrics.GhostArmyStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				newStatusData := &metrics.GhostArmyStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(float64(runtime.NumGoroutine()))
				newStatusData.MemoryAllocated = memoryUsed()

				// Success Metrics
				newStatusData.EntriesPublished = int(ghostArmyMetrics.SuccessMetrics.SessionEntriesPublished.Value())

				// Error Metrics
				newStatusData.EntryPublishingFailure = int(ghostArmyMetrics.ErrorMetrics.SessionEntryPublishFailure.Value())
				newStatusData.SessionsOverEstimate = int(ghostArmyMetrics.ErrorMetrics.PublishedSessionsOverEstimate.Value())
				newStatusData.SessionEntryMarshalFailure = int(ghostArmyMetrics.ErrorMetrics.SessionEntryMarshalFailure.Value())

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
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
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
	}

	// this var is just used to catch the situation where ghost army publishes
	// way too many slices in a given interval. Shouldn't happen anymore but just in case

	if !envvar.Exists("GHOST_ARMY_PEAK_SESSION_COUNT") {
		core.Error("GHOST_ARMY_PEAK_SESSION_COUNT not set")
		return 1
	}

	estimatedPeakSessionCount := envvar.GetInt("GHOST_ARMY_PEAK_SESSION_COUNT", 25000)

	infile := envvar.Get("GHOST_ARMY_BIN", "")
	if infile == "" {
		core.Error("GHOST_ARMY_BIN not set")
		return 1
	}

	datacenterCSV := envvar.Get("DATACENTERS_CSV", "")
	if datacenterCSV == "" {
		core.Error("DATACENTERS_CSV not set")
		return 1
	}

	buyerID := ghostarmy.GhostArmyBuyerID(env)

	// parse datacenter csv
	inputfile, err := os.Open(datacenterCSV)
	if err != nil {
		core.Error("could not open '%s': %v", datacenterCSV, err)
		return 1
	}
	defer inputfile.Close()

	lines, err := csv.NewReader(inputfile).ReadAll()
	if err != nil {
		core.Error("could not read csv data: %v", err)
		return 1
	}

	var dcmap ghostarmy.DatacenterMap
	dcmap = make(map[uint64]ghostarmy.StrippedDatacenter)

	for lineNum, line := range lines {
		if lineNum == 0 {
			continue
		}

		var datacenter ghostarmy.StrippedDatacenter
		datacenter.Name = line[0]
		id, err := strconv.ParseUint(line[1], 10, 64)
		if err != nil {
			core.Error("could not parse id for dc %s", datacenter.Name)
			continue
		}
		datacenter.Lat, err = strconv.ParseFloat(line[2], 64)
		if err != nil {
			core.Error("could not parse lat for dc %s", datacenter.Name)
			continue
		}
		datacenter.Long, err = strconv.ParseFloat(line[3], 64)
		if err != nil {
			core.Error("could not parse long for dc %s", datacenter.Name)
			continue
		}

		dcmap[id] = datacenter
	}

	// read binary file
	bin, err := ioutil.ReadFile(infile)
	if err != nil {
		core.Error("could not read '%s': %v", infile, err)
		return 1
	}

	// unmarshal
	index := 0

	var numSegments uint64
	if !encoding.ReadUint64(bin, &index, &numSegments) {
		core.Error("could not read num segments")
	}

	slices := make([]transport.SessionPortalData, 0)
	entityIndex := 0
	for s := uint64(0); s < numSegments; s++ {
		var count uint64
		if !encoding.ReadUint64(bin, &index, &count) {
			core.Error("could not read count")
		}

		newSlice := make([]transport.SessionPortalData, count)
		slices = append(slices, newSlice...)
		fmt.Printf("reading in %d entries\n", count)

		for i := uint64(0); i < count; i++ {
			var entry ghostarmy.Entry
			if !entry.ReadFrom(bin, &index) {
				core.Error("can't read entry at index %d", i)
			}

			entry.Into(&slices[entityIndex], dcmap, buyerID)
			entityIndex++
		}
	}
	bin = nil
	// publish to zero mq, sleep for 10 seconds, repeat

	publishChan := make(chan transport.SessionPortalData)

	portalPublishers := make([]pubsub.Publisher, 0)
	{
		fmt.Println("setting up portal cruncher")

		portalCruncherHosts := envvar.GetList("PORTAL_CRUNCHER_HOSTS", []string{"tcp://127.0.0.1:5555", "tcp://127.0.0.1:5556"})

		if !envvar.Exists("POST_SESSION_PORTAL_SEND_BUFFER_SIZE") {
			core.Error("POST_SESSION_PORTAL_SEND_BUFFER_SIZE not set")
			return 1
		}

		postSessionPortalSendBufferSize := envvar.GetInt("POST_SESSION_PORTAL_SEND_BUFFER_SIZE", 1000000)

		for _, host := range portalCruncherHosts {
			portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(host, int(postSessionPortalSendBufferSize))
			if err != nil {
				core.Error("could not create portal cruncher publisher: %v", err)
				return 1
			}

			portalPublishers = append(portalPublishers, portalCruncherPublisher)
		}

		fmt.Println("portal cruncher setup complete")
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()

		publisherIndex := 0

		for {
			select {
			case slice := <-publishChan:
				// TODO: switch to serialization instead of binary encoding
				sessionBytes, err := slice.MarshalBinary()
				if err != nil {
					ghostArmyMetrics.ErrorMetrics.SessionEntryMarshalFailure.Add(1)
					core.Error("could not marshal binary for slice session id %d", slice.Meta.ID)
					continue
				}

				if _, err := portalPublishers[publisherIndex].Publish(ctx, pubsub.TopicPortalCruncherSessionData, sessionBytes); err != nil {
					ghostArmyMetrics.ErrorMetrics.SessionEntryPublishFailure.Add(1)
					core.Error("failed to publish ghost army session: %v", err)
				} else {
					ghostArmyMetrics.SuccessMetrics.SessionEntriesPublished.Add(1)
				}

				publisherIndex = (publisherIndex + 1) % len(portalPublishers)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		getLastMidnight := func() time.Time {
			t := time.Now()
			year, month, day := t.Date()
			return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
		}

		currentSecs := func() int64 {
			t := time.Now()
			year, month, day := t.Date()
			t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
			return int64(t.Sub(t2).Seconds())
		}

		// slice begin is the current number of seconds into the day
		sliceBegin := currentSecs()

		// date offset is used to adjust the slice timestamp to the current day
		dateOffset := getLastMidnight()

		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				begin := time.Now()

				// reset if at end of day
				if sliceBegin >= SecondsInDay {
					i = 0
					sliceBegin = currentSecs() // account for any inaccuracy by calling sleep()
					dateOffset = getLastMidnight()
				}

				var interval int64 = 10

				// only useful at 11:50:5x pm - midnight
				// forces the last batch to be sent within the above interval
				if sliceBegin+interval > SecondsInDay {
					interval = SecondsInDay - sliceBegin
				}

				// seek to the next position slices should be from
				// mainly useful when starting the program
				for i < len(slices) && slices[i].Slice.Timestamp.Unix() < sliceBegin {
					i++
				}

				before := i

				// only read for the interval, usually 10 seconds
				for i < len(slices) && slices[i].Slice.Timestamp.Unix() < sliceBegin+interval {
					slice := slices[i]

					// slice timestamp will be in the range of 0 - SecondsInDay * 3,
					// so adjust the timestamp by the time the loop was started
					slice.Slice.Timestamp = dateOffset.Add(time.Second * time.Duration(slice.Slice.Timestamp.Unix()))

					publishChan <- slice
					i++
				}

				diff := i - before

				if diff > estimatedPeakSessionCount {
					core.Debug("sent more than %d slices this interval, num sent = %d, current interval in secs = %d - %d", estimatedPeakSessionCount, diff, sliceBegin, sliceBegin+interval)
					ghostArmyMetrics.ErrorMetrics.PublishedSessionsOverEstimate.Add(1)
				}

				// increment by the interval
				sliceBegin += interval

				time.Sleep((time.Second * time.Duration(interval)) - time.Since(begin))
			}
		}
	}()

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	<-termChan // Exit with an error code of 0 if we receive SIGINT or SIGTERM

	fmt.Println("Received shutdown signal.")

	// Wait for essential goroutines to finish
	cancel()
	wg.Wait()

	fmt.Println("Successfully shutdown.")
	return 0
}
