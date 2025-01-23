package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
	config S3StorageConfig
}

type S3StorageConfig struct {
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
}

func NewS3Storage(c S3StorageConfig) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(
		c.AccessKey,
		c.SecretKey,
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(c.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client: client,
		config: c,
	}, nil
}

func (s *S3Storage) Put(ctx context.Context, name string, file io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(name),
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %w", err)
	}

	return nil
}

func (s *S3Storage) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(name),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get file from S3: %w", err)
	}

	return output.Body, nil
}
