package event

import (
	"time"
)

type Event struct {
	ID          int       `db:"id"`
	Name        string    `db:"name" validate:"required,min=3,max=100"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	Status      string    `db:"status" validate:"required"`
	VenueID     int       `db:"venue_id"`
	ArtistID    int       `db:"artist_id"`
	Description string    `db:"description,omitempty" validate:"max=500"`
}

type EventSeatPrice struct {
	EventID int `db:"event_id"`
	SeatID  int `db:"seat_id"`
	Price   int `db:"price"`
}
