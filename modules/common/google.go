package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCPHandler struct {
	ProjectId        string
	storageClient    *storage.Client
	StorageBucketURL string
	storageBucket    *storage.BucketHandle
}

type GCPHandlerError struct {
	err error
}

func (g *GCPHandlerError) Error() string {
	return fmt.Sprintf("unknown GCP handler error: %v", g.err)
}

// Create new GCPHandler
func NewGCPHandler(ctx context.Context, projectId string, bucketURL string, opts ...option.ClientOption) (*GCPHandler, error) {
	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if bucketURL == "" {
		err := fmt.Errorf("NewGCPHandler() bucket name is empty or not defined")
		return nil, err
	}

	bucketNameTokens := strings.Split(bucketURL, "/")
	bucketName := bucketNameTokens[len(bucketNameTokens)-1]

	return &GCPHandler{
		ProjectId:        projectId,
		storageClient:    storageClient,
		StorageBucketURL: bucketURL,
		storageBucket:    storageClient.Bucket(bucketName),
	}, nil
}

func (g *GCPHandler) CopyFromBytesToStorage(ctx context.Context, inputBytes []byte, outputPath string) error {
	// Create an object writer
	writer := g.storageBucket.Object(outputPath).NewWriter(ctx)

	writer.ObjectAttrs.ContentType = "application/octet-stream"

	// Write to the file
	if _, err := writer.Write(inputBytes); err != nil {
		err = fmt.Errorf("failed to write to bucket object: %v", err)
		return err
	}

	if err := writer.Close(); err != nil {
		err = fmt.Errorf("failed to write to bucket object: %v", err)
		return err
	}

	return nil
}

func (g *GCPHandler) CopyFromBytesToRemote(inputBytes []byte, instanceNames []string, outputPath string) error {
	// Check if there is a temp file with the same name already
	if LocalFileExists(outputPath) {
		// delete temp file
		if err := os.Remove(outputPath); err != nil {
			err = fmt.Errorf("failed to remove existing output file: %v", err)
			return err
		}
	}

	// Write bytes to temp file location
	if err := ioutil.WriteFile(outputPath, inputBytes, 0644); err != nil {
		err = fmt.Errorf("failed to write to temp file: %v", err)
		return err
	}

	// If we error somewhere in the loop, don't exit until all instances have been tried (debug instance isn't live)
	var loopError error
	for _, name := range instanceNames {
		// Call gsutil to copy the tmp file over to the instance
		runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", outputPath, fmt.Sprintf("%s:%s", name, outputPath))

		buffer, err := runnable.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("failed to copy file to instance %s: %v", name, err)
			loopError = err
		}

		if string(buffer) != "" {
			core.Debug("CopyFromBytesToRemote() gcloud compute scp output: %s", buffer)
		}
	}

	if loopError != nil {
		return loopError
	}
	// delete temp file to clean up
	if err := os.Remove(outputPath); err != nil {
		err = fmt.Errorf("failed to clean up tmp file: %v", err)
		return err
	}

	return nil
}

// Copy file from local to GCP Storage Bucket
func (g *GCPHandler) CopyFromLocalToBucket(ctx context.Context, localPath string, storagePath string) error {
	if !LocalFileExists(localPath) {
		err := fmt.Errorf("local file %s does not exist", localPath)
		return err
	}

	runnable := exec.Command("gsutil", "cp", localPath, storagePath)
	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to bucket: %v", err)
		return err
	}

	if string(buffer) != "" {
		core.Debug("CopyFromLocalToBucket() gsutil cp output: %s", buffer)
	}

	return nil
}

// Copy file from local to remote location on VM (SCP function)
func (g *GCPHandler) CopyFromLocalToRemote(ctx context.Context, localPath string, outputPath string, instanceName string) error {

	// Call gsutil to copy the tmp file over to the instance
	runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", localPath, fmt.Sprintf("%s:%s", instanceName, outputPath))

	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to instance %s: %v", instanceName, err)
		return err
	}

	core.Debug("CopyFromLocalToRemote() gcloud compute scp output: %s", buffer)

	return nil
}

// Copy artifact to local a local file location
func (g *GCPHandler) CopyFromBucketToLocal(ctx context.Context, storagePath string, outputPath string) error {
	if LocalFileExists(outputPath) {
		// delete output file
		if err := os.Remove(outputPath); err != nil {
			err = fmt.Errorf("failed to remove existing output file: %v", err)
			return err
		}
	}

	runnable := exec.CommandContext(ctx, "gsutil", "cp", storagePath, outputPath)
	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to instance: %v", err)
		return err
	}

	if string(buffer) != "" {
		core.Debug("CopyFromBucketToLocal() gsutil cp output: %s", buffer)
	}

	return nil
}

// Copy artifact to remote file location (SCP function)
func (g *GCPHandler) CopyFromBucketToRemote(ctx context.Context, storagePath string, instanceNames []string, outputPath string) error {
	// grab file from bucket
	if err := g.CopyFromBucketToLocal(ctx, storagePath, outputPath); err != nil {
		err = fmt.Errorf("failed to fetch file from GCP bucket: %v", err)
		return err
	}

	// If we error somewhere in the loop, don't exit until all instances have been tried (debug instance isn't live)
	var loopError error
	for _, name := range instanceNames {
		// Call gsutil to copy the tmp file over to the instance
		runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", outputPath, fmt.Sprintf("%s:%s", name, outputPath))

		buffer, err := runnable.CombinedOutput()

		if err != nil {
			err = fmt.Errorf("failed to copy file to instance %s: %v", name, err)
			loopError = err
		}

		if string(buffer) != "" {
			core.Debug("CopyFromBucketToRemote() gcloud compute scp output: %s", buffer)
		}
	}

	if loopError != nil {
		return loopError
	}

	// delete src file to clean up
	if err := os.Remove(outputPath); err != nil {
		err = fmt.Errorf("failed to clean up tmp file: %v", err)
		return err
	}

	return nil
}

func LocalFileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
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
