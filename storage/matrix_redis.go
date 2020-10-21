package storage

type RedisMatrixStore struct{
	
}

func NewRedisMatrixStore() RedisMatrixStore {
	return RedisMatrixStore{}
}

func (s RedisMatrixStore)GetMatrix() ([]byte, error){
	return []byte{}, nil
}

func (s RedisMatrixStore)GetMatrices() ([]Matrix, uint64, error){
	return []Matrix{},0,nil
}

func (s RedisMatrixStore)GetMatrixSvcs() ([]MatrixSvcData, uint64, error){
	return []MatrixSvcData{},0,nil
}

func (s RedisMatrixStore)StoreMatrix(matrix Matrix) error{
	return nil
}

func (s RedisMatrixStore)UpdateLiveMatrix(matrixData []byte) error{
	return nil
}

func (s RedisMatrixStore)UpdateMatrixSvc(matrixSvcData MatrixSvcData) error{
	return nil
}

func (s RedisMatrixStore)UpdateMatrixSvcMaster(id uint64) error{
	return nil
}

func (s RedisMatrixStore)UpdateOptimizerMaster(id uint64) error{
	return nil
}


