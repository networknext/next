package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCPHandler struct {
	ProjectId     string
	storageClient *storage.Client
}

// Create new GCPHandler
func NewGCPHandler(ctx context.Context, projectId string, opts ...option.ClientOption) (*GCPHandler, error) {
	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GCPHandler{
		ProjectId:     projectId,
		storageClient: storageClient,
	}, nil
}

// Copy file from local to GCP Storage Bucket
func (g *GCPHandler) CopyFromLocalToBucket(ctx context.Context, fileName string, storagePath string) error {

	currentDirectory, err := os.Getwd()
	if err != nil {
		core.Error("failed to get current directory: %v", err)
		currentDirectory = "./"
	}

	path := fmt.Sprintf("%s/%s", currentDirectory, fileName)

	if !LocalFileExists(path) {
		return errors.New(fmt.Sprintf("local file %s does not exist", path))
	}

	runnable := exec.Command("gsutil", "cp", path, storagePath)
	buffer, err := runnable.CombinedOutput()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to copy file to bucket: %v", err))
	}

	if len(buffer) > 0 {
		core.Debug("gsutil cp output: %s", buffer)
	}

	return nil
}

// Copy file from local to remote location on VM (SCP function)
func (g *GCPHandler) CopyFromLocalToRemote(ctx context.Context, localPath string, outputPath string, instanceName string) error {

	// Call gsutil to copy the tmp file over to the instance
	runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", localPath, fmt.Sprintf("%s:%s", instanceName, outputPath))

	buffer, err := runnable.CombinedOutput()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to copy file to instance %s: %v", instanceName, err))
	}

	if len(buffer) > 0 {
		core.Debug("gcloud scp output: %s", buffer)
	}

	return nil
}

// Copy artifact to local a local file location
func (g *GCPHandler) CopyFromBucketToLocal(ctx context.Context, bucketURL string, outputPath string) error {

	artifactPath := strings.Trim(bucketURL, "gs://")
	pathTokens := strings.Split(artifactPath, "/")

	reader, err := g.storageClient.Bucket(pathTokens[0]).Object(pathTokens[1]).NewReader(ctx)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create bucket reader: %v", err))
	}
	defer reader.Close()

	if reader.ContentType() == "application/gzip" {
		if err := ExtractFileFromGZIP(reader, outputPath); err != nil {
			return errors.New(fmt.Sprintf("failed to extract file from gzip: %v", err))
		}
	} else {
		if err := SaveBytesToFile(reader, outputPath); err != nil {
			return errors.New(fmt.Sprintf("failed to write to file: %v", err))
		}
	}

	return nil
}

// -------------------------------------------------------

type InstanceInfo struct {
	CurrentAction  string
	Id             string
	InstanceStatus string
}

func (g *GCPHandler) GetMIGInstanceInfo(migName string) []InstanceInfo {

	if g.ProjectId == "local" {
		return make([]InstanceInfo, 0)
	}

	var instances []InstanceInfo

	// Get the latest instance list in the relay backend mig
	runnable := exec.Command("gcloud", "compute", "--project", g.ProjectId, "instance-groups", "managed", "list-instances", migName, "--zone", "us-central1-a", "--format", "json")

	instancesListJson, err := runnable.CombinedOutput()
	if err != nil {
		core.Error("failed to get instance list for mig %s: %v", migName, err)
		return instances
	}

	if err := json.Unmarshal([]byte(instancesListJson), &instances); err != nil {
		core.Error("failed to unmarshal instance list json: %v", err)
		return instances
	}

	return instances
}

func (g *GCPHandler) GetMIGInstanceNames(migName string) []string {
	instances := g.GetMIGInstanceInfo(migName)

	numInstances := len(instances)

	names := make([]string, numInstances)

	for i := 0; i < numInstances; i++ {
		names[i] = instances[i].Id
	}

	return names
}

func (g *GCPHandler) GetMIGInstanceNamesEnv(environmentVariable string, defaultValue string) []string {
	return g.GetMIGInstanceNames(envvar.GetString(environmentVariable, defaultValue))
}
