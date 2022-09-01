package common

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"context"
	"sort"
	"time"
	"sync"
	"runtime"
	"encoding/json"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/routing"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

type Service struct {

	ServiceName string
	BuildTime string
	CommitMessage string
	CommitHash string
	
	Router mux.Router

	Context context.Context
	ContextCancelFunc context.CancelFunc

	// ------------------
	
	databaseMutex sync.RWMutex
	database *routing.DatabaseBinWrapper
	databaseOverlay *routing.OverlayBinWrapper
	databaseRelayHash map[uint64]routing.Relay
	databaseRelayArray []routing.Relay

	statusMutex sync.RWMutex
	statusData *ServiceStatus
}

func CreateService(serviceName string) *Service {

	service := Service{}
	service.ServiceName = serviceName
	service.CommitMessage = commitMessage
	service.CommitHash = commitHash
	service.BuildTime = buildTime

	fmt.Printf("%s\n", service.ServiceName)

	fmt.Printf("commit: %s [%s] (%s)\n", service.CommitMessage, service.CommitHash, service.BuildTime)

	env := backend.GetEnv()

	fmt.Printf("env: %s\n", env)

	service.Router.HandleFunc("/health", transport.HealthHandlerFunc())
	service.Router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
	service.Router.HandleFunc("/status", service.StatusHandlerFunc())

	service.Context, service.ContextCancelFunc = context.WithCancel(context.Background())

	service.runStatusUpdateLoop()

	return &service
}

func (service *Service) LoadDatabase() {

	databasePath := envvar.Get("DATABASE_PATH", "database.bin")
	overlayPath := envvar.Get("OVERLAY_PATH", "overlay.bin")

	service.database, service.databaseOverlay = loadDatabase(databasePath, overlayPath)

	applyOverlay(service.database, service.databaseOverlay)

	service.databaseRelayHash, service.databaseRelayArray = generateSecondaryValues(service.database)

	service.watchDatabase(service.Context, databasePath, overlayPath)
}

func (service *Service) Database() *routing.DatabaseBinWrapper {
	service.databaseMutex.RLock()
	database := service.database
	service.databaseMutex.RUnlock()
	return database
}

func (service *Service) RelayHash() map[uint64]routing.Relay {
	service.databaseMutex.RLock()
	relayHash := service.databaseRelayHash
	service.databaseMutex.RUnlock()
	return relayHash
}

func (service *Service) RelayArray() []routing.Relay {
	service.databaseMutex.RLock()
	relayArray := service.databaseRelayArray
	service.databaseMutex.RUnlock()
	return relayArray
}

func (service *Service) DatabaseAll() (*routing.DatabaseBinWrapper, map[uint64]routing.Relay, []routing.Relay) {
	service.databaseMutex.RLock()
	database := service.database
	relayHash := service.databaseRelayHash
	relayArray := service.databaseRelayArray
	service.databaseMutex.RUnlock()
	return database, relayHash, relayArray
}

func (service *Service) StartWebServer() {
	port := envvar.Get("HTTP_PORT", "80")
	fmt.Printf("starting http server on port %s\n", port)
	go func() {
		err := http.ListenAndServe(":"+port, &service.Router)
		if err != nil {
			core.Error("error starting http server: %v", err)
			os.Exit(1)
		}
	}()
}

func (service *Service) WaitForShutdown() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	core.Debug("received shutdown signal")
	// todo: probably need to wait for some stuff...
	core.Debug("successfully shutdown")
}

// -----------------------------------------------------------------------

