package venue

type Venue struct {
	ID       int       `db:"id" json:"id,omitempty"`
	Name     string    `db:"name" json:"name" validate:"required,min=3,max=100"`
	City     string    `db:"city" json:"city" validate:"required,min=2,max=50"`
	Country  string    `db:"country" json:"country" validate:"required,min=2,max=50"`
	Sections []Section `json:"sections,omitempty" validate:"dive"` // Validate each section
}

type Section struct {
	ID   int    `db:"id" json:"id,omitempty"`
	Name string `db:"name" json:"name" validate:"required,min=3,max=100"`
	Rows []Row  `json:"rows" validate:"dive"`
}

type Row struct {
	ID    int    `db:"id" json:"id,omitempty"`
	Name  string `db:"name" json:"name" validate:"required,min=1,max=50"`
	Seats []Seat `json:"seats" validate:"dive"`
}

type Seat struct {
	ID     int `db:"id" json:"id,omitempty"`
	Number int `db:"seat_number" json:"number" validate:"required,min=1"`
}

type SectionPriceRange struct {
	SectionID int `db:"section_id"`
	MinPrice  int `db:"min_price"`
	MaxPrice  int `db:"max_price"`
}

type SeatPriceBlock struct {
	StartSeatID     int `db:"start_seat_id"`
	StartSeatNumber int `db:"start_seat_number"`
	EndSeatID       int `db:"end_seat_id"`
	EndSeatNumber   int `db:"end_seat_number"`
	RowID           int `db:"row_id"`
	Price           int `db:"price"`
}

type RowCondition struct {
	RowID          int             `db:"row_id"`
	RowName        string          `db:"row_name"`
	SeatConditions []SeatCondition `db:"seat_conditions"`
}
type SeatCondition struct {
	SeatID     int  `db:"seat_id"`
	SeatNumber int  `db:"seat_number"`
	BookedBy   *int `db:"booked_by"`
}

type ConsecutiveSeats struct {
	RowID   int    `db:"row_id"`
	RowName string `db:"row_name"`
	Length  int    `db:"length"`
}
