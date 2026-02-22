package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	appconfig "github.com/rawdah/rawdah-api/internal/config"
)

type R2Client struct {
	client    *s3.Client
	bucket    string
	publicURL string
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
	})

	return &R2Client{
		client:    client,
		bucket:    cfg.R2Bucket,
		publicURL: cfg.R2PublicURL,
	}, nil
}

// UploadAvatar resizes to 256x256, converts to webp-like jpeg, and uploads.
func (r *R2Client) UploadAvatar(ctx context.Context, familyID, userID string, fileBytes []byte, contentType string) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("image decode error: %w", err)
	}

	resized := imaging.Fill(img, 256, 256, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return "", fmt.Errorf("image encode error: %w", err)
	}

	key := fmt.Sprintf("families/%s/avatars/%s/%s.jpg", familyID, userID, uuid.New().String())

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", fmt.Errorf("r2 upload error: %w", err)
	}

	return fmt.Sprintf("%s/%s", r.publicURL, key), nil
}

// UploadLogo resizes to 512x512 and uploads as the family logo.
func (r *R2Client) UploadLogo(ctx context.Context, familyID string, fileBytes []byte, contentType string) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("image decode error: %w", err)
	}

	resized := imaging.Fill(img, 512, 512, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(90)); err != nil {
		return "", fmt.Errorf("image encode error: %w", err)
	}

	key := fmt.Sprintf("families/%s/logo/%s.jpg", familyID, uuid.New().String())

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", fmt.Errorf("r2 upload error: %w", err)
	}

	return fmt.Sprintf("%s/%s", r.publicURL, key), nil
}
