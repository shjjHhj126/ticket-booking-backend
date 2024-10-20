package event

import (
	"fmt"
	"time"

	"database/sql"
)

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (repo *EventRepository) Create(event *Event) error {
	// Define the query to insert a new event into the events table.
	eventQuery := `
        INSERT INTO events (name, created_at, updated_at, start_time, end_time, status, venue_id, artist_id, description) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
        RETURNING id`

	now := time.Now()

	err := repo.db.QueryRow(eventQuery,
		event.Name,
		now, // created_at
		now, // updated_at
		event.StartTime,
		event.EndTime,
		event.Status,
		event.VenueID,
		event.ArtistID,
		event.Description,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

func (repo *EventRepository) Exist(id int) (bool, error) {
	query := "SELECT COUNT(*) FROM events WHERE id = $1"

	var count int
	err := repo.db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func (repo *EventRepository) GetNameByID(id int) (string, error) {
	query := "SELECT name FROM events WHERE id = $1"

	var name string
	err := repo.db.QueryRow(query, id).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to check existence: %w", err)
	}

	return name, nil
}

func (repo *EventRepository) SetEventSeatPrice(eventSeatPrice EventSeatPrice) error {
	// EXCLUDED is a special table that represents the row that was proposed for insertion.
	query := `
		INSERT INTO event_seat (event_id, seat_id, price) 
		VALUES ($1, $2, $3)
		ON CONFLICT (event_id, seat_id) DO UPDATE 
		SET price = EXCLUDED.price
	`

	_, err := repo.db.Exec(query, eventSeatPrice.EventID, eventSeatPrice.SeatID, eventSeatPrice.Price)
	return err
}
