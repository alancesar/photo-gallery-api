package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alancesar/photo-gallery/api/photo"
	"github.com/alancesar/photo-gallery/api/pubsub"
	"log"
)

const eventTypeKey = "event-type"

type eventType string

const (
	metadataEventType eventType = "METADATA"
	workerEventType   eventType = "WORKER"
)

type message struct {
	Filename string       `json:"filename"`
	Images   photo.Images `json:"images"`
	Exif     photo.Exif   `json:"exif"`
}

type mergeFn func(*photo.Photo, message)

type Database interface {
	FindByFilename(ctx context.Context, filename string) (photo.Photo, bool, error)
	Update(ctx context.Context, id string, photo photo.Photo) error
}

type Consumer struct {
	db Database
}

func NewConsumer(db Database) *Consumer {
	return &Consumer{
		db: db,
	}
}

func (c Consumer) Consume(ctx context.Context, event pubsub.Event) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("recovering from panic in consumer.Consumer: %v", r)
		}
	}()

	et := event.Headers[eventTypeKey].(string)
	merge, err := getMergeFn(eventType(et))
	if err != nil {
		return err
	}

	m := message{}
	if err := json.Unmarshal(event.Message, &m); err != nil {
		return err
	}

	p, exists, err := c.db.FindByFilename(ctx, m.Filename)
	if err != nil {
		return err
	} else if !exists {
		return errors.New(fmt.Sprintf("%s does not exist", m.Filename))
	}

	merge(&p, m)
	return c.db.Update(ctx, p.ID, p)
}

func getMergeFn(et eventType) (mergeFn, error) {
	switch et {
	case metadataEventType:
		return mergeMetadata, nil
	case workerEventType:
		return mergeThumbs, nil
	default:
		return nil, errors.New(fmt.Sprintf("%s is an invalid event type", et))
	}
}

func mergeMetadata(p *photo.Photo, message message) {
	dimension, _ := message.Exif.GetDimension()
	p.Metadata.Dimension = &dimension
	p.Exif = &message.Exif
}

func mergeThumbs(p *photo.Photo, message message) {
	p.Thumbs = message.Images.GetThumbs()
}
