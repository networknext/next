package main

import (
	"fmt"
	"os"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/ip2location"
)

var licenseKey string
var bucketName string

func main() {

	service := common.CreateService("ip2location")

	licenseKey = envvar.GetString("MAXMIND_LICENSE_KEY", "")

	bucketName = envvar.GetString("IP2LOCATION_BUCKET_NAME", "")

	if licenseKey == "" {
		core.Error("you must supply a license key")
		os.Exit(1)
	}

	go downloadDatabases()

	service.WaitForShutdown()
}

func downloadDatabases() {
	for {
		core.Debug("---------------------------------------------------")
		err := ip2location.DownloadDatabases_MaxMind(licenseKey)
		if err != nil {
			core.Error("failed to download databases: %v", err)
			goto sleep
		}
		if bucketName != "" {
			core.Debug("uploading database files to google cloud bucket")
			err := ip2location.Bash(fmt.Sprintf("gsutil cp GeoIP2-*.mmdb gs://%s", bucketName))
			if err != nil {
				core.Error("failed to upload database files: %v", err)
				goto sleep
			}
			core.Debug("success!")
		}
	sleep:
		core.Debug("sleeping...")
		core.Debug("---------------------------------------------------")
		time.Sleep(time.Hour)
	}
}
