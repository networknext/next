package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/handlers"
)

var service *common.Service

var maxPacketSize int
var serverBackendAddress net.UDPAddr
var routeMatrixURI string
var routeMatrixInterval time.Duration
var privateKey []byte

var routeMatrixMutex sync.RWMutex
var routeMatrix *common.RouteMatrix
var database *routing.DatabaseBinWrapper

func main() {

	service = common.CreateService("new_server_backend5")

	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	privateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})

	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval.String())

	updateRouteMatrix()

	service.OverrideHealthHandler(healthHandler)

	service.StartUDPServer(packetHandler)

	service.UpdateMagic()

	service.StartWebServer()

	service.WaitForShutdown()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	routeMatrixMutex.RLock()
	not_ready := routeMatrix == nil
	routeMatrixMutex.RUnlock()

	if not_ready {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	handler := handlers.SDK5_HandlerData{}

	routeMatrixMutex.Lock()
	handler.Database = database
	handler.RouteMatrix = routeMatrix
	routeMatrixMutex.Unlock()

	handler.MaxPacketSize = maxPacketSize
	handler.ServerBackendAddress = serverBackendAddress
	handler.PrivateKey = privateKey
	handler.GetMagicValues = func() ([]byte, []byte, []byte) {
		return service.GetMagicValues()
	}

	handlers.SDK5_PacketHandler(&handler, conn, from, packetData)
}

func updateRouteMatrix() {

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

				routeMatrixMutex.Lock()
				currentRouteMatrix := routeMatrix
				routeMatrixMutex.Unlock()

				if currentRouteMatrix != nil && time.Now().Unix()-int64(currentRouteMatrix.CreatedAt) > 30 {
					core.Error("route matrix is stale")
					routeMatrixMutex.Lock()
					routeMatrix = nil
					routeMatrixMutex.Unlock()
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

				newRouteMatrix := common.RouteMatrix{}

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

				routeMatrixMutex.Lock()
				routeMatrix = &newRouteMatrix
				database = &newDatabase
				routeMatrixMutex.Unlock()

				core.Debug("updated route matrix: %d relays", len(routeMatrix.RelayIds))
			}
		}
	}()
}
