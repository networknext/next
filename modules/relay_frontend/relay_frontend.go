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

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

const (
	MatrixTypeCost   = "cost"
	MatrixTypeNormal = "normal"
	MatrixTypeValve  = "valve"
)

type RelayFrontendSvc struct {
	cfg                         *Config
	id                          uint64
	store                       storage.MatrixStore
	createdAt                   time.Time
	currentMasterBackendAddress string

	// cached matrix
	costMatrix       *helpers.MatrixData
	routeMatrix      *helpers.MatrixData
	routeMatrixValve *helpers.MatrixData
}

func New(store storage.MatrixStore, cfg *Config) (*RelayFrontendSvc, error) {
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
	rbArr, err := r.store.GetRelayBackendLiveData(r.cfg.RelayBackendAddresses)
	if err != nil {
		return err
	}

	masterAddress, err := chooseRelayBackendMaster(rbArr, r.cfg.MasterTimeVariance)
	if err != nil {
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

func (r *RelayFrontendSvc) cacheMatrixInternal(matrixAddr, matrixType string) error {
	matrix, err := getHttpMatrix(matrixAddr)
	if err != nil {
		return err
	}

	switch matrixType {
	case MatrixTypeCost:
		r.costMatrix.SetMatrix(matrix)
	case MatrixTypeNormal:
		r.routeMatrix.SetMatrix(matrix)
	case MatrixTypeValve:
		r.routeMatrixValve.SetMatrix(matrix)
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
		if rb.InitAt.Equal(masterRB.InitAt) && rb.Id < masterRB.Id {
			continue
		}
		masterRB = rb
	}

	if masterRB.Id == "" {
		return "", fmt.Errorf("relay backend master not found")
	}

	return masterRB.Address, nil
}

func getHttpMatrix(address string) ([]byte, error) {
	resp, err := http.Get(address)
	if err != nil {
		return []byte{}, err
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if len(buffer) == 0 {
		return []byte{}, errors.New("empty resp body")
	}
	fmt.Print()
	var newRouteMatrix routing.RouteMatrix
	rs := encoding.CreateReadStream(buffer)
	if err := newRouteMatrix.Serialize(rs); err != nil {
		return []byte{}, err
	}

	return buffer, nil
}

func (r *RelayFrontendSvc) GetMatrixAddress(matrixType string) (string, error) {
	var addr string
	switch matrixType {
	case MatrixTypeCost:
		addr = fmt.Sprintf("http:/%s/cost_matrix", r.currentMasterBackendAddress)
	case MatrixTypeNormal:
		addr = fmt.Sprintf("http:/%s/route_matrix", r.currentMasterBackendAddress)
	case MatrixTypeValve:
		addr = fmt.Sprintf("http:/%s/route_matrix_valve", r.currentMasterBackendAddress)
	default:
		return "", errors.New("matrix type not supported")
	}
	return addr, nil
}

func (r *RelayFrontendSvc) GetCostMatrix() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(r.costMatrix.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (r *RelayFrontendSvc) GetRouteMatrix() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(r.routeMatrix.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (r *RelayFrontendSvc) GetRouteMatrixValve() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(r.routeMatrixValve.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
