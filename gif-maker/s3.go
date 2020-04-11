package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

const (
	timeFormat = "2006-01-02-15-04-05"
)

// Client represents the wrapper of S3 API Client
type Client struct {
	api s3iface.S3API
}

// New creates new Client
func newS3Client(api s3iface.S3API) *Client {
	return &Client{
		api: api,
	}
}

// ListObjectKeys retrieves the list of keys in the given S3 bucket and folder
func (c *Client) ListObjectKeys(ctx context.Context, bucket, folder string) ([]string, error) {
	keys := []string{}

	err := c.api.ListObjectsV2PagesWithContext(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(folder),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, c := range page.Contents {
			keys = append(keys, aws.StringValue(c.Key))
		}

		return true
	})
	if err != nil {
		return []string{}, fmt.Errorf("cannot retrieve object list from S3 (bucket: %q, folder: %q): %w", bucket, folder, err)
	}

	return keys, nil
}

// {prefix}/2006/01/02/
func composeFolder(prefix string, year, month, day int) string {
	return filepath.Join(prefix, fmt.Sprintf("%04d/%02d/%02d", year, month, day)) + "/"
}
