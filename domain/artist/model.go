package artist

type Artist struct {
	ID          int    `db:"id" json:"id,omitempty"`
	Name        string `db:"name" json:"name" validate:"required,min=3,max=100"`
	Description string `db:"description" json:"description,omitempty"`
}
