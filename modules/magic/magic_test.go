package magic_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/magic"
	"github.com/networknext/backend/modules/metrics"
)

func getTestMagicInstanceMetadata(initTime time.Time) magic.MagicInstanceMetadata {
	return magic.MagicInstanceMetadata{
		InstanceID: backend.GenerateRandomStringSequence(8),
		InitAt:     initTime,
		UpdatedAt:  time.Now().UTC(),
	}
}

func getTestMagicValue(updatedTime time.Time) magic.MagicValue {
	var val [8]byte
	core.RandomBytes(val[:])

	return magic.MagicValue{
		UpdatedAt:  updatedTime,
		MagicBytes: val,
	}
}

func TestNewMagicService_RedisPoolMissingHostname(t *testing.T) {
	t.Parallel()

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), "", "", 5, 5, metrics.EmptyMagicMetrics)
	assert.Nil(t, ms)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "failed to validate redis pool")
}

func TestNewMagicService_Success(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)
}

func TestCreateInstanceMetadata(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	instanceID := "test-instance-id"
	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Second, instanceID, initTime, redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	metadata := ms.CreateInstanceMetadata()

	assert.Equal(t, instanceID, metadata.InstanceID)
	assert.Equal(t, initTime, metadata.InitAt)
	assert.True(t, metadata.UpdatedAt.After(metadata.InitAt))
}

func TestInsertInstanceMetadata_RedisWriteFailure(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	redisPool.Close()

	err = ms.InsertInstanceMetadata(getTestMagicInstanceMetadata(time.Now()))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "InsertInstanceMetadata(): reply is not OK, instead got")
}

func TestInsertInstanceMetadata_Success(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestInsertInstanceMetadata_Success", "", "")
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now().UTC(), redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	err = ms.InsertInstanceMetadata(getTestMagicInstanceMetadata(time.Now().UTC()))
	assert.Equal(t, float64(1), msMetrics.InsertInstanceMetadataSuccess.Value())
	assert.NoError(t, err)
}

func TestIsOldestInstance_RedisReadFailure(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	redisPool.Close()

	isOldest, err := ms.IsOldestInstance()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "IsOldestInstance(): failed to get magic instance metadatas")
	assert.False(t, isOldest)
}

func TestIsOldestInstance_NoInstanceMetadata(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	isOldest, err := ms.IsOldestInstance()
	assert.NoError(t, err)
	assert.False(t, isOldest)
}

func TestIsOldestInstance_SameInstance_Success(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id", time.Now().UTC(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	metadata := ms.CreateInstanceMetadata()
	err = ms.InsertInstanceMetadata(metadata)
	assert.NoError(t, err)

	isOldest, err := ms.IsOldestInstance()
	assert.NoError(t, err)
	assert.True(t, isOldest)
}

func TestIsOldestInstance_OtherInstance_TimeVarianceCheck(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id", initTime, redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	otherMetadata := getTestMagicInstanceMetadata(initTime)
	otherMetadata.UpdatedAt = initTime.Add(-10 * time.Second)
	err = ms.InsertInstanceMetadata(otherMetadata)
	assert.NoError(t, err)

	metadata := ms.CreateInstanceMetadata()
	err = ms.InsertInstanceMetadata(metadata)
	assert.NoError(t, err)

	isOldest, err := ms.IsOldestInstance()
	assert.NoError(t, err)
	assert.True(t, isOldest)
}

func TestIsOldestInstance_OtherInstance_InitAtCheck(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id", initTime, redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	otherMetadata := getTestMagicInstanceMetadata(initTime.Add(time.Second))
	err = ms.InsertInstanceMetadata(otherMetadata)
	assert.NoError(t, err)

	metadata := ms.CreateInstanceMetadata()
	err = ms.InsertInstanceMetadata(metadata)
	assert.NoError(t, err)

	isOldest, err := ms.IsOldestInstance()
	assert.NoError(t, err)
	assert.True(t, isOldest)
}

func TestIsOldestInstance_OtherInstance_InstanceIDCheck(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id1", initTime, redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	otherMetadata := getTestMagicInstanceMetadata(initTime)
	otherMetadata.InstanceID = "test-instance-id1"
	err = ms.InsertInstanceMetadata(otherMetadata)
	assert.NoError(t, err)

	metadata := ms.CreateInstanceMetadata()
	err = ms.InsertInstanceMetadata(metadata)
	assert.NoError(t, err)

	isOldest, err := ms.IsOldestInstance()
	assert.NoError(t, err)
	assert.True(t, isOldest)
}

func TestSetMagicValue_RedisWriteFailure(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id", initTime, redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	redisPool.Close()

	err = ms.SetMagicValue("magic-key", getTestMagicValue(initTime))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SetMagicValue(): reply is not OK, instead got")
}

func TestSetMagicValue_Success(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestSetMagicValue_Success", "", "")
	assert.NoError(t, err)

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Hour, "test-instance-id", initTime, redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	err = ms.SetMagicValue("magic-key", getTestMagicValue(initTime))
	assert.NoError(t, err)
	assert.Equal(t, float64(1), msMetrics.SetMagicValueSuccess.Value())
}

func TestGetMagicValue_RedisReadFailure(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	redisPool.Close()

	magicKey := "some-magic-key"
	retMagicVal, err := ms.GetMagicValue(magicKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("GetMagicValue(): failed to retrieve key %s from redis", magicKey))
	assert.Equal(t, magic.MagicValue{}, retMagicVal)
}

func TestGetMagicValue_NoMagicValues(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestGetMagicValue_NoMagicValues", "", "")
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	retMagicVal, err := ms.GetMagicValue("some-magic-key")
	assert.NoError(t, err)
	assert.Equal(t, magic.MagicValue{}, retMagicVal)
	assert.Equal(t, float64(1), msMetrics.GetMagicValueSuccess.Value())
}

