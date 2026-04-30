package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	modelmedia "github.com/Basu008/GymBud/model/media"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaRepo struct {
	db *pgxpool.Pool
}

func NewMediaRepo(db *pgxpool.Pool) (*MediaRepo, error) {
	repo := &MediaRepo{db: db}
	if err := repo.initTable(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *MediaRepo) initTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		CREATE TABLE IF NOT EXISTS public.media (
			id UUID PRIMARY KEY,
			owner_id UUID NOT NULL,
			entity_type TEXT NOT NULL,
			entity_id UUID,
			media_type TEXT NOT NULL,
			storage_provider TEXT NOT NULL,
			storage_path TEXT NOT NULL,
			public_url TEXT NOT NULL,
			mime_type TEXT,
			size_bytes BIGINT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("create media table: %w", err)
	}

	return nil
}

func (r *MediaRepo) Create(ctx context.Context, media *modelmedia.Media) error {
	if media == nil {
		return errors.New("media is required")
	}

	const query = `
		INSERT INTO media (
			id,
			owner_id,
			entity_type,
			entity_id,
			media_type,
			storage_provider,
			storage_path,
			public_url,
			mime_type,
			size_bytes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	if err := r.db.QueryRow(
		ctx,
		query,
		media.ID,
		media.OwnerID,
		media.EntityType,
		media.EntityID,
		media.MediaType,
		media.StorageProvider,
		media.StoragePath,
		media.PublicURL,
		media.MimeType,
		media.SizeBytes,
	).Scan(&media.CreatedAt, &media.UpdatedAt); err != nil {
		return fmt.Errorf("create media: %w", err)
	}

	return nil
}

func (r *MediaRepo) DeleteByImageURL(ctx context.Context, imageURL string) error {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return nil
	}

	if _, err := r.db.Exec(ctx, `DELETE FROM media WHERE public_url = $1`, imageURL); err != nil {
		return fmt.Errorf("delete media by image url: %w", err)
	}

	return nil
}
