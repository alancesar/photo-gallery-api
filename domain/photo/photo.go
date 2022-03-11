package photo

import (
	"github.com/alancesar/photo-gallery/api/domain/metadata"
)

type (
	Thumbs struct {
		Filename string `json:"filename" firestore:"filename"`
		metadata.Metadata
	}

	Photo struct {
		ID       string         `json:"id" firestore:"-"`
		Filename string         `json:"filename" firestore:"filename"`
		Thumbs   []Thumbs       `json:"thumbs,omitempty" firestore:"thumbs,omitempty"`
		Exif     *metadata.Exif `json:"exif,omitempty" firestore:"exif,omitempty"`
		metadata.Metadata
	}
)
