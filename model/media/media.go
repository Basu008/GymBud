package media

import "time"

const MediaTypeImage = "image"
const StorageProviderFirebase = "firebase"

type Media struct {
	ID              string
	OwnerID         string
	EntityType      string
	EntityID        *string
	MediaType       string
	StorageProvider string
	StoragePath     string
	PublicURL       string
	MimeType        string
	SizeBytes       int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
