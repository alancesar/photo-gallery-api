package request

type (
	GetPhoto struct {
		ID string `uri:"id" binding:"required"`
	}
)
