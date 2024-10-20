package artist

import (
	"fmt"

	"database/sql"
)

type ArtistRepository struct {
	db *sql.DB
}

func NewArtistRepository(db *sql.DB) *ArtistRepository {
	return &ArtistRepository{db: db}
}

func (repo *ArtistRepository) Create(artist *Artist) error {
	artistQuery := "INSERT INTO artists (name, description) VALUES ($1, $2) RETURNING id"
	var artistID int
	err := repo.db.QueryRow(artistQuery, artist.Name, artist.Description).Scan(&artistID)
	if err != nil {
		return fmt.Errorf("failed to insert artist: %w", err)
	}

	artist.ID = artistID
	return nil
}

func (repo *ArtistRepository) Exist(id int) (bool, error) {
	query := "SELECT COUNT(*) FROM artists WHERE id = $1"

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
