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

func main() {

	service := common.CreateService("pusher")

	// todo: it feels like a mistake to have "common.MaxmindFile" and "common.BinFile" types
	// This means that in the future, somebody who has to add a new "type" of file to be pushed
	// would need to 
	locationFiles := []common.MaxmindFile{
		{
			Name:        envvar.GetString("MAXMIND_ISP_FILE_NAME", "GeoIP2-ISP.mmdb"),
			Path:        envvar.GetString("MAXMIND_ISP_DOWNLOAD_PATH", "./GeoIP2-ISP.mmdb"),
			DownloadURL: envvar.GetString("MAXMIND_ISP_DB_URI", ""),
			Type:        common.MAXMIND_ISP,
			// todo: passing in env here as part of the configuration seems overkill, we typically know env, and can get it anytime we want from Service
			Env:         service.Env,
		},
		{
			Name:        envvar.GetString("MAXMIND_CITY_FILE_NAME", "GeoIP2-CITY.mmdb"),
			Path:        envvar.GetString("MAXMIND_CITY_DOWNLOAD_PATH", "./GeoIP2-City.mmdb"),
			// todo: the only difference with MaxMindFile and BinFile appears to be this "DownloadURL" member
			// why do we need two different types?
			DownloadURL: envvar.GetString("MAXMIND_CITY_DB_URI", ""),
			Type:        common.MAXMIND_CITY,
			Env:         service.Env,
		},
	}

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

	// todo: ideally we would print the configuration here with core.Log, in some clean and easy to understand way

	// todo: what if we had a more generic configuration, eg. config.FileGroup. It's an array of files, with all the information 
	// we need to download these files, and where they should go, and perhaps it has a func pointer to a validation function for the group
	// that takes the array of downloaded files (once the whole download group finishes), and runs validate on it, and then if validate
	// returns true, then those files are passed on to the push channel, which does the actual push to MIG VMs (same for all types)

	// todo: Ideally, we would have just a single configuration struct that could be passed in to a method in this service
	// that pumped the download groups, called verify, and then pushed to a channel.

	// todo: then we could have a second pusher goroutine, that listens on teh channel, and takes care of pushing file groups
	// to MIG VMs, when they come in over the channel.

	service.SetupGCPStorage()

	service.LeaderElection()

	RefreshLocationFiles(service, locationFiles)

	RefreshBinFiles(service, binFiles)

	service.StartWebServer()

	service.WaitForShutdown()
}

func RefreshLocationFiles(service *common.Service, locationFiles []common.MaxmindFile) {

	// todo: refresh interval and other things necessary for each "file group" should be in the CONFIG STRUCT
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

				// todo: commenting below here is generally redundant and does not add much

				// Download loop
				for _, locationFile := range locationFiles {

					// Toggle to avoid excess downloads - happy path - local testing
					if locationFile.DownloadURL == "" {
						continue
					}

					// ^------ todo: ideally, we would run ALL code locally in the happy path, with some special case configuration
					// so it downloads from HTTP some static files we put in gs

					fileExtensionTokens := strings.Split(locationFile.Path, ".")
					fileExtension := fileExtensionTokens[len(fileExtensionTokens)-1]

					// Download the file to local storage
					if err := service.DownloadGzipFileFromURL(locationFile.DownloadURL, locationFile.Path, fileExtension); err != nil {
						core.Error("failed to download %s file: %v", locationFile.Type, err)
						continue
					}

					// Validate the file
					// todo: pass in a validate function ptr to the config. then run it. 
					// this way we can add more types of file groups being downloaded,
					// with tha custom step that decides if they are valid
					if err := locationFile.Validate(service.Context); err != nil {
						core.Error("failed to validate %s file: %v", locationFile.Type, err)
						continue
					}

					// Don't upload unless leader VM
					// ^--- good example of a totally redundant comment
					// but in fact, we shouldn't even RUN this loop if we are not leader
					// there is no point if we are not leader in downloading the files, it is just wasteful.
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
					// ^--- todo: i dislike the terminology of "upload" here. upload to GS is fine
					// but in reality here what we are doing is PUSHING files to VMs. this is the "pusher"
					// service. we should own the fact that this is the PUSH
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

		// todo: this is basically a total duplication of RefreshLocationFiles with extremely minor changes
		// every change in this function vs. the location function should instead be expressed in CONFIG
		// not code. Make this a data driven system, where you can add new file groups with custom func pointer
		// for validate (the specific code for each type of thing), such that in the future it is easy for somebody
		// to come in here and go, OK, you know what, I have a new type of thing I need to push, and they can quickly
		// get it done, just by adding the new file group to the config, defining a new validation function, 
		// and doing NOTHING MORE

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
