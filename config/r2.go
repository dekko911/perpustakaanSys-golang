package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Storage returns an *s3.Client configured to communicate with a
// Cloudflare R2-compatible endpoint. It loads AWS SDK v2 configuration using
// static credentials from CFAccessKeyID and CFSecretAccessKey and sets
// the region from CFDefaultRegion. Configuration is performed with
// a *http.Request context. The client's BaseEndpoint is set to baseURL R2 Cloudflare with AccountID.
// If configuration loading fails, the function logs the error and exits.
func R2Storage(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(Env.CFAccessKeyID, Env.CFSecretAccessKey, "")),
		config.WithRegion(Env.CFDefaultRegion),
		config.WithRetryMaxAttempts(10))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", Env.CFAccountID))

		// set default endpoint r2
		o.UsePathStyle = true

		// turn off auto checksum
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenSupported
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenSupported
	})

	return client, nil
}
