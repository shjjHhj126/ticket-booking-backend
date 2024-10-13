package venue

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type VenueRepository struct {
	db *sqlx.DB
}

func NewVenueRepository(db *sqlx.DB) *VenueRepository {
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
				seatQuery := "INSERT INTO seats (seat_number, price, row_id) VALUES ($1, $2, $3)"
				_, err = tx.Exec(seatQuery, seat.Number, seat.Price, rowID)
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
