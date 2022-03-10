package usecase

import (
	"context"
	"github.com/alancesar/photo-gallery/api/domain/photo"
)

type (
	FindAll struct {
		db Database
	}
)

func NewGetAll(database Database) *FindAll {
	return &FindAll{
		db: database,
	}
}

func (fa FindAll) Execute(ctx context.Context) ([]photo.Photo, error) {
	return fa.db.FindAll(ctx)
}
