package photo

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const photoCollectionName = "photos"

type Database struct {
	collection *mongo.Collection
}

func NewDatabase(database *mongo.Database) *Database {
	return &Database{
		collection: database.Collection(photoCollectionName),
	}
}

func (d Database) FindAll(ctx context.Context) ([]Photo, error) {
	result, err := d.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	var photos []Photo
	err = result.All(ctx, &photos)
	return photos, err
}

func (d Database) FindById(ctx context.Context, id string) (Photo, bool, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Photo{}, false, err
	}

	photo := Photo{}
	filter := bson.D{{"_id", objectId}}
	if err := d.collection.FindOne(ctx, filter).Decode(&photo); err != nil {
		return Photo{}, false, handleError(err)
	}

	return photo, true, nil
}

func (d Database) FindByFilename(ctx context.Context, filename string) (Photo, bool, error) {
	photo := Photo{}
	filter := bson.D{{"filename", filename}}
	err := d.collection.FindOne(ctx, filter).Decode(&photo)
	if err != nil {
		return Photo{}, false, handleError(err)
	}

	return photo, true, nil
}

func (d Database) Create(ctx context.Context, photo *Photo) error {
	result, err := d.collection.InsertOne(ctx, photo)
	if err != nil {
		return err
	}

	photo.ID, err = retrieveCreatedId(result)
	return err
}

func (d Database) Update(ctx context.Context, id string, photo Photo) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	photo.ID = ""
	_, err = d.collection.UpdateByID(ctx, oid, bson.M{"$set": photo})
	return err
}

func retrieveCreatedId(result *mongo.InsertOneResult) (string, error) {
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return "", errors.New("malformed id")
}

func handleError(err error) error {
	if err == mongo.ErrNoDocuments {
		return nil
	}

	return err
}
