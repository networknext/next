package common

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/core"
	db "github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/envvar"

	// todo: we want to move this to a new module ("middleware"?) or common as needed
	"github.com/networknext/backend/modules-old/transport/middleware"

	"github.com/gorilla/mux"
	"github.com/oschwald/maxminddb-golang"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

type RelayData struct {
	NumRelays          int
	RelayIds           []uint64
	RelayHash          map[uint64]db.Relay
	RelayArray         []db.Relay
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIds []uint64
	RelayIdToIndex     map[uint64]int
	DatacenterRelays   map[uint64][]int
	DestRelays         []bool
	DatabaseBinFile    []byte
}

type Service struct {
	Env           string
	ServiceName   string
	BuildTime     string
	CommitMessage string
	CommitHash    string
	Local         bool

	Router mux.Router

	Context           context.Context
	ContextCancelFunc context.CancelFunc

	GoogleProjectId string

	// ------------------

	databaseMutex     sync.RWMutex
	database          *db.Database
	databaseOverlay   *db.Overlay
	databaseRelayData *RelayData

	statusMutex sync.RWMutex
	statusData  *ServiceStatus

	magicMutex    sync.RWMutex
	magicData     []byte
	magicCounter  uint64
	upcomingMagic []byte
	currentMagic  []byte
	previousMagic []byte

	leaderElection *RedisLeaderElection

	sendTrafficToMe  func() bool
	machineIsHealthy func() bool

	udpServer *UDPServer

	routeMatrixMutex    sync.RWMutex
	routeMatrix         *RouteMatrix
	routeMatrixDatabase *db.Database

	googleCloudHandler *GoogleCloudHandler

	lookerHandler *LookerHandler

	ip2location_isp_mutex   sync.RWMutex
	ip2location_isp_reader  *maxminddb.Reader
	ip2location_city_mutex  sync.RWMutex
	ip2location_city_reader *maxminddb.Reader
}

func CreateService(serviceName string) *Service {

	service := Service{}
	service.ServiceName = serviceName
	service.CommitMessage = commitMessage
	service.CommitHash = commitHash
	service.BuildTime = buildTime

	core.Log("%s", service.ServiceName)

	env := envvar.GetString("ENV", "local")

	core.Log("env: %s", env)

	service.Local = env == "local"

	service.Env = env

	core.Log("commit message: %s", service.CommitMessage)
	core.Log("commit hash: %s", service.CommitHash)
	core.Log("build time: %s", service.BuildTime)

	service.Router.HandleFunc("/version", versionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
	service.Router.HandleFunc("/status", service.statusHandlerFunc())
	service.Router.HandleFunc("/lb_health", service.lbHealthHandlerFunc())
	service.Router.HandleFunc("/vm_health", service.vmHealthHandlerFunc())

	service.Context, service.ContextCancelFunc = context.WithCancel(context.Background())

	service.runStatusUpdateLoop()

	service.GoogleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "")
	if service.GoogleProjectId != "" {
		core.Log("google project id: %s", service.GoogleProjectId)
	}

	service.sendTrafficToMe = func() bool { return true }
	service.machineIsHealthy = func() bool { return true }

	return &service
}

func (service *Service) SetHealthFunctions(sendTrafficToMe func() bool, machineIsHealthy func() bool) {
	service.sendTrafficToMe = sendTrafficToMe
	service.machineIsHealthy = machineIsHealthy
}

func (service *Service) LoadDatabase() {

	databasePath := envvar.GetString("DATABASE_PATH", "database.bin")
	overlayPath := envvar.GetString("OVERLAY_PATH", "overlay.bin")

	service.database, service.databaseOverlay = loadDatabase(databasePath, overlayPath)

	if validateBinFiles(service.database) {
		core.Error("bin files failed validation")
		os.Exit(1)
	}

	applyOverlay(service.database, service.databaseOverlay)

	// TODO: should this be part of the database.bin validation?
	service.databaseRelayData = generateRelayData(service.database)
	if service.databaseRelayData == nil {
		core.Error("generate relay data failed")
		os.Exit(1)
	}

	core.Log("loaded database: %s", databasePath)

	service.watchDatabase(service.Context, databasePath, overlayPath)
}

func (service *Service) ValidateBinFiles(filenames []string) bool {

	database, _ := loadDatabase(filenames[0], filenames[1])

	return validateBinFiles(database)

}

