package venue

import "database/sql"

type VenueService struct {
	repo *VenueRepository
}

func NewVenueService(db *sql.DB) *VenueService {
	return &VenueService{
		repo: NewVenueRepository(db),
	}
}

func (s *VenueService) CreateVenue(v *Venue) error {
	return s.repo.Create(v)
}

func (s *VenueService) Exist(id int) (bool, error) {
	return s.repo.Exist(id)
}

func (s *VenueService) SeatExist(id int) (bool, error) {
	return s.repo.SeatExist(id)
}

func (s *VenueService) GetSectionIds(eventID int) ([]SectionPriceRange, error) {
	return s.repo.GetSectionIds(eventID)
}

func (s *VenueService) GetSeatPriceBlocks(eventID, sectionID int) ([]SeatPriceBlock, error) {
	return s.repo.GetSeatPriceBlocks(eventID, sectionID)
}

func (s *VenueService) GetSectionNameByID(id int) (string, error) {
	return s.repo.GetSectionNameByID(id)
}

func (s *VenueService) GetRowConditionByID(rowID, eventID int) (RowCondition, error) {
	return s.repo.GetRowConditionByID(rowID, eventID)
}
