/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

type FileConfig struct {
	UploadVMs []string
}

func main() {

	service := common.CreateService("relay_pusher")

	service.SetupGCPStorage()

	service.LeaderElection()

	locationFiles := []common.MaxmindFile{
		{
			Name:        envvar.GetString("MAXMIND_ISP_FILE_NAME", "GeoIP2-ISP.mmdb"),
			Path:        envvar.GetString("MAXMIND_ISP_DOWNLOAD_PATH", "./GeoIP2-ISP.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_ISP_DB_URI", ""),
			Type:        common.MAXMIND_ISP,
			Env:         service.Env,
		},
		{
			Name:        envvar.GetString("MAXMIND_CITY_FILE_NAME", "GeoIP2-CITY.mmdb"),
			Path:        envvar.GetString("MAXMIND_CITY_DOWNLOAD_PATH", "./GeoIP2-City.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_CITY_DB_URI", ""),
			Type:        common.MAXMIND_CITY,
			Env:         service.Env,
		},
	}

	RefreshLocationFiles(service, locationFiles)

	binFiles := []common.BinFile{
		{
			Name: envvar.GetString("DATABASE_FILE_NAME", "database.bin"),
			Path: envvar.GetString("DATABASE_DOWNLOAD_PATH", "./database.bin"),
			Type: common.BIN_DATABASE,
		},
		{
			Name: envvar.GetString("OVERLAY_FILE_NAME", "overlay.bin"),
			Path: envvar.GetString("OVERLAY_DOWNLOAD_PATH", "./overlay.bin"),
			Type: common.BIN_OVERLAY,
		},
	}

	RefreshBinFiles(service, binFiles)

	service.StartWebServer()

	service.WaitForShutdown()
}

func RefreshLocationFiles(service *common.Service, locationFiles []common.MaxmindFile) {

	refreshInterval := envvar.GetDuration("LOCATION_FILE_REFRESH_INTERVAL", time.Hour*24)
	locationFileBucketPath := envvar.GetString("LOCATION_FILE_BUCKET_PATH", "gs://happy_path_testing")
	serverBackendInstanceNames := service.GcpStorage.GetMIGInstanceNamesEnv("SERVER_BACKEND_MIG_NAME", "")

	go func() {

		ticker := time.NewTicker(refreshInterval)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				// Download loop
				for _, locationFile := range locationFiles {

					// Toggle to avoid excess downloads - happy path - local testing
					if locationFile.DownloadURL == "" {
						continue
					}

					fileExtensionTokens := strings.Split(locationFile.Path, ".")
					fileExtension := fileExtensionTokens[len(fileExtensionTokens)-1]

					// Download the file to local storage
					if err := service.DownloadGzipFileFromURL(locationFile.DownloadURL, locationFile.Path, fileExtension); err != nil {
						core.Error("failed to download %s file: %v", locationFile.Type, err)
						continue
					}

					// Validate the file
					if err := locationFile.Validate(service.Context); err != nil {
						core.Error("failed to validate %s file: %v", locationFile.Type, err)
						continue
					}

					// Don't upload unless leader VM
					if !service.IsLeader() {
						continue
					}

					fullArtifactPath := fmt.Sprintf("%s/%s", locationFileBucketPath, locationFile.Name)

					// Upload file to GCP for VMs that are replacing (part of startup script)
					if err := service.GcpStorage.CopyFromLocalToBucket(service.Context, locationFile.Path, fullArtifactPath); err != nil {
						core.Error("failed to upload location file to GCP storage: %v", err)
						continue
					}

					// Upload file directly to VMs to update them
					// Upload == local file location here
					if err := service.UploadFileToGCPVirtualMachines(locationFile.Path, locationFile.Path, serverBackendInstanceNames); err != nil {
						core.Error("failed to upload location file to GCP VMs: %v", err)
					}
				}
			}
		}
	}()
}

func RefreshBinFiles(service *common.Service, binFiles []common.BinFile) {

	go func() {

		binFileRefreshInterval := envvar.GetDuration("BIN_FILE_REFRESH_INTERVAL", time.Minute*1)
		binFileBucketPath := envvar.GetString("BIN_FILE_BUCKET_PATH", "gs://happy_path_testing")
		relayGatewayInstanceNames := service.GcpStorage.GetMIGInstanceNamesEnv("RELAY_GATEWAY_MIG_NAME", "")

		ticker := time.NewTicker(binFileRefreshInterval)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				// Download and Verify
				for _, binFile := range binFiles {

					fullArtifactPath := fmt.Sprintf("%s/%s", binFileBucketPath, binFile.Name)

					// Download from GCP bucket (Portal/Next tool uploads to bucket after creation)
					if err := service.GcpStorage.CopyFromBucketToLocal(service.Context, fullArtifactPath, binFile.Path); err != nil {
						core.Error("failed to download %s file: %v", binFile.Type, err)
						continue
					}

					binFileRef, err := os.Open(binFile.Path)
					if err != nil {
						core.Error("failed to open database file")
						continue
					}
					defer binFileRef.Close()

					// Validate bin file
					if err := binFile.Validate(binFileRef); err != nil {
						core.Error("failed to validate database file: %v", err)
						continue
					}

					// Don't upload unless leader VM
					if !service.IsLeader() {
						continue
					}

					// Upload file directly to VMs to update them
					// Upload == local file path here
					if err := service.UploadFileToGCPVirtualMachines(binFile.Path, binFile.Path, relayGatewayInstanceNames); err != nil {
						core.Error("failed to upload %s file to GCP VMs: %v", binFile.Type, err)
					}
				}
			}
		}
	}()
}
