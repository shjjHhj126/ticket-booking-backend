package payment

import "time"

type Booking struct {
	ID          int       `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	EventSeatID int       `db:"event_seat_id"`
	BookedBy    int       `db:"booked_by"`
}