func validateBinFiles(database *db.Database) bool {

	if database == nil {
		return false
	}

	return database.IsEmpty()

}

func (service *Service) LoadIP2Location() {

	filenames := envvar.GetList("IP2LOCATION_FILENAMES", []string{"GeoIP2-City.mmdb", "GeoIP2-ISP.mmdb"})

	cityReader, ispReader := loadIP2Location(filenames[0], filenames[1])

	if validateIP2Location(cityReader, ispReader) {
		core.Error("ip2location failed validation")
		os.Exit(1)
	}

	service.ip2location_city_mutex.Lock()
	service.ip2location_city_reader = cityReader
	service.ip2location_city_mutex.Unlock()

	core.Log("loaded ip2location city file: %s", filenames[0])

	service.ip2location_isp_mutex.Lock()
	service.ip2location_isp_reader = ispReader
	service.ip2location_isp_mutex.Unlock()

	core.Log("loaded ip2location isp file: %s", filenames[1])

	service.watchIP2Location(service.Context, filenames)
}

func (service *Service) ValidateIP2Location(filenames []string) bool {

	cityReader, ispReader := loadIP2Location(filenames[0], filenames[1])

	return validateIP2Location(cityReader, ispReader)

}

func validateIP2Location(cityReader *maxminddb.Reader, ispReader *maxminddb.Reader) bool {

	valid := true

	ip := net.ParseIP("192.0.2.1")

	if cityReader == nil {
		core.Error("city reader is nil")
		valid = false
	} else {
		lat, long := locateIP(cityReader, ip)
		if lat == 0.0 && long == 0.0 {
			core.Error("failed to validate city")
			valid = false
		}
	}

	if ispReader == nil {
		core.Error("isp reader is nil")
		valid = false
	} else {
		asn, isp := locateISP(ispReader, ip)
		if asn == -1 {
			core.Error("failed to validate asn")
			valid = false
		}

		if isp == "" {
			core.Error("failed to validate isp")
			valid = false
		}
	}

	return valid

}

func locateIP(reader *maxminddb.Reader, ip net.IP) (float32, float32) {
	var record struct {
		Location struct {
			Latitude  float64 `maxminddb:"latitude"`
			Longitude float64 `maxminddb:"longitude"`
		} `maxminddb:"location"`
	}
	err := reader.Lookup(ip, &record)
	if err != nil {
		return 0, 0
	}
	return float32(record.Location.Latitude), float32(record.Location.Longitude)
}

func locateISP(reader *maxminddb.Reader, ip net.IP) (int, string) {
	var record struct {
		ISP struct {
			AutonomousSystemNumber uint   `maxminddb:"autonomous_system_number"`
			ISP                    string `maxminddb:"isp"`
		}
	}
	err := reader.Lookup(ip, &record)
	if err != nil {
		return -1, ""
	}
	return int(record.ISP.AutonomousSystemNumber), record.ISP.ISP
}

func (service *Service) LocateIP(ip net.IP) (float32, float32) {
	service.ip2location_city_mutex.RLock()
	reader := service.ip2location_city_reader
	service.ip2location_city_mutex.RUnlock()
	return locateIP(reader, ip)
}

func (service *Service) LocateISP(ip net.IP) (int, string) {
	service.ip2location_isp_mutex.RLock()
	reader := service.ip2location_isp_reader
	service.ip2location_isp_mutex.RUnlock()
	return locateISP(reader, ip)
}

func (service *Service) Database() *db.Database {
	service.databaseMutex.RLock()
	database := service.database
	service.databaseMutex.RUnlock()
	return database
}

func (service *Service) DatabaseBinFile() []byte {
	service.databaseMutex.RLock()
	database := service.database
	service.databaseMutex.RUnlock()
	var databaseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&databaseBuffer)
	encoder.Encode(database)
	return databaseBuffer.Bytes()
}

func (service *Service) RelayData() *RelayData {
	service.databaseMutex.RLock()
	relayData := service.databaseRelayData
	service.databaseMutex.RUnlock()
	return relayData
}

func (service *Service) UpdateMagic() {
	service.updateMagicLoop()
}

func (service *Service) GetMagicValues() ([]byte, []byte, []byte) {
	service.magicMutex.Lock()
	upcomingMagic := service.upcomingMagic
	currentMagic := service.currentMagic
	previousMagic := service.previousMagic
	service.magicMutex.Unlock()
	return upcomingMagic, currentMagic, previousMagic
}

