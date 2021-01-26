package relay_frontend

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

type RelayFrontendSvc struct {
	id                          uint64
	store                       storage.MatrixStore
	createdAt                   time.Time
	currentlyMaster             bool
	currentMasterOptimizer      uint64
	currentMasterBackendAddress string
	matrixSvcTimeVariance       time.Duration
	optimizerTimeVariance       time.Duration
	relayAddresses              []string

	// cached matrix
	routeMatrix      []byte
	routeMatrixValve []byte
	matrixMux        sync.RWMutex
}

func New(store storage.MatrixStore, matrixSvcTimeVariance, optimizerTimeVariance time.Duration) (*RelayFrontendSvc, error) {
	rand.Seed(time.Now().UnixNano())

	r := new(RelayFrontendSvc)
	r.id = rand.Uint64()
	r.store = store
	r.createdAt = time.Now().UTC()
	r.currentlyMaster = false
	r.matrixSvcTimeVariance = matrixSvcTimeVariance
	r.optimizerTimeVariance = optimizerTimeVariance

	return r, nil
}

func svcError(err error) error {
	return fmt.Errorf("route matrix svc: %s", err.Error())
}

func (r *RelayFrontendSvc) UpdateSvcDB() error {
	svcData := storage.MatrixSvcData{
		ID:        r.id,
		CreatedAt: r.createdAt,
		UpdatedAt: time.Now().UTC(),
	}

	err := r.store.UpdateMatrixSvc(svcData)
	if err != nil {
		return svcError(err)
	}
	return nil
}

func (r *RelayFrontendSvc) AmMaster() bool {
	return r.currentlyMaster
}

func (r *RelayFrontendSvc) DetermineMaster() error {
	matrixSvcs, err := r.store.GetMatrixSvcs()
	if err != nil {
		return svcError(err)
	}

	masterId, err := r.store.GetMatrixSvcMaster()
	if err != nil && err.Error() != "matrix svc master not found" {
		return svcError(err)
	}

	if !isMasterMatrixSvcValid(matrixSvcs, masterId, r.matrixSvcTimeVariance) {
		masterId = chooseMatrixSvcMaster(matrixSvcs, r.matrixSvcTimeVariance)
	}

	if r.id != masterId {
		r.currentlyMaster = false
		return nil
	}

	if !r.currentlyMaster {
		err := r.store.UpdateMatrixSvcMaster(masterId)
		if err != nil {
			return svcError(err)
		}
		r.currentlyMaster = true
	}
	return nil
}

func (r *RelayFrontendSvc) UpdateLiveRouteMatrixOptimizer() error {
	routeMatrices, err := r.store.GetOptimizerMatrices()
	if err != nil {
		return svcError(err)
	}

	masterOptimizerID, err := r.store.GetOptimizerMaster()
	if err != nil && err.Error() != "optimizer master not found" {
		return svcError(err)
	}

	if !isMasterOptimizerValid(routeMatrices, masterOptimizerID, r.optimizerTimeVariance) {
		masterOptimizerID = chooseOptimizerMaster(routeMatrices, r.optimizerTimeVariance)
	}

	if r.currentMasterOptimizer != masterOptimizerID {
		err := r.store.UpdateOptimizerMaster(masterOptimizerID)
		if err != nil {
			return svcError(err)
		}
		r.currentMasterOptimizer = masterOptimizerID
	}

	err = r.updateLiveMatrix(routeMatrices, r.currentMasterOptimizer)
	if err != nil {
		return svcError(err)
	}

	return nil
}

func (r *RelayFrontendSvc) UpdateRelayBackendMaster() error {
	rbArr, err := r.store.GetRelayBackendLiveData(r.relayAddresses)
	if err != nil {
		return err
	}

	masterRB, err := r.store.GetRelayBackendMaster()
	if err != nil && err.Error() != "relay backend master not found" {
		return svcError(err)
	}

	if !isMasterRelayBackendValid(rbArr, masterRB.Address, 3*time.Second) {
		masterAddress, err := chooseRelayBackendMaster(rbArr, 3*time.Second)
		if err != nil {
			return err
		}
		r.currentMasterBackendAddress = masterAddress
		for _, relay := range rbArr {
			if relay.Address == r.currentMasterBackendAddress {
				err := r.store.SetRelayBackendMaster(relay)
				if err != nil {
					return svcError(err)
				}
			}
		}
	}
	return nil
}

func (r *RelayFrontendSvc) UpdateLiveRouteMatrixBackend(address, matrixType string) error {
	matrix, err := getHttpMatrix(address)
	if err != nil {
		return err
	}

	err = r.store.UpdateLiveMatrix(matrix, matrixType)
	if err != nil {
		return err
	}

	return nil
}

