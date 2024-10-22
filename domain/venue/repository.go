package venue

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

type VenueRepository struct {
	db *sql.DB
}

func NewVenueRepository(db *sql.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

func (repo *VenueRepository) Create(venue *Venue) error {
	// Start a database transaction
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	// Insert the venue
	venueQuery := "INSERT INTO venues (name, city, country) VALUES ($1, $2, $3) RETURNING id"
	var venueID int
	err = tx.QueryRow(venueQuery, venue.Name, venue.City, venue.Country).Scan(&venueID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert sections
	for _, section := range venue.Sections {
		sectionQuery := "INSERT INTO sections (name, venue_id) VALUES ($1, $2) RETURNING id"
		var sectionID int
		err = tx.QueryRow(sectionQuery, section.Name, venueID).Scan(&sectionID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert venue: %w", err)
		}

		// Insert rows
		for _, row := range section.Rows {
			rowQuery := "INSERT INTO rows (name, section_id) VALUES ($1, $2) RETURNING id"
			var rowID int
			err = tx.QueryRow(rowQuery, row.Name, sectionID).Scan(&rowID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to insert rows: %w", err)
			}

			// Insert seats
			for _, seat := range row.Seats {
				seatQuery := "INSERT INTO seats (seat_number, row_id) VALUES ($1, $2)"
				_, err = tx.Exec(seatQuery, seat.Number, rowID)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to insert seats: %w", err)
				}
			}
		}
	}

	// Commit the transaction
	return tx.Commit()
}

func (repo *VenueRepository) Exist(id int) (bool, error) {
	query := "SELECT COUNT(*) FROM venues WHERE id = $1"

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

func (repo *VenueRepository) SeatExist(id int) (bool, error) {
	query := "SELECT COUNT(*) FROM seats WHERE id = $1"

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

func (repo *VenueRepository) GetSectionIds(eventID int) ([]SectionPriceRange, error) {
	query := `
		SELECT sections.id AS section_id, 
			MIN(event_seat.price) AS min_price,
			MAX(event_seat.price) AS max_price
		FROM sections 
		LEFT JOIN rows ON rows.section_id = sections.id
		LEFT JOIN seats ON seats.row_id = rows.id
		LEFT JOIN event_seat ON event_seat.seat_id = seats.id
		LEFT JOIN events ON events.id = event_seat.event_id
		WHERE events.id = $1
		GROUP BY sections.id
		`
	rows, err := repo.db.Query(query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	SectionPriceRangeArray := []SectionPriceRange{}
	for rows.Next() {
		var sectionPriceRange SectionPriceRange
		if err := rows.Scan(&sectionPriceRange.SectionID, &sectionPriceRange.MinPrice, &sectionPriceRange.MaxPrice); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		SectionPriceRangeArray = append(SectionPriceRangeArray, sectionPriceRange)
	}

	return SectionPriceRangeArray, nil
}

func (repo *VenueRepository) GetSeatPriceBlocks(eventID, sectionID int) ([]SeatPriceBlock, error) {
	/*
		seat_number - ROW_NUMBER():
		constant for consecutive seats but changes when there's a gap
	*/

	query := `
		WITH seat_info AS (
			SELECT 
				seats.id AS seat_id, 
				seats.seat_number, 
				events.id AS event_id,
				rows.id AS row_id,
				event_seat.price AS price
			FROM sections
			LEFT JOIN rows ON rows.section_id = sections.id
			LEFT JOIN seats ON seats.row_id = rows.id
			LEFT JOIN event_seat ON event_seat.seat_id = seats.id
			LEFT JOIN events ON events.id = event_seat.event_id
			WHERE sections.id = $1
			AND events.id = $2
		),
		seat_consecutive AS (
			SELECT 
				seat_id, 
				seat_number, 
				row_id, 
				price,
				seat_number - ROW_NUMBER() OVER (PARTITION BY row_id, price ORDER BY seat_number) AS grouping_key
			FROM seat_info
		)
		SELECT 
			MIN(seat_id) AS start_seat_id,
			MIN(seat_number) AS start_seat_number, 
			MAX(seat_id) AS end_seat_id,
			MAX(seat_number) AS end_seat_number, 
			row_id, 
			price
		FROM seat_consecutive
		GROUP BY row_id, price, grouping_key
		HAVING price BETWEEN 500 and 600
		ORDER BY row_id, start_seat_number
	`

	rows, err := repo.db.Query(query, sectionID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to query row: %w", err)
	}
	defer rows.Close()

	seatPriceBlocks := []SeatPriceBlock{}
	for rows.Next() {
		var seatPriceBlock SeatPriceBlock
		if err := rows.Scan(
			&seatPriceBlock.StartSeatID,
			&seatPriceBlock.StartSeatNumber,
			&seatPriceBlock.EndSeatID,
			&seatPriceBlock.EndSeatNumber,
			&seatPriceBlock.RowID,
			&seatPriceBlock.Price); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		seatPriceBlocks = append(seatPriceBlocks, seatPriceBlock)
	}

	return seatPriceBlocks, nil
}

func (repo *VenueRepository) GetSectionNameByID(id int) (string, error) {
	query := `SELECT name FROM sections WHERE sections.id = $1`

	var name string
	err := repo.db.QueryRow(query, id).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to query section name with id = %d : %w", id, err)
	}

	return name, nil
}

// row condition means seat booking condition in a row in an event.
func (repo *VenueRepository) GetRowConditionByID(rowID, eventID int) (RowCondition, error) {

	query := `
		SELECT 
			rows.id AS row_id,
			rows.name AS row_name,
			ARRAY_AGG(seats.id) AS seat_ids,
			ARRAY_AGG(seats.seat_number) AS seat_numbers,
			ARRAY_AGG(bookings.booked_by) AS booked_bys 
		FROM rows
		LEFT JOIN seats ON seats.row_id = rows.id
		LEFT JOIN event_seat ON event_seat.seat_id = seats.id
		LEFT JOIN bookings ON bookings.event_seat_id = event_seat.id
		WHERE rows.id = $1
		AND event_seat.event_id = $2
		GROUP BY rows.id
	`

	rowCondition := RowCondition{}
	var seatIDs, seatNumbers, bookedBys []sql.NullInt64

	err := repo.db.QueryRow(query, rowID, eventID).Scan(
		&rowCondition.RowID,
		&rowCondition.RowName,
		pq.Array(&seatIDs),
		pq.Array(&seatNumbers),
		pq.Array(&bookedBys))
	if err != nil {
		if err == sql.ErrNoRows {
			return RowCondition{}, nil
		}
		return RowCondition{}, err
	}

	if len(seatIDs) != len(seatNumbers) || len(seatIDs) != len(bookedBys) {
		return RowCondition{}, fmt.Errorf("mismatched data lengths: seatIDs=%d, seatNumbers=%d, bookedBys=%d", len(seatIDs), len(seatNumbers), len(bookedBys))
	}

	log.Print("GetRowConditionByID seat length", len(seatIDs))

	for i := 0; i < len(seatIDs); i++ {
		var bookedByPtr *int

		if bookedBys[i].Valid {
			bookedByInt := int(bookedBys[i].Int64)
			bookedByPtr = &bookedByInt
		}

		rowCondition.SeatConditions = append(rowCondition.SeatConditions, SeatCondition{
			SeatID:     int(seatIDs[i].Int64),
			SeatNumber: int(seatNumbers[i].Int64),
			BookedBy:   bookedByPtr,
		})
	}

	return rowCondition, nil
}
