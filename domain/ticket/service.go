package ticket

import (
	"context"
	"log"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/tool/rabbitmq"

	"database/sql"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
)

type TicketService struct {
	mq          *rabbitmq.RabbitMQ
	redisClient *redislib.Client
	db          *sql.DB
}

func NewTicketService(redisClient *redislib.Client, rmq *rabbitmq.RabbitMQ, db *sql.DB) *TicketService {
	return &TicketService{
		mq:          rmq,
		redisClient: redisClient,
		db:          db,
	}
}

func (s *TicketService) GetTickets(ctx *gin.Context, eventID, number, lowPrice, highPrice, page, pageSize int, venueService *venue.VenueService, eventService *event.EventService) ([]Ticket, error) {
	offset := (page - 1) * pageSize

	// Start Redis transaction by watching the key for sections in the price range
	var tickets []Ticket

	err := s.redisClient.Watch(ctx, func(tx *redislib.Tx) error {

		// Step 1: Check if section data is present
		sectionIDs, err := getSectionIDs(ctx.Request.Context(), tx, eventID, lowPrice, highPrice, venueService)
		if err != nil {
			return err
		}

		count := 0

		// Step 2: Fetch price blocks and seat availability for each section
		for _, sectionID := range sectionIDs {

			seatBlocks, err := getPriceBlocks(ctx, tx, eventID, sectionID, lowPrice, highPrice, venueService)
			if err != nil {
				return err
			}

			// remember:same price per priceBlock
			for _, priceBlock := range seatBlocks {

				consecutiveSeats, err := getConsecutiveSeatBlocks(ctx, tx, eventID, sectionID, venueService, &priceBlock)
				if err != nil {
					return err
				}

				log.Printf("get consecutiveSeats : %+v", consecutiveSeats)

				// Todo: cache sectionName
				sectionName, err := venueService.GetSectionNameByID(sectionID)
				if err != nil {
					return nil
				}

				log.Printf("sectionName : %s\n", sectionName)

				for _, consecutiveSeat := range consecutiveSeats {
					if consecutiveSeat.Length >= number {
						if count >= offset && len(tickets) < pageSize {
							ticket := Ticket{
								EventID:     eventID,
								SectionID:   sectionID,
								SectionName: sectionName,
								RowID:       priceBlock.RowID,
								RowName:     consecutiveSeat.RowName,
								Price:       priceBlock.Price,
								Length:      consecutiveSeat.Length,
							}
							tickets = append(tickets, ticket)
						}
						count++
						if len(tickets) == pageSize {
							return nil
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tickets, nil
}

func getSectionIDs(ctx context.Context, tx *redislib.Tx, eventID, lowPrice, highPrice int, venueService *venue.VenueService) ([]int, error) {
	sectionIDs, err := getSectionsByPriceRange(ctx, tx, eventID, lowPrice, highPrice)
	if err != nil {
		return []int{}, err
	}

	if len(sectionIDs) == 0 {
		// If no sections found in Redis, fetch from DB and cache it atomically in Redis
		sectionIDs, err = cacheSections(ctx, tx, eventID, lowPrice, highPrice, venueService)
		if err != nil {
			return []int{}, err
		}
	}

	return sectionIDs, nil
}

func getPriceBlocks(ctx context.Context, tx *redislib.Tx, eventID, sectionID, lowPrice, highPrice int, venueService *venue.VenueService) ([]venue.SeatPriceBlock, error) {
	seatBlocks, err := getSeatPriceBlocks(ctx, tx, eventID, sectionID, lowPrice, highPrice)
	if err != nil {
		return []venue.SeatPriceBlock{}, err
	}

	if len(seatBlocks) == 0 {
		seatBlocks, err = cacheSeatPriceBlocks(ctx, tx, eventID, sectionID, lowPrice, highPrice, venueService)
		if err != nil {
			return []venue.SeatPriceBlock{}, err
		}
	}

	return seatBlocks, nil
}

func getConsecutiveSeatBlocks(ctx context.Context, tx *redislib.Tx, eventID, sectionID int, venueService *venue.VenueService, priceBlock *venue.SeatPriceBlock) ([]venue.ConsecutiveSeats, error) {
	consecutiveSeats, err := getConsecutiveSeats(ctx, tx, eventID, sectionID, priceBlock)
	if err != nil {
		return []venue.ConsecutiveSeats{}, err
	}

	if len(consecutiveSeats) == 0 {
		consecutiveSeats, err = cacheConsecutiveSeats(ctx, tx, eventID, sectionID, priceBlock.RowID, priceBlock, venueService)
		if err != nil {
			return []venue.ConsecutiveSeats{}, err
		}
	}
	return consecutiveSeats, nil
}