func (service *Service) lbHealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		if !service.sendTrafficToMe() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func (service *Service) vmHealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		if !service.machineIsHealthy() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func versionHandlerFunc(buildTime string, commitMessage string, commitHash string, allowedOrigins []string) func(w http.ResponseWriter, r *http.Request) {

	version := map[string]string{
		"build_time":     buildTime,
		"commit_message": commitMessage,
		"commit_hash":    commitHash,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		middleware.CORSControlHandlerFunc(allowedOrigins, w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (service *Service) StartWebServer() {
	port := envvar.GetString("HTTP_PORT", "80")
	core.Log("starting http server on port %s", port)
	go func() {
		err := http.ListenAndServe(":"+port, &service.Router)
		if err != nil {
			core.Error("error starting http server: %v", err)
			os.Exit(1)
		}
	}()
}

func (service *Service) StartUDPServer(packetHandler func(conn *net.UDPConn, from *net.UDPAddr, packet []byte)) {
	config := UDPServerConfig{}
	config.Port = envvar.GetInt("UDP_PORT", 40000)
	config.NumThreads = envvar.GetInt("UDP_NUM_THREADS", 16)
	config.SocketReadBuffer = envvar.GetInt("UDP_SOCKET_READ_BUFFER", 1024*1024)
	config.SocketWriteBuffer = envvar.GetInt("UDP_SOCKET_READ_BUFFER", 1024*1024)
	config.MaxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	core.Log("udp port: %d", config.Port)
	core.Log("udp num threads: %d", config.NumThreads)
	core.Log("udp socket read buffer: %d", config.SocketReadBuffer)
	core.Log("udp socket write buffer: %d", config.SocketWriteBuffer)
	core.Log("udp max packet size: %d", config.MaxPacketSize)
	core.Log("starting udp server on port %d", config.Port)
	service.udpServer = CreateUDPServer(service.Context, config, packetHandler)
}

func (service *Service) LeaderElection(autoRefresh bool) {

	core.Log("started leader election")

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.GetString("REDIS_PASSWORD", "")

	config := RedisLeaderElectionConfig{}
	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.Timeout = time.Second * 10
	config.ServiceName = service.ServiceName

	var err error
	service.leaderElection, err = CreateRedisLeaderElection(service.Context, config)
	if err != nil {
		core.Error("could not create redis leader election: %v")
		os.Exit(1)
	}

	if autoRefresh {
		service.leaderElection.Start(service.Context)
	}
}

func (service *Service) UpdateLeaderStore(dataStores []DataStoreConfig) {

	if service.leaderElection.autoRefresh {
		return
	}

	service.leaderElection.Store(service.Context, dataStores...)
}

func (service *Service) LoadLeaderStore() []DataStoreConfig {

	if service.leaderElection == nil || service.leaderElection.autoRefresh {
		return []DataStoreConfig{}
	}

	return service.leaderElection.Load(service.Context)
}

func (service *Service) UpdateRouteMatrix() {

	routeMatrixURI := envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	routeMatrixInterval := envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval.String())

	httpClient := &http.Client{
		Timeout: routeMatrixInterval,
	}

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				service.routeMatrixMutex.RLock()
				currentRouteMatrix := service.routeMatrix
				service.routeMatrixMutex.RUnlock()

				if currentRouteMatrix != nil && time.Now().Unix()-int64(currentRouteMatrix.CreatedAt) > 30 {
					core.Error("route matrix is stale")
					service.routeMatrixMutex.Lock()
					service.routeMatrix = nil
					service.routeMatrixMutex.Unlock()
				}

				response, err := httpClient.Get(routeMatrixURI)
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				response.Body.Close()

				newRouteMatrix := RouteMatrix{}

				err = newRouteMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				var newDatabase db.Database

				databaseBuffer := bytes.NewBuffer(newRouteMatrix.BinFileData)
				decoder := gob.NewDecoder(databaseBuffer)
				err = decoder.Decode(&newDatabase)
				if err != nil {
					core.Error("failed to read database: %v", err)
					continue
				}

				service.routeMatrixMutex.Lock()
				service.routeMatrix = &newRouteMatrix
				service.routeMatrixDatabase = &newDatabase
				service.routeMatrixMutex.Unlock()

				core.Debug("updated route matrix: %d relays", len(newRouteMatrix.RelayIds))
			}
		}
	}()
}

