package magic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
)

const (
	MagicServiceMetadataKey = "MagicServiceMetadata"

	MagicUpcomingKey = "magic_upcoming"
	MagicCurrentKey  = "magic_current"
	MagicPreviousKey = "magic_previous"

	MagicUpdateFailsafeTimeout = 55 * time.Second           // todo: this is hardcoded
	TimeVariance               = 5 * time.Second            // todo: this is hardcoded
)

type MagicValue struct {
	UpdatedAt  time.Time `json:"updated_at"`
	MagicBytes [8]byte
}

type MagicInstanceMetadata struct {
	InstanceID string    `json:"instance_id"`
	InitAt     time.Time `json:"init_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MagicService struct {
	instanceMetadataTimeout time.Duration
	instanceID              string
	initAt                  time.Time
	redisPool               *redis.Pool
	magicMetrics            *metrics.MagicMetrics
}

func NewMagicService(instanceMetadataTimeout time.Duration,
	instanceID string,
	initAt time.Time,
	redisHostname string,
	redisPassword string,
	maxIdleConnections int,
	maxActiveConnections int,
	magicMetrics *metrics.MagicMetrics,
) (*MagicService, error) {

	// Setup redis pool
	pool := storage.NewRedisPool(redisHostname, redisPassword, maxIdleConnections, maxActiveConnections)
	if err := storage.ValidateRedisPool(pool); err != nil {
		return nil, fmt.Errorf("failed to validate redis pool: %w", err)
	}

	ms := &MagicService{
		instanceMetadataTimeout: instanceMetadataTimeout,
		instanceID:              instanceID,
		initAt:                  initAt,
		redisPool:               pool,
		magicMetrics:            magicMetrics,
	}

	return ms, nil
}

/*
Determines if this magic instance is the oldest instance.

The oldest instance has to have updated its metdata in the
last TimeVariance seconds and has to have the earliest init
time. In the event of a tie, the larger instance ID is
considered older.
*/
func (ms *MagicService) IsOldestInstance() (bool, error) {
	conn := ms.redisPool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", MagicServiceMetadataKey+"-*"))
	if err != nil && err != redis.ErrNil {
		ms.magicMetrics.ErrorMetrics.ReadFromRedisFailure.Add(1)
		return false, fmt.Errorf("IsOldestInstance(): failed to get magic instance metadatas: %w", err)
	}

	if len(keys) == 0 {
		// No keys in redis including this instance's, return false out of safety
		return false, nil
	}

	metadataArr := make([]MagicInstanceMetadata, len(keys))
	for i, key := range keys {
		metadataBin, err := redis.Bytes(conn.Do("GET", key))
		if err == redis.ErrNil {
			// Key does not exist
			continue
		} else if err != nil {
			ms.magicMetrics.ErrorMetrics.ReadFromRedisFailure.Add(1)
			return false, fmt.Errorf("IsOldestInstance(): failed to retrieve key %s from redis: %w", key, err)
		}

		// Unmarshal the metadata binary
		var metadata MagicInstanceMetadata
		if err = json.Unmarshal(metadataBin, &metadata); err != nil {
			ms.magicMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return false, fmt.Errorf("IsOldestInstance(): failed to unmarshal magic metadata for key %s: %w", key, err)
		}

		metadataArr[i] = metadata
	}

	currentTime := time.Now().UTC()
	for _, metadata := range metadataArr {

		if currentTime.Sub(metadata.UpdatedAt) > TimeVariance {
			// Instance has not updated its metadata in the last TimeVariance seconds, ignore
			continue
		}

		if ms.initAt.After(metadata.InitAt) {
			// Instance init time is before this init time
			return false, nil
		}

		if ms.initAt.Equal(metadata.InitAt) && ms.instanceID < metadata.InstanceID {
			// In a tie, instance metadata id is larger than this metadata id
			return false, nil
		}
	}

	return true, nil
}

/*
Creates this magic instance's metadata for insertion into redis.
*/
func (ms *MagicService) CreateInstanceMetadata() MagicInstanceMetadata {
	return MagicInstanceMetadata{
		InstanceID: ms.instanceID,
		InitAt:     ms.initAt,
		UpdatedAt:  time.Now().UTC(),
	}
}

/*
Inserts this magic instance's metadata into redis.
*/
func (ms *MagicService) InsertInstanceMetadata(metadata MagicInstanceMetadata) error {
	metadataBin, err := json.Marshal(metadata)
	if err != nil {
		ms.magicMetrics.ErrorMetrics.MarshalFailure.Add(1)
		ms.magicMetrics.ErrorMetrics.InsertInstanceMetadataFailure.Add(1)
		return fmt.Errorf("InsertInstanceMetadata(): failed to marshal magic metadata: %w", err)
	}

	conn := ms.redisPool.Get()
	defer conn.Close()

	// Insert metadata into redis
	key := fmt.Sprintf("%s-%s", MagicServiceMetadataKey, metadata.InstanceID)

	reply, err := conn.Do("SET", key, metadataBin, "PX", ms.instanceMetadataTimeout.Milliseconds())
	if reply != "OK" {
		ms.magicMetrics.ErrorMetrics.InsertInstanceMetadataFailure.Add(1)
		return fmt.Errorf("InsertInstanceMetadata(): reply is not OK, instead got %s: %w", reply, err)
	}

	ms.magicMetrics.InsertInstanceMetadataSuccess.Add(1)
	return nil
}

/*
Updates the magic values in redis by moving "current" to
"previous", moving "upcoming" to "current", and inserting
a new "upcoming" value.

Failsafe is in place not to touch the magic value if any
of the magic values were updated very recently.

If the upcoming and previous magic values are missing,
this function will repopulate all magic values.
*/
func (ms *MagicService) UpdateMagicValues() error {

	existingUpcomingMagic, err := ms.GetMagicValue(MagicUpcomingKey)
	if err != nil {
		return err
	}

	existingCurrentMagic, err := ms.GetMagicValue(MagicCurrentKey)
	if err != nil {
		ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
		return err
	}

	// If the existing upcoming and current values are empty, insert new values for everything
	if existingUpcomingMagic == (MagicValue{}) && existingCurrentMagic == (MagicValue{}) {
		if err = ms.SetMagicValue(MagicUpcomingKey, ms.GenerateMagicValue()); err != nil {
			ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
			return err
		}
		if err = ms.SetMagicValue(MagicCurrentKey, ms.GenerateMagicValue()); err != nil {
			ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
			return err
		}
		if err = ms.SetMagicValue(MagicPreviousKey, ms.GenerateMagicValue()); err != nil {
			ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
			return err
		}

		ms.magicMetrics.UpdateMagicValuesSuccess.Add(1)
		return nil
	}

	currentTime := time.Now().UTC()

	// Early out if either the upcoming or current magic value has been updated in the last MagicUpdateFailsafeTimeout seconds
	if currentTime.Sub(existingUpcomingMagic.UpdatedAt) < MagicUpdateFailsafeTimeout || currentTime.Sub(existingCurrentMagic.UpdatedAt) < MagicUpdateFailsafeTimeout {
		core.Debug("magic values were updated in the last %.2f seconds, skipping update", MagicUpdateFailsafeTimeout.Seconds())
		return nil
	}

	// Update the magic values while setting new updated at time
	newPreviousMagic := existingCurrentMagic
	newPreviousMagic.UpdatedAt = currentTime

	newCurrentMagic := existingUpcomingMagic
	newCurrentMagic.UpdatedAt = currentTime

	newUpcomingMagic := ms.GenerateMagicValue()

	if err = ms.SetMagicValue(MagicPreviousKey, newPreviousMagic); err != nil {
		ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
		return err
	}
	if err = ms.SetMagicValue(MagicCurrentKey, newCurrentMagic); err != nil {
		ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
		return err
	}
	if err = ms.SetMagicValue(MagicUpcomingKey, newUpcomingMagic); err != nil {
		ms.magicMetrics.ErrorMetrics.UpdateMagicValuesFailure.Add(1)
		return err
	}

	ms.magicMetrics.UpdateMagicValuesSuccess.Add(1)
	return nil
}

/*
Gets a magic value for a given key.
*/
func (ms *MagicService) GetMagicValue(magicKey string) (MagicValue, error) {
	conn := ms.redisPool.Get()
	defer conn.Close()

	var existingMagic MagicValue

	magicBin, err := redis.Bytes(conn.Do("GET", magicKey))
	if err != nil && err != redis.ErrNil {
		ms.magicMetrics.ErrorMetrics.GetMagicValueFailure.Add(1)
		return MagicValue{}, fmt.Errorf("GetMagicValue(): failed to retrieve key %s from redis: %w", magicKey, err)
	} else if err == redis.ErrNil {
		// No magic value was present for the key
		ms.magicMetrics.GetMagicValueSuccess.Add(1)
		return MagicValue{}, nil
	}

	// Unmarshal the existing magic
	if err = json.Unmarshal(magicBin, &existingMagic); err != nil {
		ms.magicMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
		return MagicValue{}, fmt.Errorf("GetMagicValue(): failed to unmarshal existing magic for key %s: %w", magicKey, err)
	}

	ms.magicMetrics.GetMagicValueSuccess.Add(1)
	return existingMagic, nil
}

/*
Sets the magic value for a key.
*/
func (ms *MagicService) SetMagicValue(magicKey string, value MagicValue) error {
	magicBin, err := json.Marshal(value)
	if err != nil {
		ms.magicMetrics.ErrorMetrics.MarshalFailure.Add(1)
		return fmt.Errorf("SetMagicValue(): failed to marshal magic value %+v: %w", value, err)
	}

	conn := ms.redisPool.Get()
	defer conn.Close()

	// Insert metadata into redis
	reply, err := conn.Do("SET", magicKey, magicBin)
	if reply != "OK" {
		ms.magicMetrics.ErrorMetrics.SetMagicValueFailure.Add(1)
		return fmt.Errorf("SetMagicValue(): reply is not OK, instead got %s: %w", reply, err)
	}

	ms.magicMetrics.SetMagicValueSuccess.Add(1)
	return nil
}

/*
Generates a new magic value.
*/
func (ms *MagicService) GenerateMagicValue() MagicValue {
	var magic [8]byte
	core.RandomBytes(magic[:])

	return MagicValue{
		UpdatedAt:  time.Now().UTC(),
		MagicBytes: magic,
	}
}
