package storage

import (
	"encoding/json"
	"fmt"
	"time"
)

type MatrixStore interface {
	SetRelayBackendLiveData(data RelayBackendLiveData) error
	GetRelayBackendLiveData() ([]RelayBackendLiveData, error)
}

func wrap(pre, err, post string) error {
	return fmt.Errorf("%s%s%s", pre, err, post)
}

type RelayBackendLiveData struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	InitAt    time.Time `json:"init_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRelayBackendLiveData(id, address string, InitAt, UpdatedAt time.Time) RelayBackendLiveData {
	rb := new(RelayBackendLiveData)
	rb.ID = id
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
