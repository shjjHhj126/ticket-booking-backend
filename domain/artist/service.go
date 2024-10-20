package artist

import "database/sql"

type ArtistService struct {
	repo *ArtistRepository
}

func NewArtistService(db *sql.DB) *ArtistService {
	return &ArtistService{
		repo: NewArtistRepository(db),
	}
}

func (s *ArtistService) CreateArtist(v *Artist) error {
	return s.repo.Create(v)
}

func (s *ArtistService) Exist(id int) (bool, error) {
	return s.repo.Exist(id)
}
