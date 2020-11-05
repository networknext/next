package storage

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

const (
	redisRelayStorePrefix = "relay:"
)

type RedisRelayStore struct{
	conn redis.Conn
	relayTimeout time.Duration
}

func NewRedisRelayStore(addr string, readTimeout, writeTimeout, relayExpire time.Duration) (*RedisRelayStore, error) {
	r := new(RedisRelayStore)
	conn, err := redis.Dial("tcp",addr,
		redis.DialReadTimeout(readTimeout),
		redis.DialWriteTimeout(writeTimeout))
	if err != nil {
		return nil, err
	}
	r.conn = conn
	r.relayTimeout = relayExpire

	return r, nil
}

func (r *RedisRelayStore) Close() error {
	return r.conn.Close()
}

func (r *RedisRelayStore) Set(relayData RelayStoreData) error{
	data, err := RelayToJSON(relayData)
	if err != nil {
		return err
	}

	_,err = r.conn.Do("SET",r.key(relayData.ID),data,"EX", r.relayTimeout.Seconds())
	if err != nil{
		return fmt.Errorf("issue with db: %s", err.Error())
	}
	return nil
}

func (r *RedisRelayStore) ExpireReset(relayID uint64) error {
	code, err := r.conn.Do("EXPIRE", r.key(relayID), r.relayTimeout)
	if code != int64(1){
		return fmt.Errorf("expire not set code %v", code)
	}
	if err == redis.ErrNil{
		return fmt.Errorf("relay not found")
	}
	return err
}

func (r *RedisRelayStore) Get(relayID uint64) (*RelayStoreData, error){
	data, err := redis.Bytes(r.conn.Do("GET",r.key(relayID)))
	if err == redis.ErrNil{
		return nil, fmt.Errorf("unable to find relay data")
	}
	if err != nil{
		return nil, fmt.Errorf("issue with db: %s", err.Error())
	}

	return RelayFromJSON(data)
}

func (r *RedisRelayStore) GetAll() ([]*RelayStoreData, error){
	keys, err := redis.Strings(r.conn.Do("KEYS", redisRelayStorePrefix+"*"))
	if err != nil {
		return nil, err
	}

	var args []interface{}
	for _, k := range keys {
		args = append(args, k)
	}
	
	dataArr, err := redis.ByteSlices(r.conn.Do("MGET", args...))
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
	_, err := r.conn.Do("DEL", r.key(relayID))
	return err
}

func (r *RedisRelayStore) key(relayID uint64) string{
	return fmt.Sprintf("%s%v",redisRelayStorePrefix,relayID)
}