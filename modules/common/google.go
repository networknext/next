package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// -----------------------------------------------------------------------

type GoogleCloudHandler struct {
	ProjectId     string
	storageClient *storage.Client
}

func NewGoogleCloudHandler(ctx context.Context, projectId string, opts ...option.ClientOption) (*GoogleCloudHandler, error) {

	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &GoogleCloudHandler{
		ProjectId:     projectId,
		storageClient: storageClient,
	}, nil
}

func (g *GoogleCloudHandler) CopyFromBucketToLocal(ctx context.Context, bucketURL string, outputPath string) error {

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

// -----------------------------------------------------------------------
