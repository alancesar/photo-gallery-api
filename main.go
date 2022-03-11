package main

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"github.com/alancesar/photo-gallery/api/internal/bucket"
	"github.com/alancesar/photo-gallery/api/internal/database"
	"github.com/alancesar/photo-gallery/api/internal/listener"
	"github.com/alancesar/photo-gallery/api/internal/publisher"
	"github.com/alancesar/photo-gallery/api/presenter/handler"
	"github.com/alancesar/photo-gallery/api/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	projectIDKey           = "PROJECT_ID"
	storageEmulatorHostKey = "STORAGE_EMULATOR_HOST"
	photosTopicID          = "photos"
	thumbsSubscriptionID   = "api_thumbs"
)

func main() {
	projectID := os.Getenv(projectIDKey)
	ctx, cancel := context.WithCancel(context.Background())

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	firestoreClient, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalln(err)
	}

	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalln(err)
	}

	photosTopic := pubsubClient.Topic(photosTopicID)
	thumbsSubscription := pubsubClient.Subscription(thumbsSubscriptionID)
	bucketHandle := storageClient.Bucket(fmt.Sprintf("%s.appspot.com", projectID))

	db := database.NewFirestoreDatabase(firestoreClient)
	b := buildBucket(bucketHandle)
	p := publisher.New[photo.Photo](photosTopic)

	uploadUseCase := usecase.NewUpload(db, b, p)
	getAllUseCase := usecase.NewGetAll(db)
	getUseCase := usecase.NewGet(db)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		l := listener.New[photo.Photo](thumbsSubscription)
		if err := l.Listen(ctx, func(ctx context.Context, p photo.Photo) error {
			return db.InsertThumbnails(ctx, p.ID, p.Thumbs)
		}); err != nil {
			log.Fatalln(err)
		}
	}()

	go func() {
		engine := gin.Default()
		engine.Use(cors.Default())
		engine.Handle(http.MethodPost, "/api/photos", handler.UploadFileHandler(uploadUseCase))
		engine.Handle(http.MethodGet, "/api/photos", handler.ListPhotosHandler(getAllUseCase))
		engine.Handle(http.MethodGet, "/api/photo/:id", handler.GetPhotoHandler(getUseCase))
		if err := engine.Run(":8080"); err != nil {
			log.Fatalln(err)
		}
	}()

	for {
		select {
		case <-signals:
			log.Println("shutting down...")
			cancel()
		case <-ctx.Done():
			log.Fatalln(ctx.Err())
		}
	}
}

func buildBucket(photosBucket *storage.BucketHandle) usecase.Bucket {
	if os.Getenv(storageEmulatorHostKey) != "" {
		return bucket.NewEmulated(photosBucket)
	}

	return bucket.New(photosBucket)
}
