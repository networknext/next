package storage

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	redisRelayStorePrefix = "relay:"
)

type RedisRelayStore struct{
	pool *redis.Pool
	relayTimeout time.Duration
}

func NewRedisRelayStore(addr string, readTimeout, writeTimeout, relayExpire time.Duration) (*RedisRelayStore, error) {
	r := new(RedisRelayStore)
	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",addr,
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout))
		},
	}
	r.pool = pool
	r.relayTimeout = relayExpire
	r.cleanupHook()
	return r, nil
}

func (r *RedisRelayStore)cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		r.pool.Close()
	}()
}

func (r *RedisRelayStore) Close() error {
	return r.pool.Close()
}

func (r *RedisRelayStore) Set(relayData RelayStoreData) error{
	data, err := RelayToJSON(relayData)
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	_,err = conn.Do("SET",r.key(relayData.ID),data,"EX", r.relayTimeout.Seconds())
	if err != nil{
		return fmt.Errorf("issue with db: %s", err.Error())
	}
	return nil
}

func (r *RedisRelayStore) ExpireReset(relayID uint64) error {
	conn := r.pool.Get()
	code, err := conn.Do("EXPIRE", r.key(relayID), r.relayTimeout.Seconds())
	if code != int64(1){
		return fmt.Errorf("expire not set code %v", code)
	}
	if err == redis.ErrNil{
		return fmt.Errorf("relay not found")
	}
	return err
}

func (r *RedisRelayStore) Get(relayID uint64) (*RelayStoreData, error){
	conn := r.pool.Get()
	data, err := redis.Bytes(conn.Do("GET",r.key(relayID)))
	if err == redis.ErrNil{
		return nil, fmt.Errorf("unable to find relay data")
	}
	if err != nil{
		return nil, fmt.Errorf("issue with db: %s", err.Error())
	}

	return RelayFromJSON(data)
}

func (r *RedisRelayStore) GetAll() ([]*RelayStoreData, error){
	conn := r.pool.Get()
	keys, err := redis.Strings(conn.Do("KEYS", redisRelayStorePrefix+"*"))
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, k := range keys {
		args = append(args, k)
	}
	
	dataArr, err := redis.ByteSlices(conn.Do("MGET", args...))
	if err == redis.ErrNil{
		return nil, fmt.Errorf("unable to get relay data")
	}

	rsdArr := make([]*RelayStoreData,len(dataArr))
	for i, data := range dataArr{
		if data == nil{return nil, fmt.Errorf("no data")}
		rsdArr[i], err = RelayFromJSON(data)
		if err != nil {
			return nil, err
		}
	}
	
	return rsdArr, nil
}

func (r *RedisRelayStore) Delete(relayID uint64) error{
	conn := r.pool.Get()
	_, err := conn.Do("DEL", r.key(relayID))
	return err
}

func (r *RedisRelayStore) key(relayID uint64) string{
	return fmt.Sprintf("%s%v",redisRelayStorePrefix,relayID)
}