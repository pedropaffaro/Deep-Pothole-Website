package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Options struct {
	Bucket        string
	Region        string
	Endpoint      string
	AccessKey     string
	SecretKey     string
	PublicBaseURL string
}

type Storage struct {
	uploader      *manager.Uploader
	bucket        string
	publicBaseURL string
}

func New(ctx context.Context, opts Options) (*Storage, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(opts.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("carregar configuração da AWS: %w", err)
	}

	client := awss3.NewFromConfig(cfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(opts.Endpoint)
		o.UsePathStyle = true
	})

	return &Storage{
		uploader:      manager.NewUploader(client),
		bucket:        opts.Bucket,
		publicBaseURL: strings.TrimRight(opts.PublicBaseURL, "/"),
	}, nil
}

func (s *Storage) Upload(ctx context.Context, key, contentType string, r io.Reader) error {
	in := &awss3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	}
	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}

	if _, err := s.uploader.Upload(ctx, in); err != nil {
		return fmt.Errorf("enviar %q para o bucket %q: %w", key, s.bucket, err)
	}
	return nil
}

func (s *Storage) URL(key string) string {
	if key == "" {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(key, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return s.publicBaseURL + "/" + strings.Join(parts, "/")
}
