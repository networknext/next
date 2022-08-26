package common

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

type Service struct {
	ServiceName string
	BuildTime string
	CommitMessage string
	CommitHash string
	Router mux.Router
}

func CreateService(serviceName string) *Service {

	service := Service{}
	service.ServiceName = serviceName
	service.CommitMessage = commitMessage
	service.CommitHash = commitHash
	service.BuildTime = buildTime

	fmt.Printf("%s\n", service.ServiceName)

	fmt.Printf("commit: %s [%s] (%s)\n", service.CommitMessage, service.CommitHash, service.BuildTime)

	env := backend.GetEnv()

	fmt.Printf("env: %s\n", env)

	service.Router.HandleFunc("/health", transport.HealthHandlerFunc())
	service.Router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))

	return &service
}

func (service *Service) StartWebServer() {
	port := envvar.Get("HTTP_PORT", "80")
	fmt.Printf("starting http server on port %s\n", port)
	go func() {
		err := http.ListenAndServe(":"+port, &service.Router)
		if err != nil {
			core.Error("error starting http server: %v", err)
			os.Exit(1)
		}
	}()
}

func (service *Service) WaitForShutdown() {
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	core.Debug("received shutdown signal")
	// todo: probably need to wait for some stuff...
	core.Debug("successfully shutdown")
}
