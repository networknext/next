package storage

import(
	"fmt"
	"time"


	"github.com/gomodule/redigo/redis"
)

const(
	hSet = "HSET"
	hGet = "HGET"
	hVals = "HVALS"
	hDel = "HDEL"
	set = "SET"
	setex = "SETEX"
	get = "GET"
	mGet = "MGET"

	optimizer = "OptimizerHash"
	masters = "MatrixMastersHash"
	matrixSvc = "MatrixServiceHash"
	svcMaster = "ServiceMaster"
	matrixMaster = "LiveMatrix"
	optimizerMaster = "OptimizerMaster"
	NormalMatrixType = "Normal"
	ValveMatrixType = "Valve"
)


type RedisMatrixStore struct{
	conn redis.Conn
	matrixTimeout time.Duration
}

func NewRedisMatrixStore(addr string, readTimeout, writeTimeout, matrixExpire time.Duration) (*RedisMatrixStore, error) {
	r := new(RedisMatrixStore)
	conn, err := redis.Dial("tcp",addr,
		redis.DialReadTimeout(readTimeout),
		redis.DialWriteTimeout(writeTimeout))
	if err != nil {
		return nil, err
	}
	r.conn = conn
	r.matrixTimeout = matrixExpire

	return r, nil
}

func (r *RedisMatrixStore) Close() error{
	return r.conn.Close()
}

func (r *RedisMatrixStore)GetLiveMatrix(matrixType string) ([]byte, error){

	data, err := redis.Bytes(r.conn.Do(get, matrixMaster+matrixType))
	if err == redis.ErrNil{
		return []byte{}, fmt.Errorf("matrix not found")
	}
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (r *RedisMatrixStore)UpdateLiveMatrix(matrixData []byte, matrixType string) error{
	_, err := r.conn.Do("SET", matrixMaster+matrixType, matrixData, "PX", r.matrixTimeout.Milliseconds())
	return err
}

func (r *RedisMatrixStore)GetOptimizerMatrices() ([]Matrix, error){
	dataArr, err := redis.ByteSlices(r.conn.Do(hVals, optimizer))
	if err == redis.ErrNil || len(dataArr) == 0 {
		return []Matrix{}, fmt.Errorf("optimizer matrices not found")
	}
	if err != nil{
		return []Matrix{}, err
	}

	matrices := make([]Matrix, len(dataArr))
	for i , v := range dataArr{
		matrix, err := MatrixFromJSON(v)
		if err != nil {
			return []Matrix{}, err
		}
		matrices[i] = matrix
	}

	return matrices,nil
}

func (r *RedisMatrixStore)UpdateOptimizerMatrix(matrix Matrix) error{
	jsonMatrix, err := MatrixToJSON(matrix)
	if err != nil {
		return err
	}

	_, err = r.conn.Do(hSet, optimizer, string(matrix.OptimizerID)+matrix.Type, jsonMatrix)
	return err
}

func (r *RedisMatrixStore)DeleteOptimizerMatrix(id uint64, matrixType string) error {
	_, err := r.conn.Do(hDel, optimizer, string(id)+matrixType)
	return err
}

func (r *RedisMatrixStore)GetMatrixSvcs() ([]MatrixSvcData, error){
	dataArr, err := redis.ByteSlices(r.conn.Do(hVals, matrixSvc))
	if err == redis.ErrNil || len(dataArr) == 0{
		return []MatrixSvcData{}, fmt.Errorf("matrix svc data not found")
	}
	if err != nil{
		return []MatrixSvcData{}, err
	}

	matrices := make([]MatrixSvcData, len(dataArr))
	for i , v := range dataArr{
		matrix, err := MatrixSvcFromJSON(v)
		if err != nil {
			return []MatrixSvcData{}, err
		}
		matrices[i] = matrix
	}

	return matrices,nil
}

func (r *RedisMatrixStore)UpdateMatrixSvc(matrixSvcData MatrixSvcData) error{
	jsonMatrixSvcData, err := MatrixSvcToJSON(matrixSvcData)
	if err != nil {
		return err
	}

	_, err = r.conn.Do(hSet, matrixSvc, matrixSvcData.ID, jsonMatrixSvcData )
	return err
}

func (r *RedisMatrixStore)DeleteMatrixSvc(id uint64) error{
	_, err := r.conn.Do(hDel, matrixSvc, id)
	return err
}

func (r *RedisMatrixStore)UpdateMatrixSvcMaster(id uint64) error{
	_, err := r.conn.Do(hSet, masters, svcMaster, id)
	return err
}

func (r *RedisMatrixStore)UpdateOptimizerMaster(id uint64) error{
	_, err := r.conn.Do(hSet, masters, optimizerMaster, id)
	return err
}

func (r *RedisMatrixStore)GetMatrixSvcMaster() (uint64, error){
	value, err := redis.Uint64(r.conn.Do(hGet, masters, svcMaster))
	if err == redis.ErrNil {
		return 0, fmt.Errorf("matrix svc master not found")
	}
	return value, err
}

func (r *RedisMatrixStore)GetOptimizerMaster() (uint64, error){
	value , err := redis.Uint64(r.conn.Do(hGet, masters, optimizerMaster))
	if err == redis.ErrNil {
		return value, fmt.Errorf("optimizer master not found")
	}

	return value, err
}

