package ticket

import (
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

func (s *TicketService) GetTickets(ctx *gin.Context, eventID, number, lowPrice, highPrice int, venueService *venue.VenueService, eventService *event.EventService) ([]Ticket, error) {
	// Start Redis transaction by watching the key for sections in the price range
	var tickets []Ticket

	err := s.redisClient.Watch(ctx, func(tx *redislib.Tx) error {

		// Step 1: Check if section data is present
		sectionIDs, err := getSectionsByPriceRange(ctx.Request.Context(), tx, eventID, lowPrice, highPrice)
		if err != nil {
			return err
		}

		if len(sectionIDs) == 0 {
			// If no sections found in Redis, fetch from DB and cache it atomically in Redis
			sectionIDs, err = cacheSections(ctx.Request.Context(), tx, eventID, lowPrice, highPrice, venueService)
			if err != nil {
				return err
			}
		}

		log.Printf("get sectionIDs : %+v", sectionIDs)

		// Step 2: Fetch price blocks and seat availability for each section
		for _, sectionID := range sectionIDs {
			seatBlocks, err := getSeatPriceBlocks(ctx.Request.Context(), tx, eventID, sectionID, lowPrice, highPrice)
			if err != nil {
				return err
			}

			if len(seatBlocks) == 0 {
				seatBlocks, err = cacheSeatPriceBlocks(ctx.Request.Context(), tx, eventID, sectionID, lowPrice, highPrice, venueService)
				if err != nil {
					return err
				}
			}

			log.Printf("get seatBlocks : %+v", seatBlocks)

			// same price per priceBlock
			for _, priceBlock := range seatBlocks {
				consecutiveSeats, err := getConsecutiveSeats(ctx.Request.Context(), tx, eventID, sectionID, &priceBlock)
				if err != nil {
					return err
				}

				if len(consecutiveSeats) == 0 {
					consecutiveSeats, err = cacheConsecutiveSeats(ctx.Request.Context(), tx, eventID, sectionID, priceBlock.RowID, &priceBlock, venueService)
					if err != nil {
						return err
					}
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
