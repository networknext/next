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
	PortalHostnameLocal = "localhost:20000"
	PortalHostnameDev   = "portal-dev.networknext.com"
	PortalHostnameProd  = "portal.networknext.com"

	RouterPublicKeyLocal = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyDev   = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyProd  = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="

	RelayBackendHostnameLocal = "http://localhost:30000"
	RelayBackendHostnameDev   = "http://relay_backend.dev.networknext.com:40000"
	RelayBackendHostnameProd  = "http://relay_backend.prod.networknext.com"

	OldRelayBackendHostnameLocal = "localhost"
	OldRelayBackendHostnameDev   = "relays.v3-dev.networknext.com"
	OldRelayBackendHostnameProd  = "relays.v3.networknext.com"

	RelayArtifactURLDev  = "https://storage.googleapis.com/artifacts.network-next-v3-dev.appspot.com/relay.dev.tar.gz"
	RelayArtifactURLProd = "https://storage.googleapis.com/us.artifacts.network-next-v3-prod.appspot.com/relay.prod.tar.gz"
)

type Environment struct {
	CLIRelease   string `json:"-"`
	CLIBuildTime string `json:"-"`

	RemoteRelease   string `json:"-"`
	RemoteBuildTime string `json:"-"`

	Name           string `json:"name"`
	Hostname       string `json:"hostname"`
	AuthToken      string `json:"auth_token"`
	SSHKeyFilePath string `json:"ssh_key_filepath`
}

func (e *Environment) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Environment: %s\n", e.Name))
	sb.WriteString(fmt.Sprintf("Hostname: %s\n", e.PortalHostname()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("CLI Release: %s\n", e.CLIRelease))
	sb.WriteString(fmt.Sprintf("CLI Build Time: %s\n", e.CLIBuildTime))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Remote Release: %s\n", e.RemoteRelease))
	sb.WriteString(fmt.Sprintf("Remote Build Time: %s\n", e.RemoteBuildTime))
	sb.WriteString("\n")
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
	if hostname, err := e.localDevOrProd(PortalHostnameLocal, PortalHostnameDev, PortalHostnameProd); err == nil {
		return hostname
	}
	return e.Hostname
}

func (e *Environment) RouterPublicKey() (string, error) {
	return e.localDevOrProd(RouterPublicKeyLocal, RouterPublicKeyDev, RouterPublicKeyProd)
}

func (e *Environment) RelayBackendHostname() (string, error) {
	return e.localDevOrProd(RelayBackendHostnameLocal, RelayBackendHostnameDev, RelayBackendHostnameProd)
}

func (e *Environment) OldRelayBackendHostname() (string, error) {
	return e.localDevOrProd(OldRelayBackendHostnameLocal, OldRelayBackendHostnameDev, OldRelayBackendHostnameProd)
}

func (e *Environment) RelayArtifactURL() (string, error) {
	return e.devOrProd(RelayArtifactURLDev, RelayArtifactURLProd)
}

func (e *Environment) localDevOrProd(ifIsLocal, ifIsDev, ifIsProd string) (string, error) {
	switch e.Name {
	case "local":
		return ifIsLocal, nil
	case "dev":
		return ifIsDev, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'local', 'dev', or 'prod'")
	}
}

func (e *Environment) devOrProd(ifIsDev, ifIsProd string) (string, error) {
	switch e.Hostname {
	case "dev":
		return ifIsDev, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'dev' or 'prod'")
	}
}
