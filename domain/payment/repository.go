package payment

import (
	"database/sql"
	"fmt"
	"ticket-booking-backend/dto"
	"time"

	"github.com/lib/pq"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Todo: change current pipeline to regular insert
func (r *PaymentRepository) PipelineInsertData(bookInfo dto.DbBookDTO) error {
	// Create a slice of seat numbers
	seatNumbers := make([]int, bookInfo.Length)
	for i := 0; i < bookInfo.Length; i++ {
		seatNumbers[i] = bookInfo.StartSeatNumber + i
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Get event_seat_ids
	query := `
        SELECT es.id 
        FROM event_seat es
        JOIN seats s ON es.seat_id = s.id
        WHERE es.event_id = $1
        AND s.row_id = $2
        AND s.seat_number = ANY($3)
        ORDER BY s.seat_number
    `

	rows, err := tx.Query(query, bookInfo.EventID, bookInfo.RowID, pq.Array(seatNumbers))
	if err != nil {
		return fmt.Errorf("failed to get event_seat_ids: %v", err)
	}
	defer rows.Close()

	var eventSeatIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan event_seat_id: %v", err)
		}
		eventSeatIDs = append(eventSeatIDs, id)
	}

	if len(eventSeatIDs) != bookInfo.Length {
		return fmt.Errorf("expected %d seats, but found %d", bookInfo.Length, len(eventSeatIDs))
	}

	// Create a pipeline using COPY
	stmt, err := tx.Prepare(pq.CopyIn("bookings", "created_at", "event_seat_id", "booked_by"))
	if err != nil {
		return fmt.Errorf("failed to prepare copy statement: %v", err)
	}
	defer stmt.Close()

	// Stream the data through the pipeline
	now := time.Now()
	for _, eventSeatID := range eventSeatIDs {
		_, err := stmt.Exec(now, eventSeatID, bookInfo.UserID) // buffer data
		if err != nil {
			return fmt.Errorf("failed to exec copy: %v", err)
		}
	}

	// Complete the COPY command
	_, err = stmt.Exec() // execute at once
	if err != nil {
		return fmt.Errorf("failed to complete copy: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
