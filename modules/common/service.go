package common

import (
	"bytes"
	"context"
	"encoding/binary"
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

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/ip2location"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/oschwald/maxminddb-golang"
	"github.com/rs/cors"
    "cloud.google.com/go/profiler"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
	tag           string
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
	RelaySellerIds     []uint64
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
	Tag           string
	Local         bool

	Router mux.Router

	Context           context.Context
	ContextCancelFunc context.CancelFunc

	GoogleProjectId string

	// ------------------

	databaseMutex     sync.RWMutex
	database          *db.Database
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
	ready            func() bool

	udpServer *UDPServer

	routeMatrixMutex    sync.RWMutex
	routeMatrix         *RouteMatrix
	routeMatrixDatabase *db.Database

	google *GoogleCloudHandler

	ip2location_mutex   sync.RWMutex
	ip2location_isp_db  *maxminddb.Reader
	ip2location_city_db *maxminddb.Reader
}

func CreateService(serviceName string) *Service {

	service := Service{}
	service.ServiceName = serviceName
	service.CommitMessage = commitMessage
	service.CommitHash = commitHash
	service.BuildTime = buildTime
	service.Tag = tag

	core.Log("%s", service.ServiceName)

	env := envvar.GetString("ENV", "")

	core.Log("env: %s", env)

	service.Local = env == "local"

	service.Env = env

	if service.Tag != "" {
		core.Log("tag: %s", service.Tag)
	}

	if service.CommitMessage != "" {
		core.Log("commit message: %s", service.CommitMessage)
	}

	if service.CommitHash != "" {
		core.Log("commit hash: %s", service.CommitHash)
	}

	if service.BuildTime != "" {
		core.Log("build time: %s", service.BuildTime)
	}

	if envvar.GetBool("ENABLE_PROFILER", false) {

		core.Log("profiler is enabled")

	    profilerConfig := profiler.Config{
	    	Service: 		serviceName,
	        ServiceVersion: tag,
	    }

	    if err := profiler.Start(profilerConfig); err != nil {
			core.Error("could not start profiler")
			os.Exit(1)
	    }
	}

	service.Router.HandleFunc("/version", versionHandlerFunc(buildTime, commitMessage, commitHash, tag, []string{}))
	service.Router.HandleFunc("/status", service.statusHandlerFunc())
	service.Router.HandleFunc("/database", service.databaseHandlerFunc())
	service.Router.HandleFunc("/lb_health", service.lbHealthHandlerFunc())
	service.Router.HandleFunc("/vm_health", service.vmHealthHandlerFunc())
	service.Router.HandleFunc("/ready", service.readyHandlerFunc())

	service.Context, service.ContextCancelFunc = context.WithCancel(context.Background())

	service.runStatusUpdateLoop()

	service.GoogleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "")
	if service.GoogleProjectId != "" {
		core.Log("google project id: %s", service.GoogleProjectId)
		google, err := NewGoogleCloudHandler(service.Context, service.GoogleProjectId)
		if err != nil {
			core.Error("failed to create google cloud handler: %v", err)
			os.Exit(1)
		}
		service.google = google
	}

	service.sendTrafficToMe = func() bool { return true }
	service.machineIsHealthy = func() bool { return true }
	service.ready = func() bool { return true }

	return &service
}

func (service *Service) SetHealthFunctions(sendTrafficToMe func() bool, machineIsHealthy func() bool, ready func() bool) {
	service.sendTrafficToMe = sendTrafficToMe
	service.machineIsHealthy = machineIsHealthy
	service.ready = ready
}

func (service *Service) LoadDatabase() {

	databasePath := envvar.GetString("DATABASE_PATH", "database.bin")

	var err error

	service.database, err = db.LoadDatabase(databasePath)
	if err != nil {
		core.Error("could not read database: %v", err)
		os.Exit(1)
	}

	if service.database == nil {
		core.Error("database is nil")
		os.Exit(1)
	}

	err = service.database.Validate()
	if err != nil {
		core.Error("database does not validate: %v", err)
		os.Exit(1)
	}

	service.databaseRelayData = generateRelayData(service.database)
	if service.databaseRelayData == nil {
		core.Error("generate relay data failed")
		os.Exit(1)
	}

	core.Log("loaded database: %s", databasePath)

	service.watchDatabase(service.Context, databasePath)
}

