package common

import (
	"fmt"
	"time"
	"context"
	"errors"
	"os"
	"strings"
	"net/http"

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

func StartFileSync(ctx context.Context, config *FileSyncConfig, googleCloudStorage *GoogleCloudStorage, isLeader func() bool) {

	googleProjectId := googleCloudStorage.ProjectId

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

						if err := DownloadFile(ctx, googleCloudStorage, syncFile.DownloadURL, syncFile.Name); err != nil {
							core.Error("failed to download %s: %v", syncFile.Name, err)
							continue
						}
					}

					if !group.ValidationFunc(fileNames) {
						core.Error("failed to validate files: %v", fileNames)
						continue
					}

					for _, fileName := range fileNames {

						if group.PushTo != "" {
							if err := googleCloudStorage.CopyFromLocalToBucket(ctx, fileName, fmt.Sprintf("%s/%s", group.UploadTo, fileName)); err != nil {
								core.Error("failed to upload location file to google cloud storage: %v", err)
								continue
							}
						}

						receivingVMs := GetMIGInstanceNames(googleProjectId, group.PushTo)

						if len(receivingVMs) > 0 {
							core.Debug("pushing %s to VMs: %v", fileName, receivingVMs)
							if err := PushFileToVMs(ctx, googleCloudStorage, fileName, receivingVMs); err != nil {
								core.Error("failed to upload location file to google cloud VMs: %v", err)
							}
						}
					}
				}
			}
		}(group)
	}
}

func DownloadFile(ctx context.Context, googleCloudStorage *GoogleCloudStorage, downloadURL string, fileName string) error {

	currentDirectory, err := os.Getwd()
	if err != nil {
		core.Error("failed to get current directory: %v", err)
		currentDirectory = "./"
	}

	path := fmt.Sprintf("%s/%s", currentDirectory, fileName)

	urlScheme := strings.Split(downloadURL, "://")[0]

	switch urlScheme {
	case "gs":
		return googleCloudStorage.CopyFromBucketToLocal(ctx, downloadURL, path)
	case "http", "https":
		return DownloadFileFromURL(ctx, downloadURL, path)
	default:
		return errors.New(fmt.Sprintf("unknown url scheme: %s", urlScheme))
	}
}

func DownloadFileFromURL(ctx context.Context, downloadURL string, filePath string) error {

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	// todo: should we do n retries here? probably yes!
	_ = ctx

	httpResponse, err := httpClient.Get(downloadURL)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http call returned status code: %s", httpResponse.Status))
	}

	contentType := httpResponse.Header.Get("content-type")

	if contentType == "application/gzip" {

		if err := ExtractFileFromGZIP(httpResponse.Body, filePath); err != nil {
			return err
		}
	} else {
		if err := SaveBytesToFile(httpResponse.Body, filePath); err != nil {
			return err
		}
	}

	return nil
}

func PushFileToVMs(ctx context.Context, googleCloudStorage *GoogleCloudStorage, filePath string, vmNames []string) error {

	if len(vmNames) == 0 {
		return nil
	}

	hadError := false
	for _, vm := range vmNames {
		if err := googleCloudStorage.CopyFromLocalToRemote(ctx, filePath, filePath, vm); err != nil {
			core.Error("failed to copy file to vm: %v", err)
			hadError = true
		}
	}

	if hadError {
		return errors.New("failed to upload file to one or more vms")
	}

	return nil
}
