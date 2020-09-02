package envvar

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
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
