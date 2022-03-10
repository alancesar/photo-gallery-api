package usecase

import (
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"github.com/alancesar/photo-gallery/api/pkg"
	"github.com/google/uuid"
	"io"
)

type (
	ErrUpload struct {
		message string
	}

	Upload struct {
		db Database
		b  Bucket
	}
)

func NewUpload(database Database, bucket Bucket) *Upload {
	return &Upload{
		db: database,
		b:  bucket,
	}
}

func (u ErrUpload) Error() string {
	return u.message
}

func (u ErrUpload) Is(target error) bool {
	return target == pkg.ErrInternal
}

func (u Upload) Execute(ctx context.Context, reader io.ReadSeeker, extension string) (photo.Photo, error) {
	filename := buildFilename(extension)
	md, err := u.b.Put(ctx, reader, filename)
	if err != nil {
		return photo.Photo{}, err
	}

	p := photo.Photo{
		Filename: filename,
		Metadata: md,
	}

	if err := u.db.Create(ctx, &p); err != nil {
		return photo.Photo{}, err
	}

	return p, nil
}

func buildFilename(extension string) string {
	return fmt.Sprintf("%s/%s%s", "photos", uuid.New().String(), extension)
}
