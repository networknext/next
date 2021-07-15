package relay_frontend

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

const (
	MatrixTypeCost   = "cost"
	MatrixTypeNormal = "normal"
)

type RelayFrontendConfig struct {
	Env                    string
	MasterTimeVariance     time.Duration
	UpdateRetryCount       int
	MatrixStoreAddress     string
	MatrixStorePassword    string
	MSMaxIdleConnections   int
	MSMaxActiveConnections int
	MSReadTimeout          time.Duration
	MSWriteTimeout         time.Duration
	MSMatrixExpireTimeout  time.Duration
}

type RelayFrontendSvc struct {
	RetryCount int

	cfg                         *RelayFrontendConfig
	id                          uint64
	store                       storage.MatrixStore
	createdAt                   time.Time
	currentMasterBackendAddress string

	// cached matrix
	costMatrix  *helpers.MatrixData
	routeMatrix *helpers.MatrixData
}

func NewRelayFrontend(store storage.MatrixStore, cfg *RelayFrontendConfig) (*RelayFrontendSvc, error) {
	rand.Seed(time.Now().UnixNano())
	r := new(RelayFrontendSvc)
	r.RetryCount = 0
	r.cfg = cfg
	r.id = rand.Uint64()
	r.store = store
	r.createdAt = time.Now().UTC()
	r.costMatrix = new(helpers.MatrixData)
	r.routeMatrix = new(helpers.MatrixData)
	return r, nil
}

func (r *RelayFrontendSvc) UpdateRelayBackendMaster() error {
	rbArr, err := r.store.GetRelayBackendLiveData()
	if err != nil {
		r.currentMasterBackendAddress = ""
		return err
	}

	masterAddress, err := chooseRelayBackendMaster(rbArr, r.cfg.MasterTimeVariance)
	if err != nil {
		r.currentMasterBackendAddress = ""
		return err
	}

	r.currentMasterBackendAddress = masterAddress

	return nil
}

func (r *RelayFrontendSvc) CacheMatrix(matrixType string) error {
	matrixAddr, err := r.GetMatrixAddress(matrixType)
	if err != nil {
		return err
	}

	return r.cacheMatrixInternal(matrixAddr, matrixType)
}

// Gets the latest matrix from the master relay backend via HTTP and caches it internally
func (r *RelayFrontendSvc) cacheMatrixInternal(matrixAddr, matrixType string) error {
	matrixBin, err := getHttpMatrix(matrixAddr)
	if err != nil {
		return err
	}

	switch matrixType {
	case MatrixTypeCost:
		r.costMatrix.SetMatrix(matrixBin)
	case MatrixTypeNormal:
		r.routeMatrix.SetMatrix(matrixBin)
	}

	return nil
}

// Determines if we have reached the retry limit for selecting the master relay backend
func (r *RelayFrontendSvc) ReachedRetryLimit() bool {
	return r.RetryCount > r.cfg.UpdateRetryCount
}

/*
	ResetCachedMatrix() sets the cached matrix type to an empty struct. This is used
	when we fail to choose a master relay backend for UpdateRetryCount times so that we
	do not provide a stale matrix to the server backend.
*/
func (r *RelayFrontendSvc) ResetCachedMatrix(matrixType string) error {
	switch matrixType {
	case MatrixTypeCost:
		emptyCostMatrix := routing.CostMatrix{Version: routing.CostMatrixSerializeVersion}
		err := emptyCostMatrix.WriteResponseData(10000)
		if err != nil {
			return err
		}

		emptyCostMatrixBin := emptyCostMatrix.GetResponseData()
		r.costMatrix.SetMatrix(emptyCostMatrixBin)
	case MatrixTypeNormal:
		emptyrouteMatrix := routing.RouteMatrix{Version: routing.RouteMatrixSerializeVersion}
		err := emptyrouteMatrix.WriteResponseData(10000)
		if err != nil {
			return err
		}

		emptyrouteMatrixBin := emptyrouteMatrix.GetResponseData()
		r.routeMatrix.SetMatrix(emptyrouteMatrixBin)
	}

	return nil
}

