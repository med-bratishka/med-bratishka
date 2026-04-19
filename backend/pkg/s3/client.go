package s3

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, objectKey string, body []byte, contentType string) (string, error)
}

type Client struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func New(endpoint, region, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("s3 config is incomplete")
	}

	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &Client{client: mc, bucket: bucket, endpoint: endpoint, useSSL: useSSL}, nil
}

func (c *Client) Upload(ctx context.Context, objectKey string, body []byte, contentType string) (string, error) {
	if len(body) == 0 {
		return "", fmt.Errorf("empty payload")
	}

	_, err := c.client.PutObject(ctx, c.bucket, objectKey, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	scheme := "https"
	if !c.useSSL {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, c.endpoint, c.bucket, path.Clean(strings.TrimPrefix(objectKey, "/"))), nil
}
