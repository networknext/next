package common

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"hash/fnv"
	"math/rand"
	"net"

	"github.com/networknext/next/modules/core"
)

func RandomBool() bool {
	value := rand.Intn(2)
	if value == 1 {
		return true
	} else {
		return false
	}
}

func RandomInt(min int, max int) int {
	difference := max - min
	value := rand.Intn(difference + 1)
	return value + min
}

func RandomBytes(array []byte) {
	for i := range array {
		array[i] = byte(rand.Intn(256))
	}
}

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	length = RandomInt(0, length-1) // IMPORTANT: for compatibility with NULL terminated C-strings in the SDK
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomStringFixedLength(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func RandomAddress() net.UDPAddr {
	return core.ParseAddress(fmt.Sprintf("%d.%d.%d.%d:%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(65536)))
}

type Number interface {
	constraints.Integer | constraints.Float
}

func Clamp[T Number](value *T, min T, max T) bool {
	if *value < min {
		*value = min
		return true
	} else if *value > max {
		*value = max
		return true
	}
	return false
}

func HashString(s string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return hash.Sum64()
}

func HashTag(tag string) uint64 {
	return HashString(tag)
}

func DatacenterId(datacenterName string) uint64 {
	return HashString(datacenterName)
}

func RelayId(relayAddress string) uint64 {
	return HashString(relayAddress)
}
