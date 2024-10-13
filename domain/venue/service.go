package venue

import "github.com/jmoiron/sqlx"

type VenueService struct {
	repo *VenueRepository
}

func NewVenueService(db *sqlx.DB) *VenueService {
	return &VenueService{
		repo: NewVenueRepository(db),
	}
}

func (s *VenueService) CreateVenue(v *Venue) error {
	return s.repo.Create(v)
}
