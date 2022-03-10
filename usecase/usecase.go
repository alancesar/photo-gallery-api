package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/api/domain/metadata"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"io"
)

type (
	Database interface {
		Create(ctx context.Context, photo *photo.Photo) error
		FindAll(ctx context.Context) ([]photo.Photo, error)
		FindByID(ctx context.Context, id string) (photo.Photo, bool, error)
	}

	Bucket interface {
		Put(ctx context.Context, file io.ReadSeeker, filename string) (metadata.Metadata, error)
	}
)