func (service *Service) LoadIP2Location() {

	bucketName := envvar.GetString("IP2LOCATION_BUCKET_NAME", "")

	if bucketName == "" {
		core.Error("you must specify ip2location bucket name")
		os.Exit(1)
	}

	err := ip2location.DownloadDatabases_CloudStorage(bucketName)
	if err != nil {
		core.Error("could not download ip2location databases from cloud storage: %v", err)
		os.Exit(1)
	}

	isp_db, city_db, err := ip2location.LoadDatabases()

	if err != nil {
		core.Error("failed to load ip2location databases: %v", err)
		os.Exit(1)
	}

	service.ip2location_mutex.Lock()
	service.ip2location_isp_db = isp_db
	service.ip2location_city_db = city_db
	service.ip2location_mutex.Unlock()

	if bucketName != "" {

		go func() {
			for {
				time.Sleep(time.Hour)

				core.Log("updating ip2location databases")

				var isp_db, city_db *maxminddb.Reader

				err := ip2location.DownloadDatabases_CloudStorage(bucketName)
				if err != nil {
					core.Warn("failed to download ip2location databases from cloud storage: %v")
					continue
				}

				isp_db, city_db, err = ip2location.LoadDatabases()
				if err != nil {
					core.Warn("failed to load ip2location databases: %v", err)
					continue
				}

				service.ip2location_mutex.Lock()
				service.ip2location_isp_db = isp_db
				service.ip2location_city_db = city_db
				service.ip2location_mutex.Unlock()
			}
		}()
	}
}

func (service *Service) GetLocation(ip net.IP) (float32, float32) {
	service.ip2location_mutex.RLock()
	city_db := service.ip2location_city_db
	service.ip2location_mutex.RUnlock()
	return ip2location.GetLocation(city_db, ip)
}

