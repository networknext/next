package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

const (
	PortalHostnameLocal = "localhost:20000"
	PortalHostnameDev   = "portal.dev.networknext.com"
	PortalHostnameProd  = "portal.dev.networknext.com"
)

type Environment struct {
	Hostname       string `json:"hostname"`
	AuthToken      string `json:"auth_token"`
	SSHKeyFilePath string `json:"ssh_key_filepath`
}

func (e *Environment) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Hostname: %s\n", e.Hostname))
	sb.WriteString(fmt.Sprintf("AuthToken: %s\n", e.AuthToken))
	sb.WriteString(fmt.Sprintf("SSHKeyFilePath: %s\n", e.SSHKeyFilePath))

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

func (e *Environment) PortalHostname() string {
	switch e.Hostname {
	case "local":
		return PortalHostnameLocal
	case "dev":
		return PortalHostnameDev
	case "prod":
		return PortalHostnameProd
	}

	return e.Hostname
}
