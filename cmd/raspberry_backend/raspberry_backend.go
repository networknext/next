package main

import (
	"net"
	"fmt"
	"time"
	"net/http"
	"context"
	"sync"

	"github.com/go-redis/redis/v8"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var redisHostname string
var redisPassword string

var magicUpdateSeconds int

type Update struct {
	Address *net.UDPAddr
}

var updateChannel chan *Update

var redisClient *redis.Client

var serversMutex sync.RWMutex
var servers string

const RaspberryBackendVersion = 1 // IMPORTANT: bump this anytime you change the redis data structures!

func main() {

	service := common.CreateService("raspberry_backend")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis password: %s", redisPassword)

	ctx := context.Background()

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisHostname,
		Password: redisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	processUpdates()

	updateServers()

	service.Router.HandleFunc("/server_update", serverUpdateHandler)
	service.Router.HandleFunc("/servers", serversHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

func processUpdates() {
	go func() {
		for {
			update := <-updateChannel
			fmt.Printf("processing update\n")
			_ = update
		}
	}()
}

func updateServers() {
	go func() {
		for {
			time.Sleep(time.Second)
			serversMutex.Lock()
			// todo: get list of servers from redis
			servers = "todo"
			serversMutex.Unlock()
		}
	}()
}

func serverUpdateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("server update\n")
	update := &Update{}
	// todo: parse update into update struct
	updateChannel <- update
}

func serversHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("servers\n")
	w.WriteHeader(http.StatusOK)
	serversMutex.RLock()
	w.Write([]byte(servers))
	serversMutex.RUnlock()
}
