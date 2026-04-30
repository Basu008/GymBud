package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	firebase "firebase.google.com/go/v4"
	modelmedia "github.com/Basu008/GymBud/model/media"
	"github.com/Basu008/GymBud/schema"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const maxImageSizeBytes = 10 << 20

func MaxMultipartImageMemory() int64 {
	return maxImageSizeBytes + 1
}

type Opts struct {
	Firebase *firebase.App
	Repo     modelmedia.Repository
	Logger   *zerolog.Logger
}

type Service struct {
	firebase *firebase.App
	repo     modelmedia.Repository
	logger   *zerolog.Logger
}

func NewMediaService(opts *Opts) *Service {
	return &Service{
		firebase: opts.Firebase,
		repo:     opts.Repo,
		logger:   opts.Logger,
	}
}

type UploadImageInput struct {
	OwnerID    string
	EntityType string
	EntityID   *string
	FileName   string
	File       io.Reader
}

func (s *Service) UploadImage(ctx context.Context, input *UploadImageInput) (*schema.UploadImageResponse, error) {
	if input == nil {
		return nil, errors.New("upload input is required")
	}
	if _, err := uuid.Parse(strings.TrimSpace(input.OwnerID)); err != nil {
		return nil, fmt.Errorf("invalid owner id: %w", err)
	}

	entityType := strings.TrimSpace(strings.ToLower(input.EntityType))
	if entityType == "" {
		return nil, errors.New("entity_type is required")
	}
	var entityID *string
	if input.EntityID != nil {
		value := strings.TrimSpace(*input.EntityID)
		if value != "" {
			if _, err := uuid.Parse(value); err != nil {
				return nil, fmt.Errorf("invalid entity_id: %w", err)
			}
			entityID = &value
		}
	}
	if input.File == nil {
		return nil, errors.New("image is required")
	}

	content, err := io.ReadAll(io.LimitReader(input.File, maxImageSizeBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	if len(content) == 0 {
		return nil, errors.New("image is required")
	}
	if len(content) > maxImageSizeBytes {
		return nil, errors.New("image must be 10MB or smaller")
	}

	mimeType := http.DetectContentType(content)
	extension, ok := allowedImageExtension(mimeType)
	if !ok {
		return nil, errors.New("image must be JPEG, PNG, GIF, or WebP")
	}

	mediaID := uuid.NewString()
	storagePath := fmt.Sprintf("media/%s/%s%s", input.OwnerID, mediaID, extension)
	publicURL, err := s.uploadToFirebase(ctx, storagePath, mimeType, content)
	if err != nil {
		return nil, err
	}

	media := &modelmedia.Media{
		ID:              mediaID,
		OwnerID:         input.OwnerID,
		EntityType:      entityType,
		EntityID:        entityID,
		MediaType:       modelmedia.MediaTypeImage,
		StorageProvider: modelmedia.StorageProviderFirebase,
		StoragePath:     storagePath,
		PublicURL:       publicURL,
		MimeType:        mimeType,
		SizeBytes:       int64(len(content)),
	}
	if err := s.repo.Create(ctx, media); err != nil {
		return nil, err
	}

	return &schema.UploadImageResponse{ImageURL: publicURL}, nil
}

func (s *Service) uploadToFirebase(ctx context.Context, storagePath, mimeType string, content []byte) (string, error) {
	storageClient, err := s.firebase.Storage(ctx)
	if err != nil {
		return "", fmt.Errorf("get firebase storage client: %w", err)
	}

	bucket, err := storageClient.DefaultBucket()
	if err != nil {
		return "", fmt.Errorf("get firebase storage bucket: %w", err)
	}

	bucketAttrs, err := bucket.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("get firebase storage bucket attrs: %w", err)
	}

	downloadToken := uuid.NewString()
	writer := bucket.Object(storagePath).NewWriter(ctx)
	writer.ContentType = mimeType
	writer.Metadata = map[string]string{
		"firebaseStorageDownloadTokens": downloadToken,
	}

	if _, err := io.Copy(writer, bytes.NewReader(content)); err != nil {
		writer.Close()
		return "", fmt.Errorf("upload image to firebase: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close firebase upload writer: %w", err)
	}

	return fmt.Sprintf(
		"https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media&token=%s",
		bucketAttrs.Name,
		url.PathEscape(storagePath),
		downloadToken,
	), nil
}

func allowedImageExtension(mimeType string) (string, bool) {
	switch mimeType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