func (service *Service) RouteMatrixAndDatabase() (*RouteMatrix, *db.Database) {
	service.routeMatrixMutex.RLock()
	routeMatrix := service.routeMatrix
	database := service.routeMatrixDatabase
	service.routeMatrixMutex.RUnlock()
	return routeMatrix, database
}

func (service *Service) IsLeader() bool {
	if service.leaderElection != nil {
		return service.leaderElection.IsLeader()
	}
	return false
}

func (service *Service) WaitForShutdown() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	core.Log("received shutdown signal")

	service.ip2location_city_mutex.Lock()
	if service.ip2location_city_reader != nil {
		service.ip2location_city_reader.Close()
		service.ip2location_city_reader = nil
	}
	service.ip2location_city_mutex.Unlock()

	service.ip2location_isp_mutex.Lock()
	if service.ip2location_isp_reader != nil {
		service.ip2location_isp_reader.Close()
		service.ip2location_isp_reader = nil
	}
	service.ip2location_isp_mutex.Unlock()

	// todo: we need some system to wait for registered (named) subsystems to complete before we shut down

	core.Log("successfully shutdown")
}

// -----------------------------------------------------------------------

func loadIP2Location(cityPath string, ispPath string) (*maxminddb.Reader, *maxminddb.Reader) {
	cityReader, err := maxminddb.Open(cityPath)
	if err != nil {
		core.Error("failed to load ip2location city file: %v", err)
		return nil, nil
	}

	core.Debug("loaded ip2location city file: '%s'", cityPath)

	ispReader, err := maxminddb.Open(ispPath)
	if err != nil {
		core.Error("failed to load ip2location isp file: %v", err)
		return nil, nil
	}

	core.Debug("loaded ip2location city file: '%s'", cityPath)

	return cityReader, ispReader
}

func (service *Service) watchIP2Location(ctx context.Context, filenames []string) {

	syncInterval := envvar.GetDuration("IP2LOCATION_SYNC_INTERVAL", time.Minute)

	go func() {

		ticker := time.NewTicker(syncInterval)

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				cityReader, ispReader := loadIP2Location(filenames[0], filenames[1])
				if validateIP2Location(cityReader, ispReader) {
					core.Error("ip2location files not valid")
					continue
				}

				service.ip2location_city_mutex.Lock()
				oldCityReader := service.ip2location_city_reader
				service.ip2location_city_reader = oldCityReader
				service.ip2location_city_mutex.Unlock()

				oldCityReader.Close()

				service.ip2location_isp_mutex.Lock()
				oldISPReader := service.ip2location_isp_reader
				service.ip2location_isp_reader = oldISPReader
				service.ip2location_isp_mutex.Unlock()

				oldISPReader.Close()

				core.Debug("reloaded ip2location file")
			}
		}
	}()
}

// -----------------------------------------------------------------------

func loadDatabase(databasePath string, overlayPath string) (*db.Database, *db.Overlay) {

	// load database (required)

	database, err := db.LoadDatabase(databasePath)
	if err != nil {
		core.Error("error: could not read database: %v", err)
		return nil, nil
	}

	if database.IsEmpty() {
		core.Error("error: database is empty")
	}

	core.Debug("loaded database: '%s'", databasePath)

	// load overlay (optional)

	overlay, err := db.LoadOverlay(overlayPath)
	if err != nil {
		core.Debug("failed to load overlay: %v", err)
		return database, nil
	}

	if overlay.IsEmpty() {
		core.Debug("overlay is empty")
		return database, nil
	}

	// IMPORTANT: discard the overlay if it's older than the database
	if database.CreationTime > overlay.CreationTime {
		core.Debug("overlay is older than database")
		return database, nil
	}

	core.Debug("loaded overlay: '%s'", overlayPath)

	return database, overlay
}

