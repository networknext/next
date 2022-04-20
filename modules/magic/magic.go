package magic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/storage"
)

const (
	MagicServiceMetadataKey = "MagicServiceMetadata"

	MagicUpcomingKey = "magic_upcoming"
	MagicCurrentKey  = "magic_current"
	MagicPreviousKey = "magic_previous"

	MagicUpdateFailsafeTimeout = 60 * time.Second
	TimeVariance               = 5 * time.Second
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
	instanceMetadataTimeout      time.Duration
	instanceID                   string
	initAt                       time.Time
	redisPool                    *redis.Pool
}

func NewMagicService(instanceMetadataTimeout time.Duration,
	instanceID string,
	initAt time.Time,
	redisHostname string,
	redisPassword string,
	maxIdleConnections int,
	maxActiveConnections int,
) (*MagicService, error) {

	// Setup redis pool
	pool := storage.NewRedisPool(redisHostname, redisPassword, maxIdleConnections, maxActiveConnections)
	if err := storage.ValidateRedisPool(pool); err != nil {
		return &MagicService{}, fmt.Errorf("failed to validate redis pool: %w", err)
	}

	ms := &MagicService{
		instanceMetadataTimeout:      instanceMetadataTimeout,
		instanceID:                   instanceID,
		initAt:                       initAt,
		redisPool:                    pool,
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

	keys, err := redis.Strings(conn.Do("KEYS", MagicServiceMetadataKey+"*"))
	if err == redis.ErrNil {
		// No keys in redis including this instance's, return false out of safety
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("IsOldestInstance(): failed to get magic instance metadatas: %w", err)
	}

	metadataArr := make([]*MagicInstanceMetadata, len(keys))
	for i, key := range keys {
		metadataBin, err := redis.Bytes(conn.Do("GET", key))
		if err == redis.ErrNil {
			// Key does not exist
			continue
		} else if err != nil {
			return false, fmt.Errorf("IsOldestInstance(): failed to retrieve key %s from redis: %w", key, err)
		}

		// Unmarshal the metadata binary
		var metadata *MagicInstanceMetadata
		if err = json.Unmarshal(metadataBin, metadata); err != nil {
			return false, fmt.Errorf("IsOldestInstance(): failed to unmarshal magic metadata for key %s: %w", key, err)
		}

		metadataArr[i] = metadata
	}

	currentTime := time.Now().UTC()
	for _, metadata := range metadataArr {

		if currentTime.Sub(metadata.UpdatedAt) > TimeVariance {
			// Instance has not updated its metdata in the last TimeVariance seconds
			return false, nil
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
	Inserts this magic instance's metadata into redis.
*/
func (ms *MagicService) InsertInstanceMetadata() error {
	// Create the instance metadata
	metadata := MagicInstanceMetadata{
		InstanceID: ms.instanceID,
		InitAt:     ms.initAt,
		UpdatedAt:  time.Now().UTC(),
	}

	metadataBin, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("InsertInstanceMetadata(): failed to marshal magic metadata: %w", err)
	}

	conn := ms.redisPool.Get()
	defer conn.Close()

	// Insert metadata into redis
	key := fmt.Sprintf("%s-%s", MagicServiceMetadataKey, metadata.InstanceID)

	reply, err := conn.Do("SET", key, metadataBin, "PX", ms.instanceMetadataTimeout.Milliseconds())
	if reply != "OK" {
		return fmt.Errorf("InsertInstanceMetadata(): reply is not OK, instead got %s: %w", reply, err)
	}

	return nil
}

/*
	Updates the magic values in redis by moving "current" to
	"previous", moving "upcoming" to "current", and inserting
	a new "upcoming" value.

	Failsafe is in place not to touch the magic value if any
	of the magic values were updated in the last 60 seconds.

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
		return err
	}

	// If the existing upcoming and current values are empty, insert new values for everything
	if existingUpcomingMagic == (MagicValue{}) && existingCurrentMagic == (MagicValue{}) {
		if err = ms.SetMagicValue(MagicUpcomingKey, ms.GenerateMagicValue()); err != nil {
			return err
		}
		if err = ms.SetMagicValue(MagicCurrentKey, ms.GenerateMagicValue()); err != nil {
			return err
		}
		if err = ms.SetMagicValue(MagicPreviousKey, ms.GenerateMagicValue()); err != nil {
			return err
		}

		return nil
	}

	currentTime := time.Now().UTC()

	// Early out if either the upcoming or current magic value has been updated in the last MagicUpdateFailsafeTimeout seconds
	if currentTime.Sub(existingUpcomingMagic.UpdatedAt) < MagicUpdateFailsafeTimeout || currentTime.Sub(existingCurrentMagic.UpdatedAt) < MagicUpdateFailsafeTimeout {
		return nil
	}

	// Update the magic values while setting new updated at time
	newPreviousMagic := existingCurrentMagic
	newPreviousMagic.UpdatedAt = currentTime

	newCurrentMagic := existingUpcomingMagic
	newCurrentMagic.UpdatedAt = currentTime

	newUpcomingMagic := ms.GenerateMagicValue()

	if err = ms.SetMagicValue(MagicPreviousKey, newPreviousMagic); err != nil {
		return err
	}
	if err = ms.SetMagicValue(MagicCurrentKey, newCurrentMagic); err != nil {
		return err
	}
	if err = ms.SetMagicValue(MagicUpcomingKey, newUpcomingMagic); err != nil {
		return err
	}

	return nil
}

/*
	Gets a magic value for a given key.
*/
func (ms *MagicService) GetMagicValue(magicKey string) (MagicValue, error) {
	conn := ms.redisPool.Get()
	defer conn.Close()

	var existingMagic *MagicValue

	magicBin, err := redis.Bytes(conn.Do("GET", magicKey))
	if err != nil && err != redis.ErrNil {
		return MagicValue{}, fmt.Errorf("GetMagicValue(): failed to retrieve key %s from redis: %w", magicKey, err)
	} else if err == redis.ErrNil {
		// No magic value was present for the key
		return MagicValue{}, nil

	}

	// Unmarshal the existing magic
	if err = json.Unmarshal(magicBin, existingMagic); err != nil {
		return MagicValue{}, fmt.Errorf("GetMagicValue(): failed to unmarshal existing magic for key %s: %w", magicKey, err)
	}

	return *existingMagic, nil
}

/*
	Sets the magic value for a key.
*/
func (ms *MagicService) SetMagicValue(magicKey string, value MagicValue) error {
	magicBin, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("SetMagicValue(): failed to marshal magic value %+v: %w", value, err)
	}

	conn := ms.redisPool.Get()
	defer conn.Close()

	// Insert metadata into redis
	reply, err := conn.Do("SET", magicKey, magicBin)
	if reply != "OK" {
		return fmt.Errorf("SetMagicValue(): reply is not OK, instead got %s: %w", reply, err)
	}

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
