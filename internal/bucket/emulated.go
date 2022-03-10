package bucket

import (
	"cloud.google.com/go/storage"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/alancesar/photo-gallery/api/domain/metadata"
	"io"
	"mime"
	"path"
)

type (
	Emulated struct {
		bucketHandle *storage.BucketHandle
	}
)

func NewEmulated(bucketHandle *storage.BucketHandle) *Emulated {
	return &Emulated{
		bucketHandle: bucketHandle,
	}
}

func (s Emulated) Put(ctx context.Context, reader io.ReadSeeker, name string) (metadata.Metadata, error) {
	writer := s.bucketHandle.Object(name).NewWriter(ctx)
	hash := md5.New()

	if _, err := io.Copy(writer, io.TeeReader(reader, hash)); err != nil {
		return metadata.Metadata{}, err
	}

	if err := writer.Close(); err != nil {
		return metadata.Metadata{}, err
	}

	dimension, err := getDimension(reader)
	if err != nil {
		return metadata.Metadata{}, err
	}

	md5hash := fmt.Sprintf("%x", hash.Sum(nil))
	return metadata.Metadata{
		ContentType: mime.TypeByExtension(path.Ext(name)),
		ETag:        md5hash,
		MD5:         md5hash,
		Dimension:   dimension,
	}, nil
}
