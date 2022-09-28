package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"

	"github.com/networknext/backend/modules-old/backend"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules-old/transport"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

type RelayData struct {
	NumRelays          int
	RelayIds           []uint64
	RelayHash          map[uint64]routing.Relay
	RelayArray         []routing.Relay
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIds []uint64
	RelayIdToIndex     map[uint64]int
	DatacenterRelays   map[uint64][]int
	DestRelays         []bool
	DestRelayNames     []string
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

	// ------------------

	databaseMutex     sync.RWMutex
	database          *routing.DatabaseBinWrapper
	databaseOverlay   *routing.OverlayBinWrapper
	databaseRelayData *RelayData

	statusMutex sync.RWMutex
	statusData  *ServiceStatus

	magicMutex    sync.RWMutex
	magicData     []byte
	upcomingMagic []byte
	currentMagic  []byte
	previousMagic []byte

	leaderElection *RedisLeaderElection

	healthHandler func(w http.ResponseWriter, r *http.Request)

	udpServer *UDPServer

	routeMatrixMutex    sync.RWMutex
	routeMatrix         *RouteMatrix
	routeMatrixDatabase *routing.DatabaseBinWrapper

	gcpStorage *GCPStorage
}

func CreateService(serviceName string) *Service {

	service := Service{}
	service.ServiceName = serviceName
	service.CommitMessage = commitMessage
	service.CommitHash = commitHash
	service.BuildTime = buildTime

	core.Log("%s", service.ServiceName)

	core.Log("commit message: %s", service.CommitMessage)
	core.Log("commit hash: %s", service.CommitHash)
	core.Log("build time: %s", service.BuildTime)

	env := backend.GetEnv()

	core.Log("env: %s", env)

	service.Local = env == "local"

	service.Env = env

	service.Router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
	service.Router.HandleFunc("/status", service.statusHandlerFunc())

	service.healthHandler = transport.HealthHandlerFunc()

	service.Context, service.ContextCancelFunc = context.WithCancel(context.Background())

	service.runStatusUpdateLoop()

	service.setupGCPStorage()

	return &service
}

func (service *Service) LoadDatabase() {

	databasePath := envvar.GetString("DATABASE_PATH", "database.bin")
	overlayPath := envvar.GetString("OVERLAY_PATH", "overlay.bin")

	service.database, service.databaseOverlay = loadDatabase(databasePath, overlayPath)

	if service.database == nil {
		core.Error("load database failed: %s", databasePath)
		os.Exit(1)
	}

	applyOverlay(service.database, service.databaseOverlay)

	service.databaseRelayData = generateRelayData(service.database)
	if service.databaseRelayData == nil {
		core.Error("generate relay data failed")
		os.Exit(1)
	}

	service.watchDatabase(service.Context, databasePath, overlayPath)
}

func (service *Service) Database() *routing.DatabaseBinWrapper {
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

func (service *Service) OverrideHealthHandler(healthHandler func(w http.ResponseWriter, r *http.Request)) {
	service.healthHandler = healthHandler
}

func (service *Service) StartWebServer() {
	port := envvar.GetString("HTTP_PORT", "80")
	core.Log("starting http server on port %s", port)
	service.Router.HandleFunc("/health", service.healthHandler)
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

func (service *Service) LeaderElection() {
	core.Log("started leader election")
	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.GetString("REDIS_PASSWORD", "")
	config := RedisLeaderElectionConfig{}
	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.ServiceName = service.ServiceName
	var err error
	service.leaderElection, err = CreateRedisLeaderElection(service.Context, config)
	if err != nil {
		core.Error("could not create redis leader election: %v")
		os.Exit(1)
	}
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:
				service.leaderElection.Update(service.Context)
			}
		}
	}()
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

				var newDatabase routing.DatabaseBinWrapper

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

func (service *Service) RouteMatrixAndDatabase() (*RouteMatrix, *routing.DatabaseBinWrapper) {
	service.routeMatrixMutex.RLock()
	routeMatrix := service.routeMatrix
	database := service.routeMatrixDatabase
	service.routeMatrixMutex.RUnlock()
	return routeMatrix, database
}

func (service *Service) IsLeader() bool {
	if service.leaderElection != nil {
		return service.leaderElection.IsLeader()
	} else {
		return false
	}
}

func (service *Service) WaitForShutdown() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	core.Log("received shutdown signal")

	service.gcpStorage.Client.Close()

	// todo: wait group
	core.Log("successfully shutdown")
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
	err = backend.DecodeBinWrapper(databaseFile, database)
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

