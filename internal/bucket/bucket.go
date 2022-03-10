package bucket

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/api/domain/metadata"
	"image"
	"io"
)

type (
	Bucket struct {
		bucketHandle *storage.BucketHandle
	}
)

func New(bucketHandle *storage.BucketHandle) *Bucket {
	return &Bucket{
		bucketHandle: bucketHandle,
	}
}

func (s Bucket) Put(ctx context.Context, reader io.ReadSeeker, name string) (metadata.Metadata, error) {
	object := s.bucketHandle.Object(name)
	writer := object.NewWriter(ctx)
	if _, err := io.Copy(writer, reader); err != nil {
		return metadata.Metadata{}, err
	}

	if err := writer.Close(); err != nil {
		return metadata.Metadata{}, err
	}

	dimension, err := getDimension(reader)
	if err != nil {
		return metadata.Metadata{}, err
	}

	attrs, err := object.Attrs(ctx)
	if err != nil {
		return metadata.Metadata{}, err
	}

	return metadata.Metadata{
		ContentType: attrs.ContentType,
		ETag:        attrs.Etag,
		MD5:         fmt.Sprintf("%x", attrs.MD5),
		Dimension:   dimension,
	}, nil
}

func getDimension(file io.ReadSeeker) (*metadata.Dimension, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	return &metadata.Dimension{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}
