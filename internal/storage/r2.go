package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	appconfig "github.com/rawdah/rawdah-api/internal/config"
)

type R2Client struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	presignTTL    time.Duration
}

var ErrObjectNotFound = errors.New("object not found")

type ObjectMetadata struct {
	SizeBytes   int64
	ContentType string
}

func NewR2Client(cfg *appconfig.Config) (*R2Client, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, ""),
		),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("r2 config error: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
		o.UsePathStyle = true
	})

	return &R2Client{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.R2Bucket,
		presignTTL:    time.Duration(cfg.PresignExpiresSeconds) * time.Second,
	}, nil
}

func (r *R2Client) PresignTTLSeconds() int {
	return int(r.presignTTL.Seconds())
}

func (r *R2Client) BuildAvatarObjectKey(familyID, userID, contentType string) (string, error) {
	ext, err := extensionForContentType(contentType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("families/%s/avatars/%s/%s.%s", familyID, userID, uuid.NewString(), ext), nil
}

func (r *R2Client) BuildLogoObjectKey(familyID, contentType string) (string, error) {
	ext, err := extensionForContentType(contentType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("families/%s/logos/%s.%s", familyID, uuid.NewString(), ext), nil
}

func (r *R2Client) CreatePresignedUpload(ctx context.Context, objectKey, contentType string) (string, error) {
	res, err := r.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = r.presignTTL
	})
	if err != nil {
		return "", fmt.Errorf("presign put failed: %w", err)
	}
	return res.URL, nil
}

func (r *R2Client) CreatePresignedDownload(ctx context.Context, objectKey string) (string, error) {
	res, err := r.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = r.presignTTL
	})
	if err != nil {
		return "", fmt.Errorf("presign get failed: %w", err)
	}
	return res.URL, nil
}

func (r *R2Client) GetObjectMetadata(ctx context.Context, objectKey string) (*ObjectMetadata, error) {
	out, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	})
	if err == nil {
		return &ObjectMetadata{
			SizeBytes:   aws.ToInt64(out.ContentLength),
			ContentType: aws.ToString(out.ContentType),
		}, nil
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		if code == "NotFound" || code == "NoSuchKey" {
			return nil, ErrObjectNotFound
		}
	}

	return nil, fmt.Errorf("head object failed: %w", err)
}

func extensionForContentType(contentType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/webp":
		return "webp", nil
	default:
		return "", fmt.Errorf("unsupported content type")
	}
}
