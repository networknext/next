package main

import (
    "net/http"
    "time"

    "github.com/gorilla/rpc/v2"
    "github.com/gorilla/rpc/v2/json2"

    "github.com/networknext/backend/modules-old/transport/looker"
    "github.com/networknext/backend/modules/common"
    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/envvar"
    "github.com/rs/cors"
)

func main() {
    service := common.CreateService("collector")

    allowedOrigins := envvar.GetList("ALLOWED_ORIGINS", []string{"127.0.0.1:3000", "127.0.0.1:8080", "127.0.0.1:80"})
    timeout := envvar.GetDuration("HTTP_TIMEOUT", time.Second*30)

    core.Log("allowed origins: ")
    for _, origin := range allowedOrigins {
        core.Log("%s", origin)
    }
    core.Log("http timeout: %s", timeout)

    rpcServer := rpc.NewServer()
    rpcServer.RegisterCodec(json2.NewCodec(), "application/json")

    publicService := PublicService{
        LookerClient: service.SetupLookerClient(),
    }
    rpcServer.RegisterService(&publicService, "")

    service.Router.Handle("/rpc", CORSHandler(rpcServer, allowedOrigins, timeout))

    service.LeaderElection()

    service.StartWebServer()

    service.WaitForShutdown()
}

type PublicService struct {
    LookerClient *looker.LookerClient
}

func (service *PublicService) LookerStats() {}
func (service *PublicService) RedisStats()  {}

func CORSHandler(handler http.Handler, allowedOrigins []string, timeout time.Duration) http.Handler {
    return cors.New(cors.Options{
        AllowedOrigins:   allowedOrigins,
        AllowCredentials: true,
        AllowedHeaders:   []string{"Content-Type"},
        AllowedMethods: []string{
            http.MethodPost,
            http.MethodGet,
            http.MethodOptions,
        },
    }).Handler(http.TimeoutHandler(handler, timeout, "Connection Timed Out!"))
}
