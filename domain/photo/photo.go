package photo

import (
	"github.com/alancesar/photo-gallery/api/domain/metadata"
)

type (
	Thumbs struct {
		Filename string `json:"filename"`
		metadata.Metadata
	}

	Photo struct {
		ID       string         `json:"id"`
		Filename string         `json:"filename"`
		Thumbs   []Thumbs       `json:"thumbs,omitempty"`
		Exif     *metadata.Exif `json:"exif,omitempty"`
		metadata.Metadata
	}
)