func (service *Service) GetISP(ip net.IP) string {
	service.ip2location_mutex.RLock()
	isp_db := service.ip2location_isp_db
	service.ip2location_mutex.RUnlock()
	return ip2location.GetISP(isp_db, ip)
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
	return database.GetBinary()
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

func (service *Service) GetMagicValues() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
	service.magicMutex.Lock()
	a := service.upcomingMagic
	b := service.currentMagic
	c := service.previousMagic
	service.magicMutex.Unlock()
	upcomingMagic := [constants.MagicBytes]byte{}
	currentMagic := [constants.MagicBytes]byte{}
	previousMagic := [constants.MagicBytes]byte{}
	copy(upcomingMagic[:], a)
	copy(currentMagic[:], b)
	copy(previousMagic[:], c)
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

func (service *Service) readyHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		if !service.ready() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func versionHandlerFunc(buildTime string, commitMessage string, commitHash string, tag string, allowedOrigins []string) func(w http.ResponseWriter, r *http.Request) {

	version := map[string]string{
		"build_time":     buildTime,
		"commit_message": commitMessage,
		"commit_hash":    commitHash,
		"tag":            tag,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (service *Service) StartWebServer() {
	port := envvar.GetString("HTTP_PORT", "80")
	allowedOrigin := envvar.GetString("ALLOWED_ORIGIN", "")
	core.Log("starting http server on port %s", port)
	go func() {
		bindAddress := ":" + port
		if service.Local {
			bindAddress = "127.0.0.1:" + port
		}
		if allowedOrigin == "" {
			// standard
			err := http.ListenAndServe(bindAddress, &service.Router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				os.Exit(1)
			}
		} else {
			// CORS
			core.Log("allowed origin: %s", allowedOrigin)
			c := cors.New(cors.Options{
				AllowedOrigins:   []string{allowedOrigin},
				AllowCredentials: true,
			})
			err := http.ListenAndServe(bindAddress, c.Handler(&service.Router))
			if err != nil {
				core.Error("error starting http server: %v", err)
				os.Exit(1)
			}
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
	config.BindAddress = envvar.GetAddress("UDP_BIND_ADDRESS", core.ParseAddress(fmt.Sprintf("0.0.0.0:%d", config.Port)))
	if service.Local {
		config.BindAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", config.Port))
	}
	core.Log("starting udp server on port %d", config.Port)
	core.Debug("udp num threads: %d", config.NumThreads)
	core.Debug("udp socket read buffer: %d", config.SocketReadBuffer)
	core.Debug("udp socket write buffer: %d", config.SocketWriteBuffer)
	core.Debug("udp max packet size: %d", config.MaxPacketSize)
	service.udpServer = CreateUDPServer(service.Context, config, packetHandler)
}

func CreateRedisClient(hostname string) redis.Conn {
	redisClient, err := redis.Dial("tcp", hostname)
	if err != nil {
		panic(err)
	}
	redisClient.Send("PING")
	redisClient.Flush()
	pong, err := redisClient.Receive()
	if err != nil || pong != "PONG" {
		panic(err)
	}
	return redisClient
}

func CreateRedisPool(hostname string, active int, idle int) *redis.Pool {
	pool := redis.Pool{
		MaxActive:   active,
		MaxIdle:     idle,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", hostname)
		},
	}
	redisClient := pool.Get()
	redisClient.Send("PING")
	redisClient.Flush()
	pong, err := redisClient.Receive()
	if err != nil || pong != "PONG" {
		panic(err)
	}
	return &pool
}

func (service *Service) LeaderElection() {

	core.Log("started leader election")

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 100)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 1000)

	pool := CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

	config := RedisLeaderElectionConfig{}
	config.Timeout = time.Second * 10
	config.ServiceName = service.ServiceName

	var err error
	service.leaderElection, err = CreateRedisLeaderElection(pool, config)
	if err != nil {
		core.Error("could not create redis leader election: %v")
		os.Exit(1)
	}

	service.leaderElection.Start(service.Context)
}

func (service *Service) Store(name string, data []byte) {
	core.Debug("store %s (%d bytes)", name, len(data))
	if service.leaderElection == nil {
		panic("leader election must be enabled to call store")
	}
	service.leaderElection.Store(service.Context, name, data)
}

func (service *Service) Load(name string) []byte {
	if service.leaderElection == nil {
		panic("leader election must be enabled to call load")
	}
	data := service.leaderElection.Load(service.Context, name)
	core.Debug("loaded %s (%d bytes)", name, len(data))
	return data
}

func (service *Service) UpdateRouteMatrix() {

	routeMatrixURL := envvar.GetString("ROUTE_MATRIX_URL", "http://127.0.0.1:30001/route_matrix")
	routeMatrixInterval := envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	core.Log("route matrix url: %s", routeMatrixURL)
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

				start := time.Now()

				service.routeMatrixMutex.RLock()
				currentRouteMatrix := service.routeMatrix
				service.routeMatrixMutex.RUnlock()

				if currentRouteMatrix != nil && time.Now().Unix()-int64(currentRouteMatrix.CreatedAt) > 30 {
					core.Error("route matrix is stale: created at %d, current time is %d (%d seconds old)", int64(currentRouteMatrix.CreatedAt), time.Now().Unix(), time.Now().Unix()-int64(currentRouteMatrix.CreatedAt))
					service.routeMatrixMutex.Lock()
					service.routeMatrix = nil
					service.routeMatrixMutex.Unlock()
				}

				response, err := httpClient.Get(routeMatrixURL)
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read response body: %v", err)
					continue
				}

				response.Body.Close()

				if len(buffer) == 0 {
					core.Debug("route matrix is empty")
					continue
				}

				newRouteMatrix := RouteMatrix{}

				err = newRouteMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				var newDatabase db.Database

				err = newDatabase.LoadBinary(newRouteMatrix.BinFileData)
				if err != nil {
					core.Error("failed to read database: %v", err)
					continue
				}

				service.routeMatrixMutex.Lock()
				service.routeMatrix = &newRouteMatrix
				service.routeMatrixDatabase = &newDatabase
				service.routeMatrixMutex.Unlock()

				duration := time.Since(start).Milliseconds()

				core.Debug("updated route matrix: %d relays, %d bytes, fetched in %dms", len(newRouteMatrix.RelayIds), len(buffer), duration)

				if duration > routeMatrixInterval.Milliseconds() {
					core.Warn("update route matrix can't keep up!")
				}
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

func (service *Service) IsReady() bool {
	if service.leaderElection != nil {
		return service.leaderElection.IsReady()
	}
	return true
}

func (service *Service) WaitForShutdown() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	core.Log("received shutdown signal")

	service.ip2location_mutex.Lock()
	if service.ip2location_city_db != nil {
		service.ip2location_city_db.Close()
		service.ip2location_city_db = nil
	}
	if service.ip2location_isp_db != nil {
		service.ip2location_isp_db.Close()
		service.ip2location_isp_db = nil
	}
	service.ip2location_mutex.Unlock()

	core.Log("successfully shutdown")
}

// -----------------------------------------------------------------------

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
		relayData.RelayHash[relayData.RelayArray[i].Id] = relayData.RelayArray[i]
	}

	relayData.RelayAddresses = make([]net.UDPAddr, numRelays)
	relayData.RelayNames = make([]string, numRelays)
	relayData.RelayLatitudes = make([]float32, numRelays)
	relayData.RelayLongitudes = make([]float32, numRelays)
	relayData.RelaySellerIds = make([]uint64, numRelays)
	relayData.RelayDatacenterIds = make([]uint64, numRelays)

	for i := 0; i < numRelays; i++ {
		relayData.RelayIds[i] = relayData.RelayArray[i].Id
		relayData.RelayAddresses[i] = relayData.RelayArray[i].PublicAddress
		relayData.RelayNames[i] = relayData.RelayArray[i].Name
		relayData.RelayLatitudes[i] = float32(relayData.RelayArray[i].Datacenter.Latitude)
		relayData.RelayLongitudes[i] = float32(relayData.RelayArray[i].Datacenter.Longitude)
		relayData.RelaySellerIds[i] = relayData.RelayArray[i].Seller.Id
		relayData.RelayDatacenterIds[i] = relayData.RelayArray[i].Datacenter.Id
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
	for i := range relayData.DestRelays {
		relayData.DestRelays[i] = true
	}

	for _, buyer := range database.BuyerMap {
		if buyer.Live {
			for _, settings := range database.BuyerDatacenterSettings[buyer.Id] {
				if !settings.EnableAcceleration {
					continue
				}
				datacenterRelays := relayData.DatacenterRelays[settings.DatacenterId]
				for j := 0; j < len(datacenterRelays); j++ {
					relayData.DestRelays[datacenterRelays[j]] = true
				}
			}
		}
	}

	// stash the database bin file in the relay data, so it's all guaranteed to be consistent

	relayData.DatabaseBinFile = database.GetBinary()

	return relayData
}

