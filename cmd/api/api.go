package main

import (
	"github.com/networknext/backend/modules/common"
)

func main() {

	service := common.CreateService("api")

	service.StartWebServer()

	service.WaitForShutdown()
}
