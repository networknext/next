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

	service.Context, service.ContextCancelFunc = context.WithCancel(context.Background())

	return &service
}

func (service *Service) LoadDatabase() {

	databasePath := envvar.Get("DATABASE_PATH", "database.bin")
	overlayPath := envvar.Get("OVERLAY_PATH", "overlay.bin")

	database, overlay := loadDatabase(databasePath, overlayPath)

	// todo: store database etc.

	_ = database
	_ = overlay

	watchDatabase(service.Context, databasePath, overlayPath)
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

	fmt.Printf("loaded database: '%s'\n", databasePath)

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

	fmt.Printf("loaded overlay: '%s'\n", overlayPath)

	return database, overlay
}

func generateSecondaryValues(database *routing.DatabaseBinWrapper) (map[uint64]routing.Relay, []routing.Relay) {

	relayHash := make(map[uint64]routing.Relay)
	relayArray := database.Relays

	sort.SliceStable(relayArray, func(i, j int) bool {
		return relayArray[i].Name < relayArray[j].Name
	})

	for i := range relayArray {
		fmt.Printf("loaded relay %s\n", relayArray[i].Name)
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

func watchDatabase(ctx context.Context, databasePath string, overlayPath string) {

	syncInterval := envvar.GetDuration("DATABASE_SYNC_INTERVAL", time.Second)//time.Minute)

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

				_ = relayHashNew
				_ = relayArrayNew

				// todo: pointer swaps under mutex (don't need to hold on to the overlay)
			}
		}
	}()
}
