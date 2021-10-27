package storage

import (
	"context"
	"github.com/alancesar/photo-gallery/api/photo"
	"github.com/minio/minio-go/v7"
	"io"
	"mime"
	"path/filepath"
)

type Storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(client *minio.Client, bucket string) *Storage {
	return &Storage{
		client: client,
		bucket: bucket,
	}
}

func (s *Storage) Put(ctx context.Context, filename string, reader io.Reader, size int64) (photo.Metadata, error) {
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	info, err := s.client.PutObject(ctx, s.bucket, filename, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	return photo.Metadata{
		ContentType: contentType,
		ETag:        info.ETag,
	}, err
}
