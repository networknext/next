/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/networknext/backend/modules-old/backend"
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

	serverBackendMIGName := envvar.GetString("SERVER_BACKEND_MIG_NAME", "")
	serverBackendInstances := service.GetMIGInstanceInfo(serverBackendMIGName)
	serverBackendInstanceNames := make([]string, len(serverBackendInstances))

	for i, instance := range serverBackendInstances {
		serverBackendInstanceNames[i] = instance.Id
	}

	relayGatewayMIGName := envvar.GetString("RELAY_GATEWAY_MIG_NAME", "")
	relayGatewayInstances := service.GetMIGInstanceInfo(relayGatewayMIGName)
	relayGatewayInstanceNames := make([]string, len(relayGatewayInstances))

	for i, instance := range relayGatewayInstances {
		relayGatewayInstanceNames[i] = instance.Id
	}

	locationFiles := make(map[string]LocationFile)
	locationFiles[ISP] = LocationFile{
		Config: FileConfig{
			Name:        envvar.GetString("MAXMIND_ISP_FILE_NAME", "GeoIP2-ISP.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_ISP_DB_URI", ""),
			VMFilePath:  envvar.GetString("MAXMIND_ISP_DOWNLOAD_PATH", "./GeoIP2-ISP.mmdb"),
			UploadVMs:   serverBackendInstanceNames,
		},
		ValidationFunc: validateISPFile,
	}

	locationFiles[CITY] = LocationFile{
		Config: FileConfig{
			Name:        envvar.GetString("MAXMIND_CITY_FILE_NAME", "GeoIP2-City.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_CITY_DB_URI", ""),
			VMFilePath:  envvar.GetString("MAXMIND_CITY_DOWNLOAD_PATH", "./GeoIP2-City.mmdb"),
			UploadVMs:   serverBackendInstanceNames,
		},
		ValidationFunc: validateCityFile,
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

func RefreshLocationFiles(service *common.Service, configs map[string]LocationFile) {

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
				for fileType, file := range configs {

					config := file.Config

					// Toggle to avoid excess downloads - happy path - local testing
					if config.DownloadURL == "" {
						continue
					}

					// Download the file to local storage
					if err := service.DownloadGzipFileFromURL(config.DownloadURL, config.VMFilePath); err != nil {
						core.Error("failed to download %s file: %v", fileType, err)
						continue
					}

					// Validate the file
					if err := file.ValidationFunc(service.Context, service.Env, config.VMFilePath); err != nil {
						core.Error("failed to validate %s file: %v", fileType, err)
						continue
					}
				}

				// Don't upload unless leader VM
				if !service.IsLeader() {
					continue
				}

				for _, file := range configs {
					config := file.Config

					fullArtifactPath := fmt.Sprintf("%s/%s", locationFileBucketPath, config.Name)

					// Upload file to GCP for VMs that are replacing (part of startup script)
					if err := service.UploadFileToGCPBucket(config.VMFilePath, fullArtifactPath); err != nil {
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
					if err := service.DownloadFileFromGCPBucket(fullArtifactPath, config.VMFilePath); err != nil {
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
						if err := validateDatabaseFile(binFile, &routing.DatabaseBinWrapper{}); err != nil {
							core.Error("failed to validate database file: %v", err)
							continue
						}
						break
					case OVERLAY:
						if err := validateOverlayFile(binFile, &routing.OverlayBinWrapper{}); err != nil {
							core.Error("failed to validate overlay file: %v", err)
							continue
						}
						break
					default:
						continue
					}
				}

				// Don't upload unless leader VM
				if !service.IsLeader() {
					continue
				}

				for fileType, file := range files {

					config := file.Config

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

// todo: move these somewhere better

func validateISPFile(ctx context.Context, env string, ispStorageName string) error {
	mmdb := &routing.MaxmindDB{
		IspFile:   ispStorageName,
		IsStaging: env == "staging",
	}

	// Validate the ISP file
	if err := mmdb.OpenISP(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateISP(); err != nil {
		return err
	}

	return nil
}

func validateCityFile(ctx context.Context, env string, cityStorageName string) error {
	mmdb := &routing.MaxmindDB{
		CityFile:  cityStorageName,
		IsStaging: env == "staging",
	}

	// Validate the City file
	if err := mmdb.OpenCity(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateCity(); err != nil {
		return err
	}

	return nil
}

func validateDatabaseFile(databaseFile *os.File, databaseNew *routing.DatabaseBinWrapper) error {
	if err := backend.DecodeBinWrapper(databaseFile, databaseNew); err != nil {
		core.Error("validateDatabaseFile() failed to decode database file: %v", err)
		return err
	}

	if databaseNew.IsEmpty() {
		// Don't want to use an empty bin wrapper
		// so early out here and use existing array and hash
		err := fmt.Errorf("new database file is empty, keeping previous values")
		core.Error(err.Error())
		return err
	}

	return nil
}

func validateOverlayFile(overlayFile *os.File, overlayNew *routing.OverlayBinWrapper) error {
	if err := backend.DecodeOverlayWrapper(overlayFile, overlayNew); err != nil {
		core.Error("validateOverlayFile() failed to decode database file: %v", err)
		return err
	}

	return nil
}
