package main

import (
	// "context"
	"fmt"
	"os"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"

	"github.com/russellcardullo/go-pingdom/pingdom"
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {

	pingdomApiToken := envvar.Get("PINGDOM_API_TOKEN", "nEVIN5R8WjzcRDEO6FA6wMwjBgla63NqNbqZ2dX-D8TTtPUF_3sGAqg_d0db2OPxkdCdMh8")

	client, err := pingdom.NewClientWithConfig(pingdom.ClientConfig{
		APIToken: pingdomApiToken,
	})

	fromTime := time.Now().Add(time.Hour * -24).Unix()
	perfRequest := pingdom.SummaryPerformanceRequest{
		Id:            6872585,
		From:          int(fromTime),
		To:            int(time.Now().Unix()),
		Resolution:    "day",
		IncludeUptime: true,
	}

	if err := perfRequest.Valid(); err != nil {
		core.Error("request not valid: %v", err)
		return 1
	}

	checks, err := client.Checks.SummaryPerformance(perfRequest)
	if err != nil {
		core.Error("failed to get checks: %v", err)
		return 1
	}

	fmt.Printf("perf: %+v\n", checks)

	return 0

}
