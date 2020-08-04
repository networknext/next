package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	RelayArtifactURLDev  = "https://storage.googleapis.com/development_artifacts/relay.dev.tar.gz"
	RelayArtifactURLProd = "https://storage.googleapis.com/prod_artifacts/relay.prod.tar.gz"

	RelayBackendHostnameLocal = "localhost"
	RelayBackendHostnameDev   = "relay_backend.dev.networknext.com"
	RelayBackendHostnameProd  = "relay_backend.prod.networknext.com"

	OldRelayBackendHostnameLocal = "localhost"
	OldRelayBackendHostnameDev   = "relays.v3-dev.networknext.com"
	OldRelayBackendHostnameProd  = "relays.v3.networknext.com"

	RelayBackendURLLocal = "http://" + RelayBackendHostnameLocal + ":30000"
	RelayBackendURLDev   = "http://" + RelayBackendHostnameDev
	RelayBackendURLProd  = "http://" + RelayBackendHostnameProd
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
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Hostname: %s\n", e.PortalHostname()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("AuthToken:\n\n    %s\n\n", e.AuthToken))
	sb.WriteString(fmt.Sprintf("SSHKeyFilePath: %s\n", e.SSHKeyFilePath))

	return sb.String()
}

func (e *Environment) Exists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
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
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Open(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(e); err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
}

func (e *Environment) Write() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	f, err := os.Create(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(e); err != nil {
		handleRunTimeError(fmt.Sprintf("failed to read environment %v\n", err), 1)
	}
}

func (e *Environment) Clean() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to clean environment %v\n", err), 1)
	}

	envFilePath := path.Join(homeDir, ".nextenv")

	err = os.RemoveAll(envFilePath)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to clean environment %v\n", err), 1)

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

func (e *Environment) RelayBackendURL() (string, error) {
	return e.localDevOrProd(RelayBackendURLLocal, RelayBackendURLDev, RelayBackendURLProd)
}

func (e *Environment) OldRelayBackendHostname() (string, error) {
	return e.localDevOrProd(OldRelayBackendHostnameLocal, OldRelayBackendHostnameDev, OldRelayBackendHostnameProd)
}

func (e *Environment) RelayArtifactURL() (string, error) {
	return e.devOrProd(RelayArtifactURLDev, RelayArtifactURLProd)
}

func (e *Environment) RelayBackendHostname() (string, error) {
	return e.localDevOrProd(RelayBackendHostnameLocal, RelayBackendHostnameDev, RelayBackendHostnameProd)
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
	switch e.Name {
	case "dev":
		return ifIsDev, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'dev' or 'prod'")
	}
}
