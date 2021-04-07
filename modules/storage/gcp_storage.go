package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCPStorage struct {
	Client *storage.Client
	Logger log.Logger
	Bucket *storage.BucketHandle
}

type GCPStorageError struct {
	err error
}

func (g *GCPStorageError) Error() string {
	return fmt.Sprintf("unknown GCP storage error: %v", g.err)
}

// Create new GCPStorage client
func NewGCPStorageClient(ctx context.Context, bucketName string, logger log.Logger, opts ...option.ClientOption) (*GCPStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	if bucketName == "" {
		err := fmt.Errorf("NewGCPStorageClient() bucket name is empty or not defined")
		level.Error(logger).Log("err", err)
		return nil, err
	}
	bucket := client.Bucket(bucketName)

	return &GCPStorage{
		Client: client,
		Logger: logger,
		Bucket: bucket,
	}, nil
}

func (g *GCPStorage) CopyFromBytesToStorage(ctx context.Context, inputBytes []byte, outputFileName string) error {
	// Create an object writer
	writer := g.Bucket.Object(outputFileName).NewWriter(ctx)

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

// Copy artifact to local a local file location
func (g *GCPStorage) CopyFromBucketToLocal(ctx context.Context, artifactName string, outputLocation string) error {
	if LocalFileExists(outputLocation) {
		// delete output file
		if err := os.Remove(outputLocation); err != nil {
			err = fmt.Errorf("failed to remove existing output file: %v", err)
			return err
		}
	}

	// setup reader
	reader, err := g.Bucket.Object(artifactName).NewReader(ctx)
	if err != nil {
		err = fmt.Errorf("failed to create bucket reader: %v", err)
		return err
	}

	defer reader.Close()

	// read in artifact as bytes to be copied to new file
	var artifactBytes []byte
	_, err = reader.Read(artifactBytes)
	if err != nil {
		err = fmt.Errorf("failed to read artifact: %v", err)
		return err
	}

	if err := ioutil.WriteFile(outputLocation, artifactBytes, 0644); err != nil {
		err = fmt.Errorf("failed to write local file: %v", err)
		return err
	}

	return nil
}

// Copy artifact to remote file location (SCP function)
func (g *GCPStorage) CopyFromBucketToRemote(ctx context.Context, artifactName string, srcFileName string, instanceName string, outputFileName string) error {
	// grab file from bucket
	if err := g.CopyFromBucketToLocal(ctx, artifactName, srcFileName); err != nil {
		err = fmt.Errorf("failed to fetch file from GCP bucket: %v", err)
		return err
	}

	// Call gsutil to copy the tmp file over to the instance
	command := fmt.Sprintf("gsutil compute scp %s %s:%s", srcFileName, instanceName, outputFileName)
	runnable := exec.Command(command)

	if err := runnable.Run(); err != nil {
		err = fmt.Errorf("failed to copy file to instance: %v", err)
		return err
	}

	// delete src file to clean up
	if err := os.Remove(srcFileName); err != nil {
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
