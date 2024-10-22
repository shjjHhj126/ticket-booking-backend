package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ticket-booking-backend/cmd/api/websocket"
	"ticket-booking-backend/domain/event"
	"ticket-booking-backend/domain/venue"
	"ticket-booking-backend/dto"
	"ticket-booking-backend/tool/rabbitmq"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
)

type TicketService struct {
	mq                *rabbitmq.RabbitMQ
	redisClient       *redislib.Client
	db                *sql.DB
	connectionManager *websocket.ConnectionManager
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

func (s *TicketService) ReserveTicket(ctx *gin.Context, eventID, sectionID, rowID, price, length int) error {
	sessionID, exists := ctx.Get("session_id")
	if !exists {
		return fmt.Errorf("session ID not found in context")
	}

	msg := dto.ReservationMsg{
		EventID:   eventID,
		SectionID: sectionID,
		RowID:     rowID,
		Price:     price,
		Length:    length,
		SessionID: sessionID.(string),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	err = s.mq.PublishMessage("book", msgBytes)
	if err != nil {
		return fmt.Errorf("failed to push to message queue: %w", err)
	}

	return nil
}

func (s *TicketService) HandleBookingMessage(data []byte) error {
	var msg dto.ReservationMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal msg, error: %w", err)
	}

	// Read the Lua script
	script, err := readLuaScript("reserve_seats_return_broadcast_info.lua")
	if err != nil {
		return err
	}

	// Execute the Lua script
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := s.redisClient.Eval(ctx, script,
		[]string{},
		fmt.Sprintf("%d", msg.EventID),
		fmt.Sprintf("%d", msg.SectionID),
		fmt.Sprintf("%d", msg.RowID),
		fmt.Sprintf("%d", msg.Length),
		msg.SessionID,
	).Result()
	if err != nil {
		return fmt.Errorf("failed to execute Lua script: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result format from Lua script")
	}

	// Extract the reserved seats
	reservedSeats, ok := resultMap["reserved"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to parse reserved seats")
	}

	// Extract max Consecutive length
	maxConsecutive, ok := resultMap["maxConsecutive"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to parse max consecutive lengths")
	}

	// Notify WebSocket client
	if err := s.NotifyReservation(msg, reservedSeats); err != nil {
		log.Printf("failed to notify WebSocket client: %v", err)
	}

	// Broadcast reservation message to all clients
	if err := s.broadcastReservation(msg, maxConsecutive); err != nil {
		log.Printf("failed to notify WebSocket clients: %v", err)
	}

	return nil
}

func (s *TicketService) NotifyReservation(msg dto.ReservationMsg, reservedSeats map[string]interface{}) error {
	log.Printf("reservedSeats:%+v", reservedSeats)

	reservationMsg := dto.ReservationMsg{
		EventID:   msg.EventID,
		SectionID: msg.SectionID,
		RowID:     msg.RowID,
		Price:     msg.Price,
		Length:    msg.Length,
		SessionID: msg.SessionID,
	}

	data, err := json.Marshal(reservationMsg)
	if err != nil {
		return fmt.Errorf("error marshaling reservation message: %w", err)
	}

	return s.connectionManager.NotifyReservation(data)
}

func (s *TicketService) broadcastReservation(msg dto.ReservationMsg, maxConsecutive map[string]interface{}) error {
	log.Printf("maxConsecutive:", maxConsecutive)

	broadcastMsg := dto.BroadcastMsg{
		EventID:     msg.EventID,
		SectionID:   msg.SectionID,
		RowID:       msg.RowID,
		Price:       msg.Price,
		MaxLength:   0,
		IsAvailable: false,
	}

	data, err := json.Marshal(broadcastMsg)
	if err != nil {
		return fmt.Errorf("error marshaling broadcast message: %w", err)
	}

	return s.connectionManager.BroadcastReservation(data)
}
