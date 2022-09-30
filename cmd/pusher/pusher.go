/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"io/ioutil"
	"net"
	"time"

	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/oschwald/geoip2-golang"
)

func main() {

	service := common.CreateService("pusher")

	service.SetupGCPStorage()

	databaseDownloadURL := envvar.GetString("DATABASE_DOWNLOAD_URI", "gs://happy_path_testing/database.bin")
	overlayDownloadURL := envvar.GetString("OVERLAY_DOWNLOAD_URI", "gs://happy_path_testing/overlay.bin")

	databaseFileName := common.GetFileNameFromPath(databaseDownloadURL)
	overlayFileName := common.GetFileNameFromPath(overlayDownloadURL)

	fileSyncConfig := &common.FileSyncConfig{
		FileGroups: []common.FileSyncGroup{
			{
				SyncInterval:   envvar.GetDuration("LOCATION_FILE_REFRESH_INTERVAL", time.Minute*5),
				ValidationFunc: validateLocationFiles,
				SaveBucket:     envvar.GetString("LOCATION_FILE_BUCKET_PATH", "gs://happy_path_testing"),
				ReceivingVMs:   service.GcpStorage.GetMIGInstanceNamesEnv("SERVER_BACKEND_MIG_NAME", ""),
				FileConfigs: []common.SyncFile{
					{
						Name:        "GeoIP2-ISP.mmdb", // download URL is a compress tar.gz so we need to know single file name
						DownloadURL: envvar.GetString("MAXMIND_ISP_DOWNLOAD_URI", "gs://happy_path_testing/GeoIP2-ISP.tar.gz"),
					},
					{
						Name:        "GeoIP2-City.mmdb", // download URL is a compress tar.gz so we need to know single file name
						DownloadURL: envvar.GetString("MAXMIND_CITY_DOWNLOAD_URI", "gs://happy_path_testing/GeoIP2-City.tar.gz"),
					},
				},
			},
			{
				SyncInterval:   envvar.GetDuration("BIN_FILE_REFRESH_INTERVAL", time.Minute*1),
				ValidationFunc: validateBinFiles,
				ReceivingVMs:   service.GcpStorage.GetMIGInstanceNamesEnv("RELAY_GATEWAY_MIG_NAME", ""),
				FileConfigs: []common.SyncFile{
					{
						Name:        databaseFileName,
						DownloadURL: databaseDownloadURL,
					},
					{
						Name:        overlayFileName,
						DownloadURL: overlayDownloadURL,
					},
				},
			},
		},
	}

	fileSyncConfig.Print()

	service.LeaderElection()

	service.StartFileSync(fileSyncConfig)

	service.StartWebServer()

	service.WaitForShutdown()
}

// todo: keep an eye on these structures while modules-new is being built out
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

// todo: keep an eye on these structures while modules-new is being built out
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
