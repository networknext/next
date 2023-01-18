/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"io/ioutil"
	"net"
	"time"

	// we should not depend on the old routing geolocation code. it's dead code
	"github.com/networknext/backend/modules-old/routing"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/oschwald/geoip2-golang"
)

func main() {

	service := common.CreateService("pusher")

	fileSyncConfig := common.CreateFileSyncConfig()

	ispSyncFile := common.SyncFile{
		Name:        "GeoIP2-ISP.mmdb", // download URL is a compress tar.gz so we need to know single file name
		DownloadURL: envvar.GetString("MAXMIND_ISP_DOWNLOAD_URI", "gs://network-next-local/GeoIP2-ISP.tar.gz"),
	}

	citySyncFile := common.SyncFile{
		Name:        "GeoIP2-City.mmdb", // download URL is a compress tar.gz so we need to know single file name
		DownloadURL: envvar.GetString("MAXMIND_CITY_DOWNLOAD_URI", "gs://network-next-local/GeoIP2-City.tar.gz"),
	}

	fileSyncConfig.AddFileSyncGroup(
		"ip2location",
		envvar.GetDuration("LOCATION_FILE_REFRESH_INTERVAL", 5*time.Minute),
		envvar.GetString("SERVER_BACKEND_MIG_NAME", ""),
		envvar.GetString("LOCATION_FILE_BUCKET_PATH", "gs://network-next-local-upload"),
		validateLocationFiles,
		ispSyncFile,
		citySyncFile,
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
		envvar.GetString("RELAY_GATEWAY_MIG_NAME", ""),
		"",
		validateBinFiles,
		databaseSyncFile,
		overlaySyncFile,
	)

	service.SyncFiles(fileSyncConfig)

	service.LeaderElection(true)

	service.StartWebServer()

	service.WaitForShutdown()
}

func validateLocationFiles(locationFiles []string) bool {

	ipStr := "192.0.2.1"
	testIP := net.ParseIP(ipStr)
	if testIP == nil {
		return false
	}

	ispFile := locationFiles[0]

	ispBytes, err := ioutil.ReadFile(ispFile)
	if err != nil {
		core.Error("failed to read location file: %v", err)
		return false
	}

	ispReader, err := geoip2.FromBytes(ispBytes)
	if err != nil {
		core.Error("failed to create geo reader: %v", err)
		return false
	}
	defer ispReader.Close()

	cityFile := locationFiles[1]

	cityBytes, err := ioutil.ReadFile(cityFile)
	if err != nil {
		core.Error("failed to read location file: %v", err)
		return false
	}

	cityReader, err := geoip2.FromBytes(cityBytes)
	if err != nil {
		core.Error("failed to create geo reader: %v", err)
		return false
	}

	location := routing.Location{}

	city, err := cityReader.City(testIP)
	if err != nil {
		core.Error("failed to look up city: %v", err)
		return false
	}
	defer cityReader.Close()

	location.Latitude = float32(city.Location.Latitude)
	location.Longitude = float32(city.Location.Longitude)

	isp, err := ispReader.ISP(testIP)
	if err != nil {
		core.Error("failed to look up ISP: %v", err)
		return false
	}

	location.ISP = isp.ISP
	location.ASN = int(isp.AutonomousSystemNumber)

	if location == routing.LocationNullIsland {
		core.Error("location returned as null island")
		return false
	}

	return true
}

func validateBinFiles(binFiles []string) bool {

	databaseFile := binFiles[0]
	overlayFile := binFiles[1]

	databaseWrapper := routing.DatabaseBinWrapper{}
	overlayWrapper := routing.OverlayBinWrapper{}

	if err := databaseWrapper.ReadDatabaseBinFile(databaseFile); err != nil {
		core.Error("failed to read database file: %v", err)
		return false
	}

	if databaseWrapper.IsEmpty() {
		core.Error("database file can not be empty")
		return false
	}

	if err := overlayWrapper.ReadOverlayBinFile(overlayFile); err != nil {
		core.Error("failed to read overlay file: %v", err)
		return false
	}

	return true
}