func loadDatabase(databasePath string, overlayPath string) (*routing.DatabaseBinWrapper, *routing.OverlayBinWrapper) {

	// load the database (required)

	databaseFile, err := os.Open(databasePath)
	if err != nil {
		core.Error("could not load database: %v", err)
		return nil, nil
	}
	defer databaseFile.Close()

	database := routing.CreateEmptyDatabaseBinWrapper()
	err = backend.DecodeBinWrapper(databaseFile, database); 
	if err != nil || database.IsEmpty() {
		core.Error("error: could not read database: %v", err)
		return nil, nil
	}

	core.Debug("loaded database: '%s'", databasePath)

	// load the overlay if it exists

	overlayFile, err := os.Open(overlayPath)
	if err != nil {
		return database, nil
	}
	defer overlayFile.Close()
	
	overlay := routing.CreateEmptyOverlayBinWrapper()
	err = backend.DecodeOverlayWrapper(overlayFile, overlay)
	if err != nil || overlay.IsEmpty() {
		return database, nil
	}

	// IMPORTANT: discard the overlay if it's older than the database
	if database.CreationTime > overlay.CreationTime {
		return database, nil
	}

	core.Debug("loaded overlay: '%s'", overlayPath)

	return database, overlay
}

func generateSecondaryValues(database *routing.DatabaseBinWrapper) (map[uint64]routing.Relay, []routing.Relay) {

	relayHash := make(map[uint64]routing.Relay)
	relayArray := database.Relays

	sort.SliceStable(relayArray, func(i, j int) bool {
		return relayArray[i].Name < relayArray[j].Name
	})

	for i := range relayArray {
		relayHash[relayArray[i].ID] = relayArray[i]
	}

	return relayHash, relayArray
}

func applyOverlay(database *routing.DatabaseBinWrapper, overlay *routing.OverlayBinWrapper) {
	if overlay != nil {
		for _, buyer := range overlay.BuyerMap {
			_, ok := database.BuyerMap[buyer.ID]
			if !ok {
				database.BuyerMap[buyer.ID] = buyer
			}
		}
	}
}

func fileExists(filename string) bool {
 	_, err := os.Stat(filename)
    if err == nil {
        return true
    }
    return false
}

func (service *Service) watchDatabase(ctx context.Context, databasePath string, overlayPath string) {

	syncInterval := envvar.GetDuration("DATABASE_SYNC_INTERVAL", time.Minute)

	go func() {

		ticker := time.NewTicker(syncInterval)

		// todo: wg done defer

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				newDatabase, newOverlay := loadDatabase(databasePath, overlayPath)

				if newDatabase == nil {
					continue
				}

				relayHashNew, relayArrayNew := generateSecondaryValues(newDatabase)

				applyOverlay(newDatabase, newOverlay)

				service.databaseMutex.Lock()
				service.database = newDatabase
				service.databaseOverlay = newOverlay
				service.databaseRelayHash = relayHashNew
				service.databaseRelayArray = relayArrayNew
				service.databaseMutex.Unlock()
			}
		}
	}()
}

// -------------------------------------------------------------------------

type ServiceStatus struct {
	ServiceName 	string  `json:"service_name"`
	CommitMessage   string  `json:"commit_message"`
	CommitHash  	string  `json:"commit_hash"`
	BuildTime       string  `json:"build_time"`
	Started     	string  `json:"started"`
	Uptime      	string  `json:"uptime"`
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`
}

func (service *Service) updateStatus(startTime time.Time) {

	memoryAllocatedMB := func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Alloc) / (1000.0 * 1000.0)
	}
	
	newStatusData := &ServiceStatus{}

	newStatusData.ServiceName = service.ServiceName
	newStatusData.CommitMessage = commitMessage
	newStatusData.CommitHash = commitHash
	newStatusData.BuildTime = buildTime
	newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
	newStatusData.Uptime = time.Since(startTime).String()
	newStatusData.Goroutines = int(runtime.NumGoroutine())
	newStatusData.MemoryAllocated = memoryAllocatedMB()

	service.statusMutex.Lock()
	service.statusData = newStatusData
	service.statusMutex.Unlock()
}

func (service *Service) runStatusUpdateLoop() {
	startTime := time.Now()
	service.updateStatus(startTime)
	go func() {
		for {
			service.updateStatus(startTime)
			time.Sleep(time.Second * 10)
		}
	}()
}

func (service *Service) StatusHandlerFunc() func (w http.ResponseWriter, r *http.Request) {

	return func (w http.ResponseWriter, r *http.Request) {
		
		service.statusMutex.RLock()
		data := service.statusData
		service.statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(*data); err != nil {
			core.Error("could not write status data to json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
