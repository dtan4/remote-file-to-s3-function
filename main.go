package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-xray-sdk-go/xray"
)

const (
	defaultTimezone = "UTC"
)

func HandleRequest(ctx context.Context) error {
	url := os.Getenv("URL")
	bucket := os.Getenv("BUCKET")
	keyPrefix := os.Getenv("KEY_PREFIX")

	timezone := os.Getenv("TIMEZONE")
	if timezone == "" {
		timezone = defaultTimezone
	}

	log.Printf("url: %q", url)
	log.Printf("bucket: %q", bucket)
	log.Printf("key prefix: %q", keyPrefix)
	log.Printf("timezone: %q", timezone)

	return do(ctx, url, bucket, keyPrefix, timezone)
}

func main() {
	lambda.Start(HandleRequest)
}

func do(ctx context.Context, url, bucket, keyPrefix, timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("cannot retrieve timezone %q; %w", timezone, err)
	}

	httpClient := xray.Client(&http.Client{
		Timeout: 5 * time.Second,
	})

	log.Printf("downloading %s", url)

	body, ext, err := download(ctx, httpClient, url)
	if err != nil {
		return fmt.Errorf("cannot download file from %q: %w", url, err)
	}

	sess := session.New()
	api := s3.New(sess)
	xray.AWS(api.Client)
	s3Client := newS3Client(api)

	now := time.Now().In(loc)
	key := composeKey(keyPrefix, now, ext)

	log.Printf("uploading to bucket: %s key: %s", bucket, key)

	if err := s3Client.UploadToS3(ctx, bucket, key, bytes.NewReader(body)); err != nil {
		return fmt.Errorf("cannot upload downloaded file to S3 (bucket: %q, key: %q): %w", bucket, key, err)
	}

	return nil
}
