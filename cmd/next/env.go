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
	PortalHostnameLocal   = "localhost:20000"
	PortalHostnameDev     = "portal-dev.networknext.com"
	PortalHostnameStaging = "portal-staging.networknext.com"
	PortalHostnameProd    = "portal.networknext.com"

	RouterPublicKeyLocal   = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyDev     = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyStaging = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
	RouterPublicKeyProd    = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="

	RelayArtifactURLDev     = "https://storage.googleapis.com/development_artifacts/relay.dev.tar.gz"
	RelayArtifactURLStaging = "https://storage.googleapis.com/staging_artifacts/relay.dev.tar.gz"
	RelayArtifactURLProd    = "https://storage.googleapis.com/prod_artifacts/relay.prod.tar.gz"

	RelayBackendHostnameLocal   = "localhost"
	RelayBackendHostnameDev     = "relay_backend.dev.networknext.com"
	RelayBackendHostnameStaging = "relay_backend.staging.networknext.com"
	RelayBackendHostnameProd    = "relay_backend.prod.networknext.com"

	RelayBackendURLLocal   = "http://" + RelayBackendHostnameLocal + ":30000"
	RelayBackendURLDev     = "http://" + RelayBackendHostnameDev
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
	if hostname, err := e.switchEnvLocal(PortalHostnameLocal, PortalHostnameDev, PortalHostnameStaging, PortalHostnameProd); err == nil {
		return hostname
	}
	return e.Hostname
}

func (e *Environment) RouterPublicKey() (string, error) {
	return e.switchEnvLocal(RouterPublicKeyLocal, RouterPublicKeyDev, RouterPublicKeyStaging, RouterPublicKeyProd)
}

func (e *Environment) RelayBackendURL() (string, error) {
	return e.switchEnvLocal(RelayBackendURLLocal, RelayBackendURLDev, RelayBackendURLStaging, RelayBackendURLProd)
}

func (e *Environment) RelayArtifactURL() (string, error) {
	return e.switchEnv(RelayArtifactURLDev, RelayArtifactURLStaging, RelayArtifactURLProd)
}

func (e *Environment) RelayBackendHostname() (string, error) {
	return e.switchEnvLocal(RelayBackendHostnameLocal, RelayBackendHostnameDev, RelayBackendHostnameStaging, RelayBackendHostnameProd)
}

func (e *Environment) switchEnvLocal(ifIsLocal, ifIsDev, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "local":
		return ifIsLocal, nil
	case "dev":
		return ifIsDev, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'local', 'dev', 'staging', or 'prod'")
	}
}

func (e *Environment) switchEnv(ifIsDev, ifIsStaging, ifIsProd string) (string, error) {
	switch e.Name {
	case "dev":
		return ifIsDev, nil
	case "staging":
		return ifIsStaging, nil
	case "prod":
		return ifIsProd, nil
	default:
		return "", errors.New("Environment does not match 'dev', 'staging', or 'prod'")
	}
}
