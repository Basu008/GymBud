package media

import "context"

type Repository interface {
	Create(ctx context.Context, media *Media) error
	DeleteByImageURL(ctx context.Context, imageURL string) error
}
