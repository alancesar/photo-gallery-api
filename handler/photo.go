package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alancesar/photo-gallery/api/photo"
	"github.com/alancesar/photo-gallery/api/pubsub"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

type Database interface {
	Create(ctx context.Context, photo *photo.Photo) error
	FindAll(ctx context.Context) ([]photo.Photo, error)
	FindById(ctx context.Context, id string) (photo.Photo, bool, error)
}

type Storage interface {
	Put(ctx context.Context, filename string, file io.Reader, size int64) (photo.Metadata, error)
}

type Publisher interface {
	Publish(event pubsub.Event) error
}

const (
	jsonContentType  = "application/json"
	defaultExtension = ".jpg"
	filenameKey      = "filename"
)

type getPhotoRequest struct {
	ID string `uri:"id" binding:"required"`
}

type createdMessage struct {
	Filename string `json:"filename"`
	photo.Metadata
}

func UploadFileHandler(storage Storage, database Database, publisher Publisher) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid file",
			})
			return
		}

		filename := fmt.Sprintf("%s%s", uuid.New().String(), getExtension(header.Filename))
		metadata, err := storage.Put(ctx.Request.Context(), filename, file, header.Size)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		metadata.Dimension, err = getDimension(file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		p := photo.Photo{
			Filename: filename,
			Metadata: metadata,
		}

		if err := database.Create(ctx.Request.Context(), &p); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		if err := publishMessage(publisher, filename, metadata); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		ctx.JSON(http.StatusCreated, p)
	}
}

func ListPhotosHandler(database Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		photos, err := database.FindAll(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		ctx.JSON(http.StatusOK, photos)
	}
}

func GetPhotoHandler(database Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request := getPhotoRequest{}
		if err := ctx.ShouldBindUri(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if p, exists, err := database.FindById(ctx.Request.Context(), request.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
		} else if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": fmt.Sprintf("%s was not found", request.ID),
			})
		} else {
			ctx.JSON(http.StatusOK, p)
		}
	}
}

func getDimension(file multipart.File) (*photo.Dimension, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	return &photo.Dimension{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return defaultExtension
	}

	return ext
}

func publishMessage(publisher Publisher, filename string, metadata photo.Metadata) error {
	messaged, err := json.Marshal(&createdMessage{
		Filename: filename,
		Metadata: metadata,
	})
	if err != nil {
		return err
	}

	return publisher.Publish(pubsub.Event{
		Headers:     map[string]interface{}{filenameKey: filename},
		ContentType: jsonContentType,
		Message:     messaged,
	})
}
