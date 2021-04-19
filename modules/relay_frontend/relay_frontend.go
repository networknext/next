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

	"github.com/networknext/backend/modules/storage"
)

const (
	MatrixTypeCost   = "cost"
	MatrixTypeNormal = "normal"
)

type RelayFrontendSvc struct {
	cfg                         *RelayFrontendConfig
	id                          uint64
	store                       storage.MatrixStore
	createdAt                   time.Time
	currentMasterBackendAddress string
	relayStatsAddress           string

	// cached matrix
	costMatrix       *helpers.MatrixData
	routeMatrix      *helpers.MatrixData
	routeMatrixValve *helpers.MatrixData
}

func NewRelayFrontend(store storage.MatrixStore, cfg *RelayFrontendConfig) (*RelayFrontendSvc, error) {
	rand.Seed(time.Now().UnixNano())
	r := new(RelayFrontendSvc)
	r.cfg = cfg
	r.id = rand.Uint64()
	r.store = store
	r.createdAt = time.Now().UTC()
	r.costMatrix = new(helpers.MatrixData)
	r.routeMatrix = new(helpers.MatrixData)
	r.routeMatrixValve = new(helpers.MatrixData)
	return r, nil
}

func (r *RelayFrontendSvc) UpdateRelayBackendMaster() error {
	rbArr, err := r.store.GetRelayBackendLiveData()
	if err != nil {
		return err
	}

	masterAddress, err := chooseRelayBackendMaster(rbArr, r.cfg.MasterTimeVariance)
	if err != nil {
		r.currentMasterBackendAddress = ""
		r.relayStatsAddress = ""
		return err
	}

	r.currentMasterBackendAddress = masterAddress
	r.relayStatsAddress = fmt.Sprintf("http://%s/relay_stats", r.currentMasterBackendAddress)

	return nil
}

func (r *RelayFrontendSvc) CacheMatrix(matrixType string) error {
	matrixAddr, err := r.GetMatrixAddress(matrixType)
	if err != nil {
		return err
	}

	return r.cacheMatrixInternal(matrixAddr, matrixType)
}

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

func (r *RelayFrontendSvc) GetCostMatrix() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		data := r.costMatrix.GetMatrix()
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

func (r *RelayFrontendSvc) GetRouteMatrix() func(w http.ResponseWriter, req *http.Request) {
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

func (r *RelayFrontendSvc) GetRelayStats() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		if r.relayStatsAddress == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp, err := http.Get(r.relayStatsAddress)
		defer resp.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bin, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bin)
	}
}