func (service *Service) watchDatabase(ctx context.Context, databasePath string) {

	databaseURL := envvar.GetString("DATABASE_URL", "")

	if databaseURL == "" {
		return
	}

	core.Debug("database url: %s", databaseURL)

	syncInterval := envvar.GetDuration("DATABASE_SYNC_INTERVAL", time.Minute)

	if service.google == nil {
		core.Error("You must pass in GOOGLE_PROJECT_ID")
		os.Exit(1)
	}

	go func() {

		ticker := time.NewTicker(syncInterval)

		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				if databaseURL != "" {

					// reload from google cloud storage

					core.Debug("reloading database from google cloud storage: %s", databaseURL)

					tempFile := databasePath + "-temp"

					service.google.CopyFromBucketToLocal(ctx, databaseURL, tempFile)

					newDatabase, err := db.LoadDatabase(tempFile)
					if err != nil {
						core.Warn("could not read new database: %v", err)
						break
					}

					if newDatabase == nil {
						core.Warn("new database is nil")
						break
					}

					err = service.database.Validate()
					if err != nil {
						core.Warn("new database does not validate: %v", err)
						break
					}

					newRelayData := generateRelayData(service.database)
					if newRelayData == nil {
						core.Warn("new database failed to generate relay data")
						break
					}

					os.Rename(tempFile, databasePath)

					service.databaseMutex.Lock()
					service.database = newDatabase
					service.databaseRelayData = newRelayData
					service.databaseMutex.Unlock()

				} else {

					// reload from disk

					core.Debug("reloading database from disk: %s", databasePath)

					newDatabase, err := db.LoadDatabase(databasePath)
					if err != nil {
						core.Warn("could not read new database: %v", err)
						break
					}

					if newDatabase == nil {
						core.Warn("new database is nil")
						break
					}

					err = service.database.Validate()
					if err != nil {
						core.Warn("new database does not validate: %v", err)
						break
					}

					newRelayData := generateRelayData(service.database)
					if newRelayData == nil {
						core.Warn("new database failed to generate relay data")
						break
					}

					service.databaseMutex.Lock()
					service.database = newDatabase
					service.databaseRelayData = newRelayData
					service.databaseMutex.Unlock()
				}
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

func (service *Service) databaseHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		database := service.Database()
		if database == nil {
			service.routeMatrixMutex.RLock()
			database = service.routeMatrixDatabase
			service.routeMatrixMutex.RUnlock()
		}
		if database != nil {
			database.WriteHTML(w)
			w.Header().Set("Content-Type", "text/html")
		} else {
			fmt.Fprintf(w, "no database\n")
			w.Header().Set("Content-Type", "text/plain")
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

	magicURL := envvar.GetString("MAGIC_URL", "http://127.0.0.1:41007/magic")
	magicInterval := envvar.GetDuration("MAGIC_INTERVAL", time.Second)

	core.Debug("magic url: %s", magicURL)

	httpClient := &http.Client{
		Timeout: time.Second,
	}

	var magicData []byte
	for i := 0; i < 10; i++ {
		var err error
		magicData, err = getMagic(httpClient, magicURL)
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

		ticker := time.NewTicker(magicInterval)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:
				magicData, err := getMagic(httpClient, magicURL)
				if err == nil {
					service.updateMagicValues(magicData)
				}
			}
		}
	}()
}

// ---------------------------------------------------------------------------------------------------
