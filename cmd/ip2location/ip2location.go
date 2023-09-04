package main

import (
	"os"
	"fmt"
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
		core.Log("---------------------------------------------------")
		err := ip2location.DownloadDatabases(licenseKey)
		if err != nil {
			core.Error("failed to download databases: %v", err)
			goto sleep;
		}
		if bucketName != "" {
			core.Log("uploading database files to google cloud bucket")
			err := ip2location.Bash(fmt.Sprintf("gsutil cp GeoIP2-*.mmdb gs://%s", bucketName))
			if err != nil {
				core.Error("failed to upload database files: %v", err)
				goto sleep;
			}
			core.Log("success!")
		}
	sleep:
		fmt.Printf("sleeping...\n")
		core.Log("---------------------------------------------------")
		time.Sleep(time.Hour)
	}
}
