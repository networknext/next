package core_test

import (
	"math/rand"
)

func RandomPublicKey() []byte {
	arr := make([]byte, 64)
	for i := 0; i < 64; i++ {
		arr[i] = byte(rand.Int())
	}
	return arr
}

func RandomString(length int) string {
	arr := make([]byte, length)
	for i := 0; i < length; i++ {
		arr[i] = byte(rand.Int()%26 + 65)
	}
	return string(arr)
}
