package main

import (
	// ...

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var redisHostName string
var redisPassword string

var magicUpdateSeconds int

func main() {

	service := common.CreateService("raspberry_backend")

	redisHostName = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Debug("redis hostname: %s", redisHostName)
	core.Debug("redis password: %s", redisPassword)

	// todo: setup handler for "raspberry_server" endpoint

	service.StartWebServer()

	service.WaitForShutdown()
}