func TestGetMagicValue_Success(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestGetMagicValue_Success", "", "")
	assert.NoError(t, err)

	initTime := time.Now().UTC()

	ms, err := magic.NewMagicService(time.Second, "", initTime, redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	testMagicValue := getTestMagicValue(initTime)

	err = ms.SetMagicValue("magic-key", testMagicValue)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), msMetrics.SetMagicValueSuccess.Value())

	retMagicVal, err := ms.GetMagicValue("magic-key")
	assert.NoError(t, err)
	assert.Equal(t, testMagicValue, retMagicVal)
	assert.Equal(t, float64(1), msMetrics.GetMagicValueSuccess.Value())
}

func TestGenerateMagicValue(t *testing.T) {
	t.Parallel()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, metrics.EmptyMagicMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	emptyMagic := magic.MagicValue{}
	prevMagic := getTestMagicValue(time.Now().UTC())
	for i := 0; i < 100; i++ {
		newMagic := ms.GenerateMagicValue()
		assert.NotEqual(t, emptyMagic, newMagic)
		assert.NotEqual(t, prevMagic, newMagic)
		prevMagic = newMagic
	}
}

func TestUpdateMagicValues_RedisEmpty(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestUpdateMagicValues_RedisEmpty", "", "")
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	err = ms.UpdateMagicValues()
	assert.NoError(t, err)
	assert.Equal(t, float64(1), msMetrics.UpdateMagicValuesSuccess.Value())

	upcoming, err := ms.GetMagicValue(magic.MagicUpcomingKey)
	assert.NoError(t, err)
	assert.NotEqual(t, magic.MagicValue{}, upcoming)

	current, err := ms.GetMagicValue(magic.MagicCurrentKey)
	assert.NoError(t, err)
	assert.NotEqual(t, magic.MagicValue{}, current)

	previous, err := ms.GetMagicValue(magic.MagicPreviousKey)
	assert.NoError(t, err)
	assert.NotEqual(t, magic.MagicValue{}, previous)
}

func TestUpdateMagicValues_FailsafeTimeoutCheck(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestUpdateMagicValues_FailsafeTimeoutCheck", "", "")
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	testMagic := getTestMagicValue(time.Now().UTC())

	err = ms.SetMagicValue(magic.MagicUpcomingKey, testMagic)
	assert.NoError(t, err)

	err = ms.SetMagicValue(magic.MagicCurrentKey, testMagic)
	assert.NoError(t, err)

	err = ms.SetMagicValue(magic.MagicPreviousKey, testMagic)
	assert.NoError(t, err)

	err = ms.UpdateMagicValues()
	assert.NoError(t, err)
	assert.Zero(t, msMetrics.UpdateMagicValuesSuccess.Value())
}

func TestUpdateMagicValues_Success(t *testing.T) {
	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	msMetrics, err := metrics.NewMagicBackendMetrics(context.Background(), &metrics.LocalHandler{}, "TestUpdateMagicValues_Success", "", "")
	assert.NoError(t, err)

	ms, err := magic.NewMagicService(time.Second, "", time.Now(), redisPool.Addr(), "", 5, 5, msMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, ms)

	magicUpdateTime := time.Now().UTC().Add(-61 * time.Second)

	testMagicUpcoming := getTestMagicValue(magicUpdateTime)
	testMagicCurrent := getTestMagicValue(magicUpdateTime)
	testMagicPrevious := getTestMagicValue(magicUpdateTime)

	err = ms.SetMagicValue(magic.MagicUpcomingKey, testMagicUpcoming)
	assert.NoError(t, err)

	err = ms.SetMagicValue(magic.MagicCurrentKey, testMagicCurrent)
	assert.NoError(t, err)

	err = ms.SetMagicValue(magic.MagicPreviousKey, testMagicPrevious)
	assert.NoError(t, err)

	timeBeforeUpdate := time.Now().UTC()

	err = ms.UpdateMagicValues()
	assert.NoError(t, err)
	assert.Equal(t, float64(1), msMetrics.UpdateMagicValuesSuccess.Value())

	newMagicUpcoming, err := ms.GetMagicValue(magic.MagicUpcomingKey)
	assert.NoError(t, err)
	newMagicCurrent, err := ms.GetMagicValue(magic.MagicCurrentKey)
	assert.NoError(t, err)
	newMagicPrevious, err := ms.GetMagicValue(magic.MagicPreviousKey)
	assert.NoError(t, err)

	assert.NotEqual(t, magic.MagicValue{}, newMagicUpcoming)
	assert.NotEqual(t, magic.MagicValue{}, newMagicCurrent)
	assert.NotEqual(t, magic.MagicValue{}, newMagicPrevious)

	assert.Equal(t, testMagicCurrent.MagicBytes, newMagicPrevious.MagicBytes)
	assert.Equal(t, testMagicUpcoming.MagicBytes, newMagicCurrent.MagicBytes)
	assert.NotEqual(t, testMagicUpcoming.MagicBytes, newMagicUpcoming.MagicBytes)
	assert.NotEqual(t, testMagicCurrent.MagicBytes, newMagicUpcoming.MagicBytes)
	assert.NotEqual(t, testMagicPrevious.MagicBytes, newMagicUpcoming.MagicBytes)

	assert.True(t, newMagicUpcoming.UpdatedAt.After(timeBeforeUpdate))
	assert.True(t, newMagicCurrent.UpdatedAt.After(timeBeforeUpdate))
	assert.True(t, newMagicPrevious.UpdatedAt.After(timeBeforeUpdate))
}
