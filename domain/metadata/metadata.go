package metadata

type (
	Exif map[string]map[string]string

	Dimension struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	Metadata struct {
		ContentType string     `json:"content_type"`
		ETag        string     `json:"etag,omitempty"`
		MD5         string     `json:"md5"`
		Dimension   *Dimension `json:"dimension,omitempty"`
	}
)
