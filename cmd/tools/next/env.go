package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

const (
	RouterPublicKeyDev  = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyProd = "placeholder"

	BackendHostnameDev  = "http://relay_backend.dev.spacecats.net:40000"
	BackendHostnameProd = "http://relay_backend.prod.spacecats.net:40000"
)

type Environment struct {
	Hostname         string `json:"hostname"`
	AuthToken        string `json:"auth_token"`
	SSHKeyFilePath   string `json:"ssh_key_filepath`
	RelayEnvironment string `json:"relay_environment"`
}

func (e *Environment) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Hostname: %s\n", e.Hostname))
	sb.WriteString(fmt.Sprintf("AuthToken: %s\n", e.AuthToken))
	sb.WriteString(fmt.Sprintf("SSHKeyFilePath: %s\n", e.SSHKeyFilePath))
	sb.WriteString(fmt.Sprintf("RelayEnvironment: %s\n", e.RelayEnvironment))

	return sb.String()
}

func (e *Environment) Exists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to read environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	if _, err := os.Stat(envFilePath); err != nil {
		return false
	}

	return true
}

func (e *Environment) Read() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to read environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Open(envFilePath)
	if err != nil {
		log.Fatal("failed to read environment", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(e); err != nil {
		log.Fatal("failed to read environment", err)
	}
}

func (e *Environment) Write() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to write environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Create(envFilePath)
	if err != nil {
		log.Fatal("failed to write environment", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		log.Fatal("failed to write environment", err)
	}
}

func (e *Environment) Clean() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("failed to clean environment", err)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	err = os.RemoveAll(envFilePath)
	if err != nil {
		log.Fatal("failed to clean environment", err)
	}
}

func (e *Environment) RouterPublicKey() (string, error) {
	return e.devOrProd(RouterPublicKeyDev, RouterPublicKeyProd)
}

func (e *Environment) BackendHostname() (string, error) {
	return e.devOrProd(BackendHostnameDev, BackendHostnameProd)
}

func (e *Environment) devOrProd(ifIsDev, ifIsProd string) (string, error) {
	switch e.RelayEnvironment {
	case "dev":
		return ifIsDev, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Invalid relay environment, please set it to either 'dev' or 'prod'")
	}
}