func generateRelayData(database *db.Database) *RelayData {

	relayData := &RelayData{}

	numRelays := len(database.Relays)

	relayData.NumRelays = numRelays
	relayData.RelayIds = make([]uint64, numRelays)
	relayData.RelayHash = make(map[uint64]db.Relay)
	relayData.RelayArray = database.Relays

	sort.SliceStable(relayData.RelayArray, func(i, j int) bool {
		return relayData.RelayArray[i].Name < relayData.RelayArray[j].Name
	})

	for i := range relayData.RelayArray {
		relayData.RelayHash[relayData.RelayArray[i].ID] = relayData.RelayArray[i]
	}

	relayData.RelayAddresses = make([]net.UDPAddr, numRelays)
	relayData.RelayNames = make([]string, numRelays)
	relayData.RelayLatitudes = make([]float32, numRelays)
	relayData.RelayLongitudes = make([]float32, numRelays)
	relayData.RelayDatacenterIds = make([]uint64, numRelays)

	for i := 0; i < numRelays; i++ {
		relayData.RelayIds[i] = relayData.RelayArray[i].ID
		relayData.RelayAddresses[i] = relayData.RelayArray[i].Addr
		relayData.RelayNames[i] = relayData.RelayArray[i].Name
		relayData.RelayLatitudes[i] = float32(relayData.RelayArray[i].Datacenter.Latitude)
		relayData.RelayLongitudes[i] = float32(relayData.RelayArray[i].Datacenter.Longitude)
		relayData.RelayDatacenterIds[i] = relayData.RelayArray[i].Datacenter.ID
	}

	// build a mapping from relay id to relay index

	relayData.RelayIdToIndex = make(map[uint64]int)
	for i := 0; i < numRelays; i++ {
		relayData.RelayIdToIndex[relayData.RelayIds[i]] = i
	}

	// build a mapping from datacenter id to the set of relays in that datacenter

	relayData.DatacenterRelays = make(map[uint64][]int)

	for i := 0; i < numRelays; i++ {
		datacenterId := relayData.RelayDatacenterIds[i]
		relayData.DatacenterRelays[datacenterId] = append(relayData.DatacenterRelays[datacenterId], i)
	}

	// determine which relays are dest relays for at least one buyer

	relayData.DestRelays = make([]bool, numRelays)

	for _, buyer := range database.BuyerMap {
		if buyer.Live {
			for _, datacenter := range database.DatacenterMaps[buyer.ID] {
				datacenterRelays := relayData.DatacenterRelays[datacenter.DatacenterID]
				for j := 0; j < len(datacenterRelays); j++ {
					relayData.DestRelays[datacenterRelays[j]] = true
				}
			}
		}
	}

	// stash the database bin file in the relay data, so it's all guaranteed to be consistent

	var databaseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&databaseBuffer)
	encoder.Encode(database)

	relayData.DatabaseBinFile = databaseBuffer.Bytes()

	return relayData
}

func applyOverlay(database *db.Database, overlay *db.Overlay) {
	if overlay != nil {
		for _, buyer := range overlay.BuyerMap {
			_, ok := database.BuyerMap[buyer.ID]
			if !ok {
				database.BuyerMap[buyer.ID] = buyer
			}
		}
	}
}

func (service *Service) watchDatabase(ctx context.Context, databasePath string, overlayPath string) {

	syncInterval := envvar.GetDuration("DATABASE_SYNC_INTERVAL", time.Minute)

	go func() {

		ticker := time.NewTicker(syncInterval)

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				newDatabase, newOverlay := loadDatabase(databasePath, overlayPath)

				if validateBinFiles(newDatabase) {
					core.Error("new bin file failed validation")
					continue
				}

				// TODO: should this be part of database.bin validation?
				newRelayData := generateRelayData(newDatabase)

				if newRelayData == nil {
					continue
				}

				applyOverlay(newDatabase, newOverlay)

				service.databaseMutex.Lock()
				service.database = newDatabase
				service.databaseOverlay = newOverlay
				service.databaseRelayData = newRelayData
				service.databaseMutex.Unlock()
			}
		}
	}()
}

// -------------------------------------------------------------------------