func (r *RelayFrontendSvc) CacheMatrix(matrixType string) error {

	matrix, err := r.store.GetLiveMatrix(matrixType)
	if err != nil {
		return err
	}
	r.matrixMux.Lock()
	switch matrixType {
	case storage.MatrixTypeNormal:
		r.routeMatrix = matrix
	case storage.MatrixTypeValve:
		r.routeMatrixValve = matrix
	}
	r.matrixMux.Unlock()

	return nil
}

func (r *RelayFrontendSvc) CleanUpDB() error {
	currentTime := time.Now().UTC()
	matrixSvcs, err := r.store.GetMatrixSvcs()
	if err != nil {
		return svcError(err)
	}

	for _, m := range matrixSvcs {
		if currentTime.Sub(m.UpdatedAt) > r.matrixSvcTimeVariance {
			err := r.store.DeleteMatrixSvc(m.ID)
			if err != nil {
				return err
			}
		}
	}

	opMatrices, err := r.store.GetOptimizerMatrices()
	if err != nil {
		return svcError(err)
	}

	for _, m := range opMatrices {
		if currentTime.Sub(m.CreatedAt) > r.optimizerTimeVariance {
			err := r.store.DeleteOptimizerMatrix(m.OptimizerID, m.Type)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RelayFrontendSvc) updateLiveMatrix(matrices []storage.Matrix, id uint64) error {
	found := false
	for _, m := range matrices {
		if m.OptimizerID == id {
			found = true
			err := r.store.UpdateLiveMatrix(m.Data, m.Type)
			if err != nil {
				return err
			}
		}
	}
	if !found {
		return fmt.Errorf("unable to find master matrix to update")
	}
	return nil
}

func isMasterMatrixSvcValid(matrices []storage.MatrixSvcData, id uint64, timeVariance time.Duration) bool {
	for _, m := range matrices {
		if m.ID == id {
			if time.Now().Sub(m.UpdatedAt) < timeVariance {
				return true
			}
			break
		}
	}
	return false
}

func isMasterOptimizerValid(matrices []storage.Matrix, id uint64, timeVariance time.Duration) bool {
	for _, m := range matrices {
		if m.OptimizerID == id {
			if time.Now().Sub(m.CreatedAt) < timeVariance {
				return true
			}
			break
		}
	}
	return false
}

func isMasterRelayBackendValid(rbData []storage.RelayBackendLiveData, masterAddress string, timeVariance time.Duration) bool {
	for _, rb := range rbData {
		if rb.Address == masterAddress {
			if time.Now().Sub(rb.UpdatedAt) < timeVariance {
				return true
			}
			break
		}
	}
	return false
}

func chooseMatrixSvcMaster(matrices []storage.MatrixSvcData, timeVariance time.Duration) uint64 {
	currentTime := time.Now().UTC()
	masterSvc := storage.NewMatrixSvcData(0, currentTime, currentTime)
	for _, m := range matrices {
		if currentTime.Sub(m.UpdatedAt) > timeVariance {
			continue
		}
		if m.CreatedAt.After(masterSvc.CreatedAt) {
			continue
		}
		if m.CreatedAt.Equal(masterSvc.CreatedAt) && m.ID > masterSvc.ID {
			continue
		}
		masterSvc = m
	}
	return masterSvc.ID
}

func chooseOptimizerMaster(matrices []storage.Matrix, timeVariance time.Duration) uint64 {
	currentTime := time.Now().UTC()
	masterOp := storage.NewMatrix(0, currentTime, currentTime, "", []byte{})
	for _, m := range matrices {
		if currentTime.Sub(m.CreatedAt) > timeVariance {
			continue
		}
		if m.OptimizerCreatedAt.After(masterOp.OptimizerCreatedAt) {
			continue
		}
		if m.OptimizerCreatedAt.Equal(masterOp.OptimizerCreatedAt) && m.OptimizerID > masterOp.OptimizerID {
			continue
		}
		masterOp = m
	}
	return masterOp.OptimizerID
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
	case storage.MatrixTypeNormal:
		addr = fmt.Sprintf("http:/%s/route_matrix", r.currentMasterBackendAddress)
	case storage.MatrixTypeValve:
		addr = fmt.Sprintf("http:/%s/route_matrix_valve", r.currentMasterBackendAddress)
	default:
		return "", errors.New("matrix type not supported")
	}
	return addr, nil
}

func (r *RelayFrontendSvc) GetMatrix() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(r.routeMatrix)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (r *RelayFrontendSvc) GetMatrixValve() func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		r.matrixMux.RLock()
		defer r.matrixMux.RUnlock()
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(r.routeMatrixValve)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
