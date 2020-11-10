package envvar

import (
	"encoding/base64"
	"fmt"
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

func GetInt(name string, defaultValue int) (int, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as an integer. Value: %s", name, valueString)
	}

	return int(value), nil
}

func GetFloat(name string, defaultValue float64) (float64, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as a float. Value: %s", name, valueString)
	}

	return value, nil
}

func GetBool(name string, defaultValue bool) (bool, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.ParseBool(valueString)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as a bool. Value: %s", name, valueString)
	}

	return value, nil
}

func GetDuration(name string, defaultValue time.Duration) (time.Duration, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := time.ParseDuration(valueString)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as a duration. Value: %s", name, valueString)
	}

	return value, nil
}

func GetBase64(name string, defaultValue []byte) ([]byte, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := base64.StdEncoding.DecodeString(valueString)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as a base64 encoded value. Value: %s", name, valueString)
	}

	return value, nil
}

func GetAddress(name string, defaultValue *net.UDPAddr) (*net.UDPAddr, error) {
	valueString, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue, nil
	}

	value, err := net.ResolveUDPAddr("udp", valueString)
	if err != nil {
		return defaultValue, fmt.Errorf("could not parse value of env var %s as an address. Value: %s", name, valueString)
	}

	return value, nil
}
