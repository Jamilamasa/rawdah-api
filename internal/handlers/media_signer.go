package handlers

import (
	"context"
	"strings"

	"github.com/rawdah/rawdah-api/internal/models"
)

type mediaURLSigner interface {
	CreatePresignedDownload(ctx context.Context, objectKey string) (string, error)
}

func applySignedMediaToUser(ctx context.Context, signer mediaURLSigner, user *models.User) {
	if user == nil {
		return
	}
	user.AvatarURL = signObjectKey(ctx, signer, user.AvatarURL)
}

func applySignedMediaToUsers(ctx context.Context, signer mediaURLSigner, users []*models.User) {
	for _, u := range users {
		applySignedMediaToUser(ctx, signer, u)
	}
}

func applySignedMediaToFamily(ctx context.Context, signer mediaURLSigner, family *models.Family) {
	if family == nil {
		return
	}
	family.LogoURL = signObjectKey(ctx, signer, family.LogoURL)
}

func signObjectKey(ctx context.Context, signer mediaURLSigner, objectKey *string) *string {
	if signer == nil || objectKey == nil {
		return objectKey
	}

	key := strings.TrimSpace(*objectKey)
	if key == "" {
		return objectKey
	}

	// Keep legacy full URLs untouched.
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return objectKey
	}

	signedURL, err := signer.CreatePresignedDownload(ctx, key)
	if err != nil {
		return objectKey
	}

	return &signedURL
}
