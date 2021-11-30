package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/networknext/backend/modules/core"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCPStorage struct {
	Client *storage.Client
	Bucket *storage.BucketHandle
}

type GCPStorageError struct {
	err error
}

func (g *GCPStorageError) Error() string {
	return fmt.Sprintf("unknown GCP storage error: %v", g.err)
}

// Create new GCPStorage client
func NewGCPStorageClient(ctx context.Context, bucketName string, opts ...option.ClientOption) (*GCPStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if bucketName == "" {
		err := fmt.Errorf("NewGCPStorageClient() bucket name is empty or not defined")
		return nil, err
	}
	bucket := client.Bucket(bucketName)

	return &GCPStorage{
		Client: client,
		Bucket: bucket,
	}, nil
}

func (g *GCPStorage) CopyFromBytesToStorage(ctx context.Context, inputBytes []byte, outputFileName string) error {
	// Create an object writer
	writer := g.Bucket.Object(outputFileName).NewWriter(ctx)

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

func (g *GCPStorage) CopyFromBytesToRemote(inputBytes []byte, instanceNames []string, outputFileName string) error {
	// Check if there is a temp file with the same name already
	if LocalFileExists(outputFileName) {
		// delete temp file
		if err := os.Remove(outputFileName); err != nil {
			err = fmt.Errorf("failed to remove existing output file: %v", err)
			return err
		}
	}

	// Write bytes to temp file location
	if err := ioutil.WriteFile(outputFileName, inputBytes, 0644); err != nil {
		err = fmt.Errorf("failed to write to temp file: %v", err)
		return err
	}

	// If we error somewhere in the loop, don't exit until all instances have been tried (debug instance isn't live)
	var loopError error
	for _, name := range instanceNames {
		// Call gsutil to copy the tmp file over to the instance
		runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", outputFileName, fmt.Sprintf("%s:%s", name, outputFileName))

		buffer, err := runnable.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("failed to copy file to instance %s: %v", name, err)
			loopError = err
		} else {
			core.Debug("CopyFromBytesToRemote() gcloud compute scp output: %s", buffer)
		}
	}

	if loopError != nil {
		return loopError
	}
	// delete temp file to clean up
	if err := os.Remove(outputFileName); err != nil {
		err = fmt.Errorf("failed to clean up tmp file: %v", err)
		return err
	}

	return nil
}

// Copy file from local to GCP Storage Bucket
func (g *GCPStorage) CopyFromLocalToBucket(ctx context.Context, outputLocation string, artifactBucketName string, storageFileName string) error {
	if !LocalFileExists(outputLocation) {
		err := fmt.Errorf("local file %s does not exist", outputLocation)
		return err
	}

	bucketPath := fmt.Sprintf("gs://%s/%s", artifactBucketName, storageFileName)

	runnable := exec.Command("gsutil", "cp", outputLocation, bucketPath)
	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to bucket: %v", err)
		return err
	}
	core.Debug("CopyFromLocalToBucket() gsutil cp output: %s", buffer)

	return nil
}

// Copy file from local to remote location on VM (SCP function)
func (g *GCPStorage) CopyFromLocalToRemote(ctx context.Context, outputFileName string, instanceName string) error {

	// Call gsutil to copy the tmp file over to the instance
	runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", outputFileName, fmt.Sprintf("%s:%s", instanceName, outputFileName))

	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to instance %s: %v", instanceName, err)
		return err
	}

	core.Debug("CopyFromLocalToRemote() gcloud compute scp output: %s", buffer)

	return nil
}

// Copy artifact to local a local file location
func (g *GCPStorage) CopyFromBucketToLocal(ctx context.Context, artifactName string, outputLocation string) error {
	if LocalFileExists(outputLocation) {
		// delete output file
		if err := os.Remove(outputLocation); err != nil {
			err = fmt.Errorf("failed to remove existing output file: %v", err)
			return err
		}
	}

	runnable := exec.Command("gsutil", "cp", artifactName, outputLocation)
	buffer, err := runnable.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to copy file to instance: %v", err)
		return err
	}
	core.Debug("CopyFromBucketToLocal() gsutil cp output: %s", buffer)

	return nil
}

// Copy artifact to remote file location (SCP function)
func (g *GCPStorage) CopyFromBucketToRemote(ctx context.Context, artifactName string, instanceNames []string, outputFileName string) error {
	// grab file from bucket
	if err := g.CopyFromBucketToLocal(ctx, artifactName, outputFileName); err != nil {
		err = fmt.Errorf("failed to fetch file from GCP bucket: %v", err)
		return err
	}

	// If we error somewhere in the loop, don't exit until all instances have been tried (debug instance isn't live)
	var loopError error
	for _, name := range instanceNames {
		// Call gsutil to copy the tmp file over to the instance
		runnable := exec.Command("gcloud", "compute", "scp", "--zone", "us-central1-a", outputFileName, fmt.Sprintf("%s:%s", name, outputFileName))

		buffer, err := runnable.CombinedOutput()

		if err != nil {
			err = fmt.Errorf("failed to copy file to instance %s: %v", name, err)
			loopError = err
		} else {
			core.Debug("CopyFromBucketToRemote() gcloud compute scp output: %s", buffer)
		}
	}

	if loopError != nil {
		return loopError
	}

	// delete src file to clean up
	if err := os.Remove(outputFileName); err != nil {
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
