package metadata

type (
	Dimension struct {
		Width  int `json:"width" firestore:"width"`
		Height int `json:"height" firestore:"height"`
	}

	Metadata struct {
		ContentType string     `json:"content_type" firestore:"content_type"`
		ETag        string     `json:"etag,omitempty" firestore:"etag,omitempty"`
		MD5         string     `json:"md5,omitempty" firestore:"md5,omitempty"`
		Dimension   *Dimension `json:"dimension,omitempty" firestore:"dimension"`
	}

	Exif map[string]interface{}
)