func generateRelayData(database *routing.DatabaseBinWrapper) *RelayData {

	relayData := &RelayData{}

	numRelays := len(database.Relays)

	relayData.NumRelays = numRelays
	relayData.RelayIds = make([]uint64, numRelays)
	relayData.RelayHash = make(map[uint64]routing.Relay)
	relayData.RelayArray = database.Relays

	sort.SliceStable(relayData.RelayArray, func(i, j int) bool {
		return relayData.RelayArray[i].Name < relayData.RelayArray[j].Name
	})

	for i := range relayData.RelayArray {
		if relayData.RelayArray[i].State != routing.RelayStateEnabled {
			core.Error("generateRelayData: database.bin must contain only enabled relays!")
			return nil
		}
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
		relayData.RelayLatitudes[i] = float32(relayData.RelayArray[i].Datacenter.Location.Latitude)
		relayData.RelayLongitudes[i] = float32(relayData.RelayArray[i].Datacenter.Location.Longitude)
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
	relayData.DestRelayNames = []string{}

	for _, buyer := range database.BuyerMap {
		if buyer.Live {
			for _, datacenter := range database.DatacenterMaps[buyer.ID] {
				datacenterRelays := relayData.DatacenterRelays[datacenter.DatacenterID]
				for j := 0; j < len(datacenterRelays); j++ {
					if !relayData.DestRelays[j] {
						relayData.DestRelayNames = append(relayData.DestRelayNames, relayData.RelayArray[j].Name)
						relayData.DestRelays[j] = true
					}
				}
			}
		}
	}

	sort.Strings(relayData.DestRelayNames)

	// stash the database bin file in the relay data, so it's all guaranteed to be consistent

	var databaseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&databaseBuffer)
	encoder.Encode(database)

	relayData.DatabaseBinFile = databaseBuffer.Bytes()

	return relayData
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

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				newDatabase, newOverlay := loadDatabase(databasePath, overlayPath)

				if newDatabase == nil {
					continue
				}

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

	if len(buffer) != 24 {
		return nil, errors.New(fmt.Sprintf("expected magic data to be 24 bytes, got %d", len(buffer)))
	}

	return buffer, nil
}

func (service *Service) updateMagicValues(magicData []byte) {

	if bytes.Equal(magicData, service.magicData) {
		return
	}

	service.magicMutex.Lock()
	service.magicData = magicData
	service.upcomingMagic = magicData[0:8]
	service.currentMagic = magicData[8:16]
	service.previousMagic = magicData[16:24]
	service.magicMutex.Unlock()

	core.Debug("updated magic values: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
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

func (service *Service) setupGCPStorage() {

	storageLocation := envvar.GetString("GCP_STORAGE_BUCKET", "gs://happy_path_testing")

	core.Log("gcp storage location: %s", storageLocation)

	gcpStorage, err := NewGCPStorageClient(service.Context, storageLocation)
	if err != nil {
		core.Error("failed to create gcp storage client: %v", err)
		os.Exit(1)
	}

	service.gcpStorage = gcpStorage
}

func (service *Service) DownloadFileFromGCPBucket(storageFileLocation string, localFileLocation string) error {
	if err := service.gcpStorage.CopyFromBucketToLocal(service.Context, storageFileLocation, localFileLocation); err != nil {
		return errors.New(fmt.Sprintf("failed to copy maxmind ISP file to GCP Cloud Storage: %v", err))
	}

	return nil
}

func (service *Service) UploadFileToGCPBucket(localFileLocation string, storageFileLocation string) error {
	if err := service.gcpStorage.CopyFromLocalToBucket(service.Context, localFileLocation, storageFileLocation); err != nil {
		return errors.New(fmt.Sprintf("failed to copy maxmind ISP file to GCP Cloud Storage: %v", err))
	}

	return nil
}

func (service *Service) UploadFileToGCPVirtualMachines(fileLocation string, uploadPath string, vmNames []string) error {

	hadError := false
	for _, vm := range vmNames {
		if err := service.gcpStorage.CopyFromLocalToRemote(service.Context, fileLocation, uploadPath, vm); err != nil {
			core.Error("failed to copy file to vm: %v", err)
			hadError = true
		}
	}

	if hadError {
		return errors.New("failed to upload file to one or more vms")
	}

	return nil
}

func (service *Service) DownloadGzipFileFromURL(url string, outputLocation string) error {

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	httpResponse, err := httpClient.Get(url)
	if err != nil {
		return err
	}

	if httpResponse.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http call returned status code: %s", httpResponse.Status))
	}

	// Decompress file in memory
	gz, err := gzip.NewReader(httpResponse.Body)
	if err != nil {
		return err
	}

	fileBuffer := bytes.NewBuffer(nil)
	tr := tar.NewReader(gz)
	for {
		var hdr *tar.Header

		hdr, err = tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New(fmt.Sprintf("failed to read from GZIP file: %v", err))
		}

		if strings.HasSuffix(hdr.Name, "mmdb") {
			_, err = io.Copy(fileBuffer, tr)
			if err != nil {
				return errors.New(fmt.Sprintf("failed to copy file data to buffer: %v", err))
			}
		}
	}

	gz.Close()
	httpResponse.Body.Close()

	// Write file to disk
	filePath, err := os.Create(outputLocation)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create file at %s: %v", outputLocation, err))
	}

	_, err = io.Copy(filePath, fileBuffer)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to write file to disk at %s: %v", outputLocation, err))
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------

func (service *Service) GetMIGInstanceInfo(migName string) []InstanceInfo {
	return service.gcpStorage.GetMIGInstanceInfo(migName)
}