type ServiceStatus struct {
	ServiceName     string  `json:"service_name"`
	CommitMessage   string  `json:"commit_message"`
	CommitHash      string  `json:"commit_hash"`
	BuildTime       string  `json:"build_time"`
	Started         string  `json:"started"`
	Uptime          string  `json:"uptime"`
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`
	IsLeader        bool    `json:"is_leader"`
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
	newStatusData.IsLeader = service.IsLeader()

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

func (service *Service) statusHandlerFunc() func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

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

// ----------------------------------------------------------

func getMagic(httpClient *http.Client, uri string) ([]byte, error) {

	response, err := httpClient.Get(uri)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to http get magic values: %v", err))
	}

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read magic data: %v", err))
	}

	response.Body.Close()

	if len(buffer) != 32 {
		return nil, errors.New(fmt.Sprintf("expected magic data to be 32 bytes, got %d", len(buffer)))
	}

	return buffer, nil
}

func (service *Service) updateMagicValues(magicData []byte) {

	if bytes.Equal(magicData, service.magicData) {
		return
	}

	// IMPORTANT: ignore any magic with older counters than we currently have
	// this avoids flapping when one instance of the magic backend is very slightly
	// delayed vs. another, and we get an older set of magic values
	magicCounter := binary.LittleEndian.Uint64(magicData[0:8])
	if magicCounter <= service.magicCounter {
		return
	}

	service.magicMutex.Lock()
	service.magicData = magicData
	service.magicCounter = magicCounter
	service.upcomingMagic = magicData[8:16]
	service.currentMagic = magicData[16:24]
	service.previousMagic = magicData[24:32]
	service.magicMutex.Unlock()

	core.Debug("updated magic values: %x -> %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
		service.magicCounter,
		service.upcomingMagic[0],
		service.upcomingMagic[1],
		service.upcomingMagic[2],
		service.upcomingMagic[3],
		service.upcomingMagic[4],
		service.upcomingMagic[5],
		service.upcomingMagic[6],
		service.upcomingMagic[7],
		service.currentMagic[0],
		service.currentMagic[1],
		service.currentMagic[2],
		service.currentMagic[3],
		service.currentMagic[4],
		service.currentMagic[5],
		service.currentMagic[6],
		service.currentMagic[7],
		service.previousMagic[0],
		service.previousMagic[1],
		service.previousMagic[2],
		service.previousMagic[3],
		service.previousMagic[4],
		service.previousMagic[5],
		service.previousMagic[6],
		service.previousMagic[7])
}

func (service *Service) updateMagicLoop() {

	magicURI := envvar.GetString("MAGIC_URI", "http://127.0.0.1:41007/magic")

	core.Log("magic uri: %s", magicURI)

	httpClient := &http.Client{
		Timeout: time.Second,
	}

	var magicData []byte
	for i := 0; i < 10; i++ {
		var err error
		magicData, err = getMagic(httpClient, magicURI)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if magicData == nil {
		core.Error("could not get initial magic values")
		os.Exit(1)
	}

	service.updateMagicValues(magicData)

	// start the goroutine to watch and update the magic every n seconds

	go func() {

		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:
				magicData, err := getMagic(httpClient, magicURI)
				if err == nil {
					service.updateMagicValues(magicData)
				}
			}
		}
	}()
}

// ----------------------------------------------------------

func isLeaderFunc(service *Service) func() bool {
	return func() bool {
		return service.IsLeader()
	}
}

func (service *Service) setupStorage() {

	googleCloudHandler, err := NewGoogleCloudHandler(service.Context, service.GoogleProjectId)
	if err != nil {
		core.Error("failed to create google cloud handler: %v", err)
		os.Exit(1)
	}

	service.googleCloudHandler = googleCloudHandler
}

func (service *Service) SyncFiles(config *FileSyncConfig) {
	config.Print()
	service.setupStorage()
	StartFileSync(service.Context, config, service.googleCloudHandler, isLeaderFunc(service))
}

// ---------------------------------------------------------------------------------------------------

func (service *Service) UseLooker() {

	config := LookerHandlerConfig{}

	config.HostURL = envvar.GetString("LOOKER_HOST_URL", "")
	config.ClientID = envvar.GetString("LOOKER_CLIENT_ID", "")
	config.Secret = envvar.GetString("LOOKER_CLIENT_SECRET", "")
	config.APISecret = envvar.GetString("LOOKER_API_SECRET", "")

	core.Log("looker host url: %s", config.HostURL)
	core.Log("looker client id: %s", config.ClientID)
	core.Log("looker client secret: %s", config.Secret)
	core.Log("looker api secret: %s", config.APISecret)

	lookerHandler, err := NewLookerHandler(config)
	if err != nil {
		core.Error("failed to create looker handler: %v", err)
		os.Exit(1)
	}

	service.lookerHandler = lookerHandler
}

func (service *Service) FetchWebsiteStats() error {
	newLookerStats, err := service.lookerHandler.RunWebsiteStatsQuery()
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", newLookerStats)

	return nil
}

// ----------------------------------------------------------
