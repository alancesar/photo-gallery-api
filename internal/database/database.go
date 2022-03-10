package database

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/alancesar/photo-gallery/api/domain/metadata"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	photosCollectionName = "photos"
	thumbsCollectionName = "thumbs"
)

type Database struct {
	collection *mongo.Collection
}

type (
	FirestoreDatabase struct {
		client *firestore.Client
	}
)

func NewFirestoreDatabase(client *firestore.Client) *FirestoreDatabase {
	return &FirestoreDatabase{
		client: client,
	}
}

func (d FirestoreDatabase) FindAll(ctx context.Context) ([]photo.Photo, error) {
	docs, err := d.client.Collection(photosCollectionName).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	photos := make([]photo.Photo, len(docs))
	for index := range docs {
		if err := docs[index].DataTo(&photos[index]); err != nil {
			return nil, err
		}

		photos[index].ID = docs[index].Ref.ID
	}

	return photos, nil
}

func (d FirestoreDatabase) FindByID(ctx context.Context, id string) (photo.Photo, bool, error) {
	doc, err := d.client.Collection(photosCollectionName).Doc(id).Get(ctx)
	if err != nil {
		return photo.Photo{}, false, err
	} else if doc == nil || !doc.Exists() {
		return photo.Photo{}, false, nil
	}

	var p photo.Photo
	if err := doc.DataTo(&p); err != nil {
		return photo.Photo{}, false, err
	}

	p.ID = doc.Ref.ID
	return p, true, nil
}

func (d FirestoreDatabase) Create(ctx context.Context, photo *photo.Photo) error {
	ref, _, err := d.client.Collection(photosCollectionName).Add(ctx, photo)
	if err != nil {
		return err
	}

	photo.ID = ref.ID
	return nil
}

func (d FirestoreDatabase) UpdateMetadata(ctx context.Context, id string, metadata metadata.Metadata) error {
	_, err := d.client.Collection(photosCollectionName).Doc(id).Update(ctx, []firestore.Update{
		{
			Path:  "metadata",
			Value: metadata,
		},
	})

	return err
}

func (d FirestoreDatabase) InsertThumbnails(ctx context.Context, id string, thumbs []photo.Thumbs) error {
	_, _, err := d.client.
		Collection(photosCollectionName).
		Doc(id).
		Collection(thumbsCollectionName).
		Add(ctx, thumbs)

	return err
}
