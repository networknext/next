package common

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/networknext/backend/modules/core"
)

func RandomBool() bool {
	value := rand.Intn(2)
	if value == 1 {
		return true
	} else {
		return false
	}
}

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomAddress() net.UDPAddr {
	return *core.ParseAddress(fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(65536)))
}
