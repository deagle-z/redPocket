package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	r2Once     sync.Once
	r2Uploader *manager.Uploader
	r2Err      error
)

func getR2Uploader(ctx context.Context) (*manager.Uploader, error) {
	r2Once.Do(func() {
		cfg := GlobalConfig.R2
		if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKeyId == "" || cfg.AccessKeySecret == "" {
			r2Err = errors.New("r2 config is missing")
			return
		}
		resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					URL:               cfg.Endpoint,
					SigningRegion:     "auto",
					HostnameImmutable: true,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		awsCfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion("auto"),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyId, cfg.AccessKeySecret, "")),
			config.WithEndpointResolverWithOptions(resolver),
		)
		if err != nil {
			r2Err = err
			return
		}
		client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.UsePathStyle = true
		})
		r2Uploader = manager.NewUploader(client)
	})
	return r2Uploader, r2Err
}

// UploadR2 uploads a multipart file to Cloudflare R2 and returns the public URL.
func UploadR2(ctx context.Context, fileHeader *multipart.FileHeader, keyPrefix string) (string, error) {
	if fileHeader == nil {
		return "", errors.New("file is nil")
	}
	uploader, err := getR2Uploader(ctx)
	if err != nil {
		return "", err
	}
	cfg := GlobalConfig.R2
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	key := buildR2ObjectKey(keyPrefix, fileHeader.Filename)
	contentType := fileHeader.Header.Get("Content-Type")
	input := &s3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	_, err = uploader.Upload(ctx, input)
	if err != nil {
		return "", err
	}
	return buildR2PublicURL(key), nil
}

func buildR2ObjectKey(prefix string, filename string) string {
	ext := filepath.Ext(filename)
	randomPart := RandomString(10)
	datePart := time.Now().Format("20060102/150405")
	base := strings.Trim(prefix, "/")
	if base == "" {
		return fmt.Sprintf("%s_%s%s", datePart, randomPart, ext)
	}
	return path.Join(base, fmt.Sprintf("%s_%s%s", datePart, randomPart, ext))
}

func buildR2PublicURL(key string) string {
	cfg := GlobalConfig.R2
	if cfg.PublicBaseURL != "" {
		return strings.TrimRight(cfg.PublicBaseURL, "/") + "/" + strings.TrimLeft(key, "/")
	}
	endpoint := strings.TrimRight(cfg.Endpoint, "/")
	return endpoint + "/" + cfg.Bucket + "/" + strings.TrimLeft(key, "/")
}
