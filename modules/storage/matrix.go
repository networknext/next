package storage

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	MatrixTypeNormal = "Normal"
	MatrixTypeValve  = "Valve"
)

//go:generate moq -out matrix_test_mocks.go . MatrixStore
type MatrixStore interface {
	//Live Matrix for server backend
	GetLiveMatrix(matrixType string) ([]byte, error)
	UpdateLiveMatrix(matrixData []byte, matrixType string) error

	//optimizer matrices
	GetOptimizerMatrices() ([]Matrix, error)
	UpdateOptimizerMatrix(matrix Matrix) error
	DeleteOptimizerMatrix(id uint64, matrixType string) error

	//matrix svc data
	GetMatrixSvcs() ([]MatrixSvcData, error)
	UpdateMatrixSvc(matrixSvcData MatrixSvcData) error
	DeleteMatrixSvc(id uint64) error

	//matrix master
	GetMatrixSvcMaster() (uint64, error)
	UpdateMatrixSvcMaster(id uint64) error

	//optimizer master
	GetOptimizerMaster() (uint64, error)
	UpdateOptimizerMaster(id uint64) error

	//relay backend
	SetRelayBackendLiveData(data RelayBackendLiveData) error
	GetRelayBackendLiveData(address []string) ([]RelayBackendLiveData, error)
	SetRelayBackendMaster(RelayBackendLiveData) error
	GetRelayBackendMaster() (RelayBackendLiveData, error)
}

type Matrix struct {
	OptimizerID        uint64    `json:"optimizer_id"`
	OptimizerCreatedAt time.Time `json:"optimizer_created_at"`
	CreatedAt          time.Time `json:"created_at"`
	Type               string    `json:"type"`
	Data               []byte    `json:"data"`
}

func wrap(pre, err, post string) error {
	return fmt.Errorf("%s%s%s", pre, err, post)
}

func NewMatrix(optimizerID uint64, optimizerCreated, createdAt time.Time, matrixType string, data []byte) Matrix {
	m := new(Matrix)
	m.OptimizerID = optimizerID
	m.OptimizerCreatedAt = optimizerCreated
	m.CreatedAt = createdAt
	m.Type = matrixType
	m.Data = data
	return *m
}

func MatrixToJSON(matrix Matrix) ([]byte, error) {
	return json.Marshal(matrix)
}

func MatrixFromJSON(data []byte) (Matrix, error) {
	m := new(Matrix)
	err := json.Unmarshal(data, m)
	return *m, err
}

type MatrixSvcData struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"Updated_at"`
}

func NewMatrixSvcData(id uint64, createdAt, updatedAt time.Time) MatrixSvcData {
	m := new(MatrixSvcData)
	m.ID = id
	m.CreatedAt = createdAt
	m.UpdatedAt = updatedAt
	return *m
}

func MatrixSvcToJSON(matrixSvc MatrixSvcData) ([]byte, error) {
	return json.Marshal(matrixSvc)
}

func MatrixSvcFromJSON(data []byte) (MatrixSvcData, error) {
	m := new(MatrixSvcData)
	err := json.Unmarshal(data, m)
	if err != nil {
		return *m, wrap("unable to unmarshal", err.Error(), "")
	}
	return *m, nil
}

type RelayBackendLiveData struct {
	Id        string    `json:"id"`
	Address   string    `json:"address"`
	InitAt    time.Time `json:"init_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRelayBackendLiveData(id, address string, InitAt, UpdatedAt time.Time) RelayBackendLiveData {
	rb := new(RelayBackendLiveData)
	rb.Id = id
	rb.Address = address
	rb.InitAt = InitAt
	rb.UpdatedAt = UpdatedAt

	return *rb
}

func RelayBackendLiveDataToJSON(data RelayBackendLiveData) ([]byte, error) {
	return json.Marshal(data)
}

func RelayBackendLiveDataFromJson(data []byte) (RelayBackendLiveData, error) {
	r := new(RelayBackendLiveData)
	err := json.Unmarshal(data, r)
	if err != nil {
		err = wrap("unable to unmarshal relay backend live data", err.Error(), "")
	}
	return *r, err
}
