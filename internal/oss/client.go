package oss

import (
	"fmt"
	"io"
	"os"
	"time"

	alioss "github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Client wraps the Aliyun OSS bucket for file uploads.
type Client struct {
	bucket   *alioss.Bucket
	basePath string
	domain   string
}

// NewClient initializes an OSS client from environment variables.
func NewClient() (*Client, error) {
	accessKeyID := os.Getenv("OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("OSS_ACCESS_KEY_SECRET")
	endpoint := os.Getenv("OSS_ENDPOINT")
	bucketName := os.Getenv("OSS_BUCKET_NAME")
	basePath := os.Getenv("OSS_BASE_PATH")
	domain := os.Getenv("OSS_DOMAIN")

	if accessKeyID == "" || accessKeySecret == "" || endpoint == "" || bucketName == "" {
		return nil, fmt.Errorf("OSS config incomplete: check OSS_ACCESS_KEY_ID, OSS_ACCESS_KEY_SECRET, OSS_ENDPOINT, OSS_BUCKET_NAME")
	}

	client, err := alioss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create oss client: %w", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("get oss bucket: %w", err)
	}

	return &Client{bucket: bucket, basePath: basePath, domain: domain}, nil
}

// UploadResult holds the result of a successful upload.
type UploadResult struct {
	URL string
	Key string
}

// Upload streams a reader to OSS under a date-based path.
func (c *Client) Upload(r io.Reader, filename string) (*UploadResult, error) {
	datePath := time.Now().Format("2006/01/02")
	objectKey := fmt.Sprintf("%s/%s/%s", c.basePath, datePath, filename)

	if err := c.bucket.PutObject(objectKey, r); err != nil {
		return nil, fmt.Errorf("oss put object: %w", err)
	}

	return &UploadResult{URL: fmt.Sprintf("%s/%s", c.domain, objectKey), Key: fmt.Sprintf("%s/%s", datePath, filename)}, nil
}
