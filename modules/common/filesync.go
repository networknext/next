package common

import (
	// "fmt"
	"time"
	"context"

	"github.com/networknext/backend/modules/core"
)

type FileSyncConfig struct {
	FileGroups []FileSyncGroup
}
type FileSyncGroup struct {
	Name           string
	Files          []SyncFile
	SyncInterval   time.Duration
	ValidationFunc func([]string) bool
	UploadTo       string
	PushTo         string
}

type SyncFile struct {
	Name        string
	DownloadURL string
}

func (config *FileSyncConfig) Print() {
	core.Log("file sync groups:")
	for index, group := range config.FileGroups {
		core.Log("%d: %s [%s]", index, group.Name, group.SyncInterval.String())
		for _, file := range group.Files {
			core.Log(" + %s -> %s ", file.DownloadURL, file.Name)
		}
	}
}

func StartFileSync(ctx context.Context, config *FileSyncConfig, googleProjectId string, isLeader func() bool) {

	for _, group := range config.FileGroups {

		go func(group FileSyncGroup) {

			ticker := time.NewTicker(group.SyncInterval)

			for {
				select {

				case <-ctx.Done():
					return

				case <-ticker.C:

					if !isLeader() {
						continue
					}

					core.Debug("syncing %s", group.Name)

					fileNames := make([]string, len(group.Files))

					for i, syncFile := range group.Files {

						fileNames[i] = syncFile.Name

						core.Debug("downloading %s", syncFile.DownloadURL)
						// todo
						/*
						if err := service.DownloadFile(syncFile.DownloadURL, syncFile.Name); err != nil {
							core.Error("failed to download %s: %v", syncFile.Name, err)
							continue
						}
						*/
					}

					if !group.ValidationFunc(fileNames) {
						core.Error("failed to validate files: %v", fileNames)
						continue
					}

					for _, fileName := range fileNames {

						// todo
						/*
						if group.SaveBucket != "" {
							if err := service.googleCloudStorage.CopyFromLocalToBucket(service.Context, fileName, fmt.Sprintf("%s/%s", group.SaveBucket, fileName)); err != nil {
								core.Error("failed to upload location file to google cloud storage: %v", err)
								continue
							}
						}
						*/

						receivingVMs := GetMIGInstanceNames(googleProjectId, group.PushTo)
						if len(receivingVMs) > 0 {
							core.Debug("pushing %s to VMs: %v", fileName, receivingVMs)
							/*
							if err := service.PushFileToVMs(fileName, receivingVMs); err != nil {
								core.Error("failed to upload location file to google cloud VMs: %v", err)
							}
							*/
						}
					}
				}
			}
		}(group)
	}
}
