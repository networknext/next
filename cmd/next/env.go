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
	PortalHostnameLocal   = "localhost:20000"
	PortalHostnameDev     = "portal-dev.networknext.com"
	PortalHostnameNRB     = "104.197.11.70"
	PortalHostnameStaging = "portal-staging.networknext.com"
	PortalHostnameProd    = "portal.networknext.com"

	RouterPublicKeyLocal   = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyDev     = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyNRB     = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyStaging = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyProd    = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="

	RelayArtifactURLDev     = "https://storage.googleapis.com/development_artifacts/relay.dev.tar.gz"
	RelayArtifactURLNRB     = "https://storage.googleapis.com/nrb_artifacts/relay.nrb.tar.gz"
	RelayArtifactURLStaging = "https://storage.googleapis.com/staging_artifacts/relay.staging.tar.gz"
	RelayArtifactURLProd    = "https://storage.googleapis.com/prod_artifacts/relay.prod.tar.gz"

	RelayBackendHostnameLocal   = "localhost"
	RelayBackendHostnameDev     = "34.117.47.154"
	RelayBackendHostnameNRB     = "10.128.0.7"
	RelayBackendHostnameStaging = "35.190.44.124"
	RelayBackendHostnameProd    = "35.227.196.44"

	RelayBackendURLLocal   = "http://" + RelayBackendHostnameLocal + ":30005"
	RelayBackendURLDev     = "http://" + RelayBackendHostnameDev
	RelayBackendURLNRB     = "http://" + RelayBackendHostnameNRB
	RelayBackendURLStaging = "http://" + RelayBackendHostnameStaging
	RelayBackendURLProd    = "http://" + RelayBackendHostnameProd
)

type Environment struct {
	CLIRelease   string `json:"-"`
	CLIBuildTime string `json:"-"`

	RemoteRelease   string `json:"-"`
	RemoteBuildTime string `json:"-"`

	Name           string `json:"name"`
	Hostname       string `json:"hostname"`
	AuthToken      string `json:"auth_token"`
	SSHKeyFilePath string `json:"ssh_key_filepath"`
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
	if hostname, err := e.switchEnvLocal(PortalHostnameLocal, PortalHostnameDev, PortalHostnameNRB, PortalHostnameStaging, PortalHostnameProd); err == nil {
		return hostname
	}
	return e.Hostname
}

func (e *Environment) RouterPublicKey() (string, error) {
	return e.switchEnvLocal(RouterPublicKeyLocal, RouterPublicKeyDev, RouterPublicKeyNRB, RouterPublicKeyStaging, RouterPublicKeyProd)
}

func (e *Environment) RelayBackendURL() (string, error) {
	return e.switchEnvLocal(RelayBackendURLLocal, RelayBackendURLDev, RelayBackendURLNRB, RelayBackendURLStaging, RelayBackendURLProd)
}

func (e *Environment) RelayArtifactURL() (string, error) {
	return e.switchEnv(RelayArtifactURLDev, RelayArtifactURLNRB, RelayArtifactURLStaging, RelayArtifactURLProd)
}

func (e *Environment) RelayBackendHostname() (string, error) {
	return e.switchEnvLocal(RelayBackendHostnameLocal, RelayBackendHostnameDev, RelayBackendHostnameNRB, RelayBackendHostnameStaging, RelayBackendHostnameProd)
}

func (e *Environment) switchEnvLocal(ifIsLocal, ifIsDev, ifIsNRB, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "local":
		return ifIsLocal, nil
	case "dev":
		return ifIsDev, nil
	case "nrb":
		return ifIsNRB, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'local', 'dev', 'staging', or 'prod'")
	}
}

func (e *Environment) switchEnv(ifIsDev, ifIsNRB, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "dev":
		return ifIsDev, nil
	case "nrb":
		return ifIsNRB, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'dev', 'staging', or 'prod'")
	}
}
