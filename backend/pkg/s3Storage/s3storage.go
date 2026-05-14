package s3storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
}

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewStorage(ctx context.Context, cfg Config) (*S3Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, err
	}

	return &S3Client{
		client: s3.NewFromConfig(awsCfg),
		bucket: cfg.Bucket,
	}, nil
}
