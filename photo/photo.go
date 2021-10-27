package photo

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	sizeRegex            = regexp.MustCompile(`(\d{1,8})`)
	errSizeTagNotPresent = errors.New("size tag not present")
)

const (
	imageWidthKey  = "Image Width"
	imageHeightKey = "Image Height"
)

type Type string

const (
	Thumb Type = "THUMB"
)

type Metadata struct {
	ContentType string     `json:"content_type" bson:"content_type"`
	ETag        string     `json:"etag,omitempty" bson:"etag,omitempty"`
	Dimension   *Dimension `json:"dimension,omitempty" bson:"dimension,omitempty"`
}

type Images []struct {
	Type     Type   `json:"type"`
	Filename string `json:"filename"`
	Metadata
}

func (i Images) GetThumbs() []Thumbs {
	var output []Thumbs

	for _, image := range i {
		if image.Type == Thumb {
			output = append(output, Thumbs{
				Filename: image.Filename,
				Metadata: image.Metadata,
			})
		}
	}

	return output
}

type Thumbs struct {
	Filename string `json:"filename"`
	Metadata `bson:",inline"`
}

type Exif map[string]string

type Dimension struct {
	Width  int
	Height int
}

func (e Exif) GetDimension() (Dimension, error) {
	stringWidth, ok := e[imageWidthKey]
	if !ok {
		return Dimension{}, errSizeTagNotPresent
	}

	stringHeight, ok := e[imageHeightKey]
	if !ok {
		return Dimension{}, errSizeTagNotPresent
	}

	width, err := strconv.Atoi(sizeRegex.FindString(stringWidth))
	if err != nil {
		return Dimension{}, err
	}

	height, err := strconv.Atoi(sizeRegex.FindString(stringHeight))
	if err != nil {
		return Dimension{}, err
	}

	return Dimension{
		Width:  width,
		Height: height,
	}, nil
}

type Photo struct {
	ID       string   `json:"id" bson:"_id,omitempty"`
	Filename string   `json:"filename"`
	Thumbs   []Thumbs `json:"thumbs,omitempty" bson:"thumbs,omitempty"`
	Exif     *Exif    `json:"exif,omitempty" bson:"exif,omitempty"`
	Metadata `bson:",inline"`
}
