package handler

import (
	"context"
	"errors"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"github.com/alancesar/photo-gallery/api/pkg"
	"github.com/alancesar/photo-gallery/api/presenter/request"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"path/filepath"
)

const (
	defaultExtension = ".jpg"
)

type (
	UploadUseCase interface {
		Execute(context.Context, io.ReadSeeker, string) (photo.Photo, error)
	}

	GetUseCase interface {
		Execute(context.Context, string) (photo.Photo, error)
	}

	GetAllUseCase interface {
		Execute(context.Context) ([]photo.Photo, error)
	}
)

func UploadFileHandler(uc UploadUseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		file, header, err := ctx.Request.FormFile("file")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid file",
			})
			return
		}

		extension := getExtension(header.Filename)
		p, err := uc.Execute(ctx.Request.Context(), file, extension)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		ctx.JSON(http.StatusCreated, p)
	}
}

func ListPhotosHandler(uc GetAllUseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		photos, err := uc.Execute(ctx.Request.Context())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "unexpected error. please try again later",
			})
			return
		}

		ctx.JSON(http.StatusOK, photos)
	}
}

func GetPhotoHandler(uc GetUseCase) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetPhoto{}
		if err := ctx.ShouldBindUri(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}

		if p, err := uc.Execute(ctx.Request.Context(), req.ID); err != nil {
			if err != nil {
				if errors.Is(err, pkg.ErrNotFound) {
					ctx.JSON(http.StatusNotFound, gin.H{
						"error": err.Error(),
					})
					return
				}
			}

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "unexpected error. please try again later",
			})
			return
		} else {
			ctx.JSON(http.StatusOK, p)
		}
	}
}

func getExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return defaultExtension
	}

	return ext
}
