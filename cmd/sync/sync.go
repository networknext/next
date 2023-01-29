/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
)

func main() {
	fmt.Printf("sync service\n")
}

/*
import (
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

func main() {

	service := common.CreateService("sync")

	fileSyncConfig := common.CreateFileSyncConfig()

	ispSyncFile := common.SyncFile{
		Name:        "GeoIP2-ISP.mmdb",
		DownloadURL: envvar.GetString("MAXMIND_ISP_DOWNLOAD_URI", "gs://network-next-local/GeoIP2-ISP.tar.gz"),
	}

	citySyncFile := common.SyncFile{
		Name:        "GeoIP2-City.mmdb",
		DownloadURL: envvar.GetString("MAXMIND_CITY_DOWNLOAD_URI", "gs://network-next-local/GeoIP2-City.tar.gz"),
	}

	fileSyncConfig.AddFileSyncGroup(
		"ip2location",
		envvar.GetDuration("LOCATION_FILE_REFRESH_INTERVAL", 24*time.Hour),
		envvar.GetList("LOCATION_FILE_DESTINATION_MIGS", []string{}),
		"",
		envvar.GetString("LOCATION_FILE_BUCKET_PATH", "gs://network-next-local-upload"),
		service.ValidateIP2Location,
		citySyncFile,
		ispSyncFile,
	)

	databaseSyncFile := common.SyncFile{
		DownloadURL: envvar.GetString("DATABASE_DOWNLOAD_URI", "gs://network-next-local/database.bin"),
	}

	overlaySyncFile := common.SyncFile{
		DownloadURL: envvar.GetString("OVERLAY_DOWNLOAD_URI", "gs://network-next-local/overlay.bin"),
	}

	fileSyncConfig.AddFileSyncGroup(
		"database",
		envvar.GetDuration("BIN_FILE_REFRESH_INTERVAL", time.Minute),
		envvar.GetList("BIN_FILE_DESTINATION_MIGS", []string{}),
		envvar.GetString("OUTPUT_PATH", ""),
		"",
		service.ValidateBinFiles,       // todo: no, create your own function here to load the database.bin and validate them. do not call out to service here, we need to validate BEFORE we move the files to a place where service tries to load them
		databaseSyncFile,
		overlaySyncFile,
	)

	service.SyncFiles(fileSyncConfig)

	service.LeaderElection(true)

	service.StartWebServer()

	service.WaitForShutdown()
}
*/