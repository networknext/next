package common

import (
    "fmt"
    "golang.org/x/exp/constraints"
    "hash/fnv"
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

func RandomInt(min int, max int) int {
    difference := max - min
    value := rand.Intn(difference + 1)
    return value - min
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

func ClampString(value *string, maxLength int) bool {
    // IMPORTANT: only on simple ascii strings please
    byteArray := []byte(*value)
    if len(byteArray) > maxLength {
        *value = string(byteArray[:maxLength])
        return true
    }
    return false
}

func DatacenterId(datacenterName string) uint64 {
    hash := fnv.New64a()
    hash.Write([]byte(datacenterName))
    return hash.Sum64()
}
