package common

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/networknext/backend/modules/core"
)

const (
	MAX_RETRY_COUNT = 5
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
	PushTo         []string
	OutputPath     string
}

type SyncFile struct {
	Name        string
	DownloadURL string
}

func CreateFileSyncConfig() *FileSyncConfig {
	return &FileSyncConfig{
		FileGroups: make([]FileSyncGroup, 0),
	}
}

func (config *FileSyncConfig) AddFileSyncGroup(groupName string, syncInterval time.Duration, migNames []string, outputPath string, uploadBucketURL string, validationFunc func([]string) bool, files ...SyncFile) {
	config.FileGroups = append(config.FileGroups, FileSyncGroup{
		Name:           groupName,
		SyncInterval:   syncInterval,
		PushTo:         migNames,
		ValidationFunc: validationFunc,
		UploadTo:       uploadBucketURL,
		Files:          files,
		OutputPath:     outputPath,
	})
}

func (config *FileSyncConfig) Print() {
	core.Log("----------------------------------------------------------")
	core.Log("file sync groups:")
	for index, group := range config.FileGroups {
		core.Log("%d: %s [%s]", index, group.Name, group.SyncInterval.String())
		for _, file := range group.Files {
			fileName := file.Name

			if fileName == "" {
				fileName = GetFileNameFromPath(file.DownloadURL)
			}
			core.Log(" + %s -> %s ", file.DownloadURL, fmt.Sprintf("%s%s", group.OutputPath, fileName))
		}
	}
	core.Log("----------------------------------------------------------")
}

func StartFileSync(ctx context.Context, config *FileSyncConfig, googleCloudHandler *GoogleCloudHandler, isLeader func() bool) {

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

						if syncFile.DownloadURL == "" {
							core.Debug("no download url specified: %s", syncFile.Name)
							continue
						}

						fileName := syncFile.Name

						if fileName == "" {
							fileName = GetFileNameFromPath(syncFile.DownloadURL)
						}

						outputPath := fmt.Sprintf("%s%s", group.OutputPath, fileName)

						core.Debug("downloading %s", syncFile.DownloadURL)

						if err := DownloadFile(ctx, googleCloudHandler, syncFile.DownloadURL, fileName, outputPath); err != nil {
							core.Error("failed to download %s: %v", fileName, err)
							continue
						}

						fileNames[i] = fileName
					}

					if !group.ValidationFunc(fileNames) {
						core.Error("failed to validate files: %v", fileNames)
						continue
					}

					for _, fileName := range fileNames {

						outputPath := fmt.Sprintf("%s%s", group.OutputPath, fileName)

						if group.UploadTo != "" {
							core.Debug("uploading files to: %s", group.UploadTo)
							if err := googleCloudHandler.CopyFromLocalToBucket(ctx, fileName, fmt.Sprintf("%s/%s", group.UploadTo, fileName)); err != nil {
								core.Error("failed to upload location file to google cloud storage: %v", err)
								continue
							}
						}

						for _, destination := range group.PushTo {

							receivingVMs := googleCloudHandler.GetMIGInstanceNames(destination)

							core.Debug("pushing %s to VMs: %v", fileName, receivingVMs)

							if err := PushFileToVMs(ctx, googleCloudHandler, fileName, outputPath, receivingVMs); err != nil {
								core.Error("failed to upload location file to google cloud VMs: %v", err)
							}

						}
					}
				}
			}
		}(group)
	}
}

func DownloadFile(ctx context.Context, googleCloudHandler *GoogleCloudHandler, downloadURL string, fileName string, outputPath string) error {

	urlScheme := strings.Split(downloadURL, "://")[0]

	switch urlScheme {
	case "gs":
		return googleCloudHandler.CopyFromBucketToLocal(ctx, downloadURL, outputPath)
	case "http", "https":
		return DownloadFileFromURL(ctx, downloadURL, outputPath)
	default:
		return errors.New(fmt.Sprintf("unknown url scheme: %s", urlScheme))
	}

}

func DownloadFileFromURL(ctx context.Context, downloadURL string, filePath string) error {

	var httpResponse *http.Response
	var err error

	retryCount := 0

	httpClient := http.Client{
		Timeout: time.Second * 30,
	}

	for retryCount < MAX_RETRY_COUNT {
		httpResponse, err = httpClient.Get(downloadURL)
		if err != nil {
			core.Debug("http get request failed: %v", err)
			continue
		}

		if httpResponse.StatusCode == http.StatusOK {
			break
		}

		retryCount = retryCount + 1

		if retryCount == MAX_RETRY_COUNT {
			continue
		}

		backOff := float64(time.Second) / (2 * math.Exp(float64(retryCount)))

		time.Sleep(time.Duration(backOff))
	}

	if retryCount == MAX_RETRY_COUNT {
		return errors.New(fmt.Sprintf("http call hit max retries"))
	}

	defer httpResponse.Body.Close()

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

func PushFileToVMs(ctx context.Context, googleCloudHandler *GoogleCloudHandler, filePath string, outputPath string, vmNames []string) error {

	if len(vmNames) == 0 {
		return nil
	}

	hadError := false
	for _, vm := range vmNames {
		if err := googleCloudHandler.CopyFromLocalToRemote(ctx, filePath, outputPath, vm); err != nil {
			core.Error("failed to copy file to vm: %v", err)
			hadError = true
		}
	}

	if hadError {
		return errors.New("failed to upload file to one or more vms")
	}

	return nil
}
