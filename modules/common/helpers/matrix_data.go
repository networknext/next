package helpers

import "sync"

type MatrixData struct {
	locker sync.RWMutex
	data   []byte
	binary []byte
}

func (m *MatrixData) GetMatrix() []byte {
	m.locker.RLock()
	defer m.locker.RUnlock()
	return m.data
}

func (m *MatrixData) GetMatrixBinary() []byte {
	m.locker.RLock()
	defer m.locker.RUnlock()
	return m.binary
}

func (m *MatrixData) SetMatrix(matrix []byte) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.data = matrix
}

func (m *MatrixData) SetMatrixBinary(matrix []byte) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.binary = matrix
}

func (m *MatrixData) GetMatrixDataSize() int {
	m.locker.RLock()
	defer m.locker.RUnlock()
	return len(m.data)
}
