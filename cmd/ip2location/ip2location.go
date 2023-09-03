package main

import (
	"os"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/ip2location"
)

var licenseKey string

func main() {

	service := common.CreateService("ip2location")

	licenseKey = envvar.GetString("MAXMIND_LICENSE_KEY", "")

	if licenseKey == "" {
		core.Error("you must supply a license key")
		os.Exit(1)
	}

	go downloadDatabases()

	service.WaitForShutdown()
}

func downloadDatabases() {
	for {
		core.Log("---------------------------------------------------")
		err := ip2location.DownloadDatabases(licenseKey)
		if err != nil {
			core.Error("failed to download databases: %v", err)
		}
		core.Log("---------------------------------------------------")
		time.Sleep(time.Hour)
	}
}