/*
	chooseRelayBackendMaster() selects the oldest relay backend (in terms of uptime) as the master,
	as long as the it has updated its status in redis in the last timeVariance seconds. In the event
	that multiple relay backends have the same uptime, the one with a lesser VM instanceID will be chosen
	as the master.

	The chosen relay backend produces the cost and route matrix that are cached and provided to the server backend.
*/
func chooseRelayBackendMaster(rbArr []storage.RelayBackendLiveData, timeVariance time.Duration) (string, error) {
	currentTime := time.Now().UTC()
	masterRB := storage.NewRelayBackendLiveData("", "", currentTime, currentTime)

	for _, rb := range rbArr {
		if currentTime.Sub(rb.UpdatedAt) > timeVariance {
			continue
		}
		if rb.InitAt.After(masterRB.InitAt) {
			continue
		}
		if rb.InitAt.Equal(masterRB.InitAt) && rb.ID < masterRB.ID {
			continue
		}
		masterRB = rb
	}

	if masterRB.ID == "" {
		return "", fmt.Errorf("relay backend master not found")
	}

	return masterRB.Address, nil
}

// Gets a matrix from the master relay backend using HTTP
func getHttpMatrix(address string) ([]byte, error) {
	resp, err := http.Get(address)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	buffer, err := ioutil.ReadAll(resp.Body)
	if len(buffer) == 0 {
		return []byte{}, errors.New("empty resp body")
	}

	return buffer, nil
}

// Gets the full address for a matrix type
func (r *RelayFrontendSvc) GetMatrixAddress(matrixType string) (string, error) {
	var addr string
	switch matrixType {
	case MatrixTypeCost:
		addr = fmt.Sprintf("http://%s/cost_matrix", r.currentMasterBackendAddress)
	case MatrixTypeNormal:
		addr = fmt.Sprintf("http://%s/route_matrix", r.currentMasterBackendAddress)
	default:
		return "", errors.New("matrix type not supported")
	}
	return addr, nil
}

// Gets the cached route matrix
func (r *RelayFrontendSvc) GetRouteMatrixHandlerFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		data := r.routeMatrix.GetMatrix()
		if len(data) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// Generic handler to get handler information from the master relay backend
func (r *RelayFrontendSvc) GetRelayBackendHandlerFunc(endpoint string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp, err := http.Get(fmt.Sprintf("http://%s/%s", r.currentMasterBackendAddress, endpoint))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		bin, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bin)
	}
}

// Gets the IP address of the master relay backend
func (r *RelayFrontendSvc) GetRelayBackendMasterHandlerFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		bin := []byte(r.currentMasterBackendAddress)

		w.WriteHeader(http.StatusOK)
		w.Write(bin)
	}
}

// Gets the relay dashboard from the master relay backend via basic authentication
func (r *RelayFrontendSvc) GetRelayDashboardHandlerFunc(username string, password string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		u, p, _ := req.BasicAuth()
		if u != username && p != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/relay_dashboard", r.currentMasterBackendAddress), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req.SetBasicAuth(username, password)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		bin, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bin)
	}
}

func (r *RelayFrontendSvc) GetRouteMatrix() []byte {
	return r.routeMatrix.GetMatrix()
}

func (r *RelayFrontendSvc) GetCostMatrix() []byte {
	return r.costMatrix.GetMatrix()
}

func (r *RelayFrontendSvc) GetRelayDashboardDataHandlerFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/relay_dashboard_data", r.currentMasterBackendAddress), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		jsonData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func (r *RelayFrontendSvc) GetRelayDashboardAnalysisHandlerFunc() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/relay_dashboard_analysis", r.currentMasterBackendAddress), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		jsonData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

func (r *RelayFrontendSvc) GetCostMatrixHandlerFunc() func(w http.ResponseWriter, req *http.Request) {

	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		if r.currentMasterBackendAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		client := &http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/cost_matrix", r.currentMasterBackendAddress), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		jsonData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}
