package media

import (
	"io"

	firebase "firebase.google.com/go/v4"
	modelmedia "github.com/Basu008/GymBud/model/media"
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
