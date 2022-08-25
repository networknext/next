package envvar

import (
	"encoding/base64"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func Exists(name string) bool {
	_, ok := os.LookupEnv(name)
	return ok
}

func Get(name string, defaultValue string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	return value
}

func GetList(name string, defaultValue []string) []string {
	valueStrings, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value := strings.Split(valueStrings, ",")
	return value
}

func GetInt(name string, defaultValue int) int {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return defaultValue
	}
	return int(value)
}

func GetFloat(name string, defaultValue float64) float64 {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetBool(name string, defaultValue bool) bool {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueString)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetDuration(name string, defaultValue time.Duration) time.Duration {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := time.ParseDuration(valueString)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetBase64(name string, defaultValue []byte) []byte {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := base64.StdEncoding.DecodeString(valueString)
	if err != nil {
		return defaultValue
	}
	return value
}

func GetAddress(name string, defaultValue *net.UDPAddr) *net.UDPAddr {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := net.ResolveUDPAddr("udp", valueString)
	if err != nil {
		return defaultValue
	}
	return value
}
