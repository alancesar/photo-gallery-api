package main

import (
	"context"
	"github.com/alancesar/photo-gallery/api/consumer"
	"github.com/alancesar/photo-gallery/api/handler"
	"github.com/alancesar/photo-gallery/api/mongodb"
	"github.com/alancesar/photo-gallery/api/photo"
	"github.com/alancesar/photo-gallery/api/pubsub"
	"github.com/alancesar/photo-gallery/api/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	dbHostEnv             = "DB_HOST"
	dbNameEnv             = "DB_NAME"
	rabbitMQUrlEnv        = "RABBITMQ_URL"
	minioEndpointEnv      = "MINIO_ENDPOINT"
	minioRootUserEnv      = "MINIO_ROOT_USER"
	minioRootPasswordEnv  = "MINIO_ROOT_PASSWORD"
	bucketNameEnv         = "PHOTOS_BUCKET"
	workerExchangeNameEnv = "WORKER_EXCHANGE_NAME"
	bucketExchangeNameEnv = "BUCKET_EXCHANGE_NAME"
	queueNameEnv          = "QUEUE_NAME"
	fanoutExchangeKind    = "fanout"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	minioClient, err := minio.New(os.Getenv(minioEndpointEnv), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv(minioRootUserEnv), os.Getenv(minioRootPasswordEnv), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	mongoClient, err := mongodb.NewClient(ctx, os.Getenv(dbHostEnv))
	if err != nil {
		log.Fatalln(err)
	}
	defer func(client *mongo.Client) {
		_ = client.Disconnect(ctx)
	}(mongoClient)

	connection, err := amqp.Dial(os.Getenv(rabbitMQUrlEnv))
	if err != nil {
		log.Fatalln(err)
	}
	defer func(conn *amqp.Connection) {
		_ = conn.Close()
	}(connection)

	channel, err := connection.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	defer func(ch *amqp.Channel) {
		_ = ch.Close()
	}(channel)

	if err = declareExchange(channel, os.Getenv(bucketExchangeNameEnv)); err != nil {
		log.Fatalln(err)
	}

	if err = declareExchange(channel, os.Getenv(workerExchangeNameEnv)); err != nil {
		log.Fatalln(err)
	}

	queue, err := declareAndBindQueue(channel, os.Getenv(queueNameEnv), os.Getenv(workerExchangeNameEnv))
	if err != nil {
		log.Fatalln(err)
	}

	db := photo.NewDatabase(mongoClient.Database(os.Getenv(dbNameEnv)))
	c := consumer.NewConsumer(db)
	p := pubsub.NewPublisher(channel, os.Getenv(bucketExchangeNameEnv))
	s := storage.NewStorage(minioClient, os.Getenv(bucketNameEnv))

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		subscriber := pubsub.NewSubscriber(channel, queue)
		if err := subscriber.Subscribe(ctx, c); err != nil {
			log.Println(err)
		}
	}()

	go func() {
		engine := gin.Default()
		engine.Use(cors.Default())
		engine.Handle(http.MethodPost, "/api/photos", handler.UploadFileHandler(s, db, p))
		engine.Handle(http.MethodGet, "/api/photos", handler.ListPhotosHandler(db))
		engine.Handle(http.MethodGet, "/api/photo/:id", handler.GetPhotoHandler(db))
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

func declareExchange(channel *amqp.Channel, exchangeName string) error {
	return channel.ExchangeDeclare(
		exchangeName,
		fanoutExchangeKind,
		true,
		false,
		false,
		false,
		nil,
	)
}

func declareAndBindQueue(channel *amqp.Channel, queue, exchange string) (amqp.Queue, error) {
	q, err := channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, err
	}

	return q, channel.QueueBind(q.Name, "", exchange, false, nil)
}
