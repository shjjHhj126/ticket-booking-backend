package ticket

type Ticket struct {
	TicketID int64  `json:"ticketID"` // Use int64 instead of long in Go
	Section  string `json:"section"`
	Row      string `json:"row"`
	Seats    []Seat `json:"seats"`
	Price    int    `json:"price"`
}

type Seat struct {
	SeatNumber int `json:"seat_number"`
}

type BookingRequest struct {
	UserID   string `json:"user_id"`
	TicketID string `json:"ticket_id"`
}
