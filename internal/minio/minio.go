package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type MinioClient struct {
	client *minio.Client
	bucket string
	logger *zap.Logger
}

func NewMinioClient(endpoint, accessKey, secretKey, bucket string, logger *zap.Logger) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("Created MinIO bucket", zap.String("bucket", bucket))
	}

	return &MinioClient{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

func (m *MinioClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		m.logger.Error("Failed to upload file to MinIO",
			zap.String("object_name", objectName),
			zap.Error(err))
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", m.client.EndpointURL().Host, m.bucket, objectName)
	m.logger.Info("File uploaded successfully",
		zap.String("object_name", objectName),
		zap.String("url", url))
	return url, nil
}

func (m *MinioClient) GetFile(ctx context.Context, objectName string) (io.Reader, error) {
	object, err := m.client.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		m.logger.Error("Failed to get file from MinIO",
			zap.String("object_name", objectName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return object, nil
}
