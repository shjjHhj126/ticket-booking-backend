package ticket

type Ticket struct {
	EventID     int    `json:"event_id"`
	SectionID   int    `json:"section_id"`
	SectionName string `json:"section_name"`
	RowID       int    `json:"row_id"`
	RowName     string `json:"row_name"`
	Price       int    `json:"price"`
	Length      int    `json:"length"`
}

type Seat struct {
	SeatNumber int `json:"seat_number"`
}

type BookingRequest struct {
	Section string `json:"section"`
	Row     string `json:"row"`
	Price   int    `json:"price"`
	Length  int    `json:"length"`
}
