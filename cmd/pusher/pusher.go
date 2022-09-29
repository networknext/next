/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

const (
	ISP      = "isp"
	CITY     = "city"
	DATABASE = "database"
	OVERLAY  = "overlay"
)

type FileConfig struct {
	Name        string
	DownloadURL string
	VMFilePath  string
	UploadVMs   []string
}

type LocationFile struct {
	Config         FileConfig
	ValidationFunc func(context.Context, string, string) error
}

type BinFile struct {
	Config FileConfig
}

func main() {

	service := common.CreateService("relay_pusher")

	service.SetupGCPStorage()

	serverBackendInstanceNames := service.GcpStorage.GetMIGInstanceNamesEnv("SERVER_BACKEND_MIG_NAME", "")

	relayGatewayInstanceNames := service.GcpStorage.GetMIGInstanceNamesEnv("RELAY_GATEWAY_MIG_NAME", "")

	locationFiles := make(map[string]LocationFile)
	locationFiles[ISP] = LocationFile{
		Config: FileConfig{
			Name:        envvar.GetString("MAXMIND_ISP_FILE_NAME", "GeoIP2-ISP.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_ISP_DB_URI", ""),
			VMFilePath:  envvar.GetString("MAXMIND_ISP_DOWNLOAD_PATH", "./GeoIP2-ISP.mmdb"),
			UploadVMs:   serverBackendInstanceNames,
		},
		ValidationFunc: common.ValidateISPFile,
	}

	locationFiles[CITY] = LocationFile{
		Config: FileConfig{
			Name:        envvar.GetString("MAXMIND_CITY_FILE_NAME", "GeoIP2-City.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_CITY_DB_URI", ""),
			VMFilePath:  envvar.GetString("MAXMIND_CITY_DOWNLOAD_PATH", "./GeoIP2-City.mmdb"),
			UploadVMs:   serverBackendInstanceNames,
		},
		ValidationFunc: common.ValidateCityFile,
	}

	RefreshLocationFiles(service, locationFiles)

	binFiles := make(map[string]BinFile)
	binFiles[DATABASE] = BinFile{
		Config: FileConfig{
			Name:       envvar.GetString("DATABASE_FILE_NAME", "database.bin"),
			VMFilePath: envvar.GetString("DATABASE_DOWNLOAD_PATH", "./database.bin"),
			UploadVMs:  relayGatewayInstanceNames,
		},
	}

	binFiles[OVERLAY] = BinFile{
		Config: FileConfig{
			Name:       envvar.GetString("OVERLAY_FILE_NAME", "overlay.bin"),
			VMFilePath: envvar.GetString("OVERLAY_DOWNLOAD_PATH", "./overlay.bin"),
			UploadVMs:  relayGatewayInstanceNames,
		},
	}

	RefreshBinFiles(service, binFiles)

	service.StartWebServer()

	service.LeaderElection()

	service.WaitForShutdown()
}

func RefreshLocationFiles(service *common.Service, files map[string]LocationFile) {

	refreshInterval := envvar.GetDuration("LOCATION_FILE_REFRESH_INTERVAL", time.Hour*24)
	locationFileBucketPath := envvar.GetString("LOCATION_FILE_BUCKET_PATH", "gs://happy_path_testing")

	go func() {

		ticker := time.NewTicker(refreshInterval)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				// Download loop
				for fileType, file := range files {

					config := file.Config

					// Toggle to avoid excess downloads - happy path - local testing
					if config.DownloadURL == "" {
						continue
					}

					fileExtensionTokens := strings.Split(config.Name, ".")
					fileExtension := fileExtensionTokens[len(fileExtensionTokens)-1]

					// Download the file to local storage
					if err := service.DownloadGzipFileFromURL(config.DownloadURL, config.VMFilePath, fileExtension); err != nil {
						core.Error("failed to download %s file: %v", fileType, err)
						continue
					}

					// Validate the file
					if err := file.ValidationFunc(service.Context, service.Env, config.VMFilePath); err != nil {
						core.Error("failed to validate %s file: %v", fileType, err)
						continue
					}

					// Don't upload unless leader VM
					if !service.IsLeader() {
						continue
					}

					fullArtifactPath := fmt.Sprintf("%s/%s", locationFileBucketPath, config.Name)

					// Upload file to GCP for VMs that are replacing (part of startup script)
					if err := service.GcpStorage.CopyFromLocalToBucket(service.Context, config.VMFilePath, fullArtifactPath); err != nil {
						core.Error("failed to upload location file to GCP storage: %v", err)
						continue
					}

					// Upload file directly to VMs to update them
					// Upload == local file location here
					if err := service.UploadFileToGCPVirtualMachines(config.VMFilePath, config.VMFilePath, config.UploadVMs); err != nil {
						core.Error("failed to upload location file to GCP VMs: %v", err)
					}
				}
			}
		}
	}()
}

func RefreshBinFiles(service *common.Service, files map[string]BinFile) {

	go func() {

		binFileRefreshInterval := envvar.GetDuration("BIN_FILE_REFRESH_INTERVAL", time.Minute*1)
		binFileBucketPath := envvar.GetString("BIN_FILE_BUCKET_PATH", "gs://happy_path_testing")

		ticker := time.NewTicker(binFileRefreshInterval)

		for {
			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				// Download and Verify
				for fileType, file := range files {

					config := file.Config

					fullArtifactPath := fmt.Sprintf("%s/%s", binFileBucketPath, config.Name)

					// Download from GCP bucket (Portal/Next tool uploads to bucket after creation)
					if err := service.GcpStorage.CopyFromBucketToLocal(service.Context, fullArtifactPath, config.VMFilePath); err != nil {
						core.Error("failed to download %s file: %v", fileType, err)
						continue
					}

					binFile, err := os.Open(config.VMFilePath)
					if err != nil {
						core.Error("failed to open database file")
						continue
					}
					defer binFile.Close()

					// Validate bin file
					// todo: figure out how to avoid this like location file
					switch fileType {
					case DATABASE:
						if err := common.ValidateDatabaseFile(binFile, &routing.DatabaseBinWrapper{}); err != nil {
							core.Error("failed to validate database file: %v", err)
							continue
						}
						break
					case OVERLAY:
						if err := common.ValidateOverlayFile(binFile, &routing.OverlayBinWrapper{}); err != nil {
							core.Error("failed to validate overlay file: %v", err)
							continue
						}
						break
					default:
						continue
					}

					// Don't upload unless leader VM
					if !service.IsLeader() {
						continue
					}

					// Upload file directly to VMs to update them
					// Upload == local file path here
					if err := service.UploadFileToGCPVirtualMachines(config.VMFilePath, config.VMFilePath, config.UploadVMs); err != nil {
						core.Error("failed to upload %s file to GCP VMs: %v", fileType, err)
					}
				}
			}
		}
	}()
}
