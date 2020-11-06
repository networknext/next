package route_matrix

import (
	"fmt"
	"github.com/networknext/backend/storage"
	"math/rand"
	"time"
)

type RouteMatrixSvc struct {
	id                     uint64
	store                  storage.MatrixStore
	createdAt              time.Time
	currentlyMaster        bool
	currentMasterOptimizer uint64
	matrixSvcTimeVariance  time.Duration
	optimizerTimeVariance  time.Duration
}

func New(store storage.MatrixStore, matrixSvcTimeVariance, optimizerTimeVariance int64) (*RouteMatrixSvc, error) {
	rand.Seed(time.Now().UnixNano())

	r := new(RouteMatrixSvc)
	r.id = rand.Uint64()
	r.store = store
	r.createdAt = time.Now().UTC()
	r.currentlyMaster = false
	r.matrixSvcTimeVariance = time.Duration(matrixSvcTimeVariance) * time.Millisecond
	r.optimizerTimeVariance = time.Duration(optimizerTimeVariance) * time.Millisecond

	return r, nil
}

func svcError(err error) error {
	return fmt.Errorf("route matrix svc: %s", err.Error())
}

func (r *RouteMatrixSvc) UpdateSvcDB() error {
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

func (r *RouteMatrixSvc) AmMaster() bool {
	return r.currentlyMaster
}

func (r *RouteMatrixSvc) DetermineMaster() error {
	matrixSvcs, err := r.store.GetMatrixSvcs()
	if err != nil {
		return svcError(err)
	}

	masterId, err := r.store.GetMatrixSvcMaster()
	if err != nil {
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

func (r *RouteMatrixSvc) UpdateLiveRouteMatrix() error {
	routeMatrices, err := r.store.GetOptimizerMatrices()
	if err != nil {
		return svcError(err)
	}

	masterOptimizerID, err := r.store.GetOptimizerMaster()
	if err != nil {
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

func (r *RouteMatrixSvc) CleanUpDB() error{
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

func (r *RouteMatrixSvc) updateLiveMatrix(matrices []storage.Matrix, id uint64) error {
	for _, m := range matrices {
		if m.OptimizerID == id {
			return r.store.UpdateLiveMatrix(m.Data, m.Type)
		}
	}
	return fmt.Errorf("unable to find master matrix to update")
}

func isMasterMatrixSvcValid(matrices []storage.MatrixSvcData, id uint64, timeVariance time.Duration) bool {
	found := false
	for _, m := range matrices {
		if m.ID == id {
			found = true
			if time.Now().Sub(m.UpdatedAt) > timeVariance {
				return false
			}
		}
	}

	if !found {
		return false
	}
	return true
}

func isMasterOptimizerValid(matrices []storage.Matrix, id uint64, timeVariance time.Duration) bool {
	found := false
	for _, m := range matrices {
		if m.OptimizerID == id {
			found = true
			if time.Now().Sub(m.CreatedAt) > timeVariance {
				return false
			}
		}
	}

	if !found {
		return false
	}
	return true
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
		if m.CreatedAt.Equal(masterSvc.CreatedAt) && m.ID < masterSvc.ID {
			continue
		}
		masterSvc = m
	}
	return masterSvc.ID
}

func chooseOptimizerMaster(matrices []storage.Matrix, timeVariance time.Duration) uint64 {
	currentTime := time.Now().UTC()
	masterOp := storage.NewMatrix(0, currentTime, currentTime,"", []byte{})
	for _, m := range matrices {
		if currentTime.Sub(m.CreatedAt) > timeVariance {
			continue
		}
		if m.CreatedAt.After(masterOp.CreatedAt) {
			continue
		}
		if m.CreatedAt.Equal(masterOp.CreatedAt) && m.OptimizerID < masterOp.OptimizerID {
			continue
		}
		masterOp = m
	}
	return masterOp.OptimizerID
}
