package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

var redisHostname string

var magicUpdateSeconds int

type Update struct {
	address *net.UDPAddr
}

var updateChannel chan *Update

var redisClient *redis.Client

var serversMutex sync.RWMutex
var servers string

const RaspberryBackendVersion = 1 // IMPORTANT: bump this anytime you change the redis data structures!

func main() {

	service := common.CreateService("raspberry_backend")

	updateChannel = make(chan *Update, 10*1024)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	core.Debug("redis hostname: %s", redisHostname)

	ctx := context.Background()

	redisClient = redis.NewClient(&redis.Options{
		Addr: redisHostname,
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
	ctx := context.Background()
	go func() {
		for {
			update := <-updateChannel
			pipe := redisClient.TxPipeline()
			pipe.Set(ctx, fmt.Sprintf("raspberry-server-%d/%s", RaspberryBackendVersion, update.address.String()), "", 30*time.Second)
			_, err := pipe.Exec(ctx)
			if err != nil {
				core.Error("failed to store server update: %v", err)
				return
			}
		}
	}()
}

func parseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func updateServers() {
	ctx := context.Background()
	prefix := fmt.Sprintf("raspberry-server-%d/", RaspberryBackendVersion)
	go func() {
		for {
			time.Sleep(time.Second)
			newServers := ""
			itor := redisClient.Scan(ctx, 0, fmt.Sprintf("%s*", prefix), 0).Iterator()
			for itor.Next(ctx) {
				_, server, result := strings.Cut(itor.Val(), prefix)
				if result {
					newServers += fmt.Sprintf("%s\n", server)
				}
			}
			if err := itor.Err(); err != nil {
				fmt.Printf("failed to update servers: %v", err)
				continue
			}
			serversMutex.Lock()
			servers = newServers
			serversMutex.Unlock()
		}
	}()
}

func serverUpdateHandler(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	address := parseAddress(string(data))
	if address.IP == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Debug("server update from %s", address)
	updateChannel <- &Update{address: address}
	w.WriteHeader(http.StatusOK)
}

func serversHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	serversMutex.RLock()
	w.Write([]byte(servers))
	serversMutex.RUnlock()
}
