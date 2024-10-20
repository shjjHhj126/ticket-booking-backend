package event

import "database/sql"

type EventService struct {
	repo *EventRepository
}

func NewEventService(db *sql.DB) *EventService {
	return &EventService{
		repo: NewEventRepository(db),
	}
}

func (s *EventService) CreateEvent(v *Event) error {
	return s.repo.Create(v)
}

func (s *EventService) SetEventSeatPrice(eventSeatPrice EventSeatPrice) error {
	return s.repo.SetEventSeatPrice(eventSeatPrice)
}

func (s *EventService) SetEventSeatsPrice(eventSeatPrices []EventSeatPrice) error {
	for _, item := range eventSeatPrices {
		if err := s.repo.SetEventSeatPrice(item); err != nil {
			return err
		}
	}
	return nil
}

func (s *EventService) Exist(id int) (bool, error) {
	return s.repo.Exist(id)
}

func (s *EventService) GetNameByID(id int) (string, error) {
	return s.repo.GetNameByID(id)
}
