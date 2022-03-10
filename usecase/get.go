package usecase

import (
	"context"
	"fmt"
	"github.com/alancesar/photo-gallery/api/domain/photo"
	"github.com/alancesar/photo-gallery/api/pkg"
)

type (
	ErrNotFound struct {
		message string
	}

	Get struct {
		db Database
	}
)

func NewErrNotFound(id string) ErrNotFound {
	return ErrNotFound{
		message: fmt.Sprintf("not found photo with id %s", id),
	}
}

func (nf ErrNotFound) Error() string {
	return nf.message
}

func (nf ErrNotFound) Is(target error) bool {
	return target == pkg.ErrNotFound
}

func NewGet(db Database) *Get {
	return &Get{
		db: db,
	}
}

func (g Get) Execute(ctx context.Context, id string) (photo.Photo, error) {
	p, exists, err := g.db.FindByID(ctx, id)
	if err == nil && !exists {
		return photo.Photo{}, NewErrNotFound(id)
	}

	return p, err
}
