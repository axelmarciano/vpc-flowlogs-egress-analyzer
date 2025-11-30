package services

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"sync"
	"vpc_flowlogs_egress_analyzer/internal/config"
)

var (
	s3Client     *s3.Client
	initS3Client sync.Once
)

func GetS3Client() (*s3.Client, error) {
	var err error

	initS3Client.Do(func() {
		var cfg aws.Config
		opts := []func(*awsconfig.LoadOptions) error{
			awsconfig.WithRegion(config.GetEnv("AWS_REGION")),
		}
		accessKey := config.GetEnv("AWS_ACCESS_KEY_ID")
		secretKey := config.GetEnv("AWS_SECRET_ACCESS_KEY")
		if accessKey != "" && secretKey != "" {
			opts = append(opts, awsconfig.WithCredentialsProvider(
				aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
					return aws.Credentials{
						AccessKeyID:     accessKey,
						SecretAccessKey: secretKey,
					}, nil
				}),
			))
		}

		cfg, err = awsconfig.LoadDefaultConfig(context.TODO(), opts...)
		if err == nil {
			s3Client = s3.NewFromConfig(cfg)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration: %w", err)
	}
	return s3Client, nil
}
