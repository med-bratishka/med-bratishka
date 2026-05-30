package s3

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, objectKey string, body []byte, contentType string) (string, error)
	Delete(ctx context.Context, objectKey string) error
}

type Client struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func New(endpoint, publicBaseURL, region, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
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

	if publicBaseURL == "" {
		scheme := "https"
		if !useSSL {
			scheme = "http"
		}
		publicBaseURL = fmt.Sprintf("%s://%s", scheme, endpoint)
	}

	return &Client{client: mc, bucket: bucket, publicBaseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

func (c *Client) Upload(ctx context.Context, objectKey string, body []byte, contentType string) (string, error) {
	if len(body) == 0 {
		return "", fmt.Errorf("empty payload")
	}

	_, err := c.client.PutObject(ctx, c.bucket, objectKey, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	segments := strings.Split(strings.TrimPrefix(objectKey, "/"), "/")
	for i := range segments {
		segments[i] = url.PathEscape(segments[i])
	}
	return fmt.Sprintf("%s/%s/%s", c.publicBaseURL, url.PathEscape(c.bucket), strings.Join(segments, "/")), nil
}

func (c *Client) Delete(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return nil
	}
	if err := c.client.RemoveObject(ctx, c.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}
