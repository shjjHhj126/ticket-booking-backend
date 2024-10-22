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

func (s *TicketService) GetTickets(ctx *gin.Context,
	eventID, number, lowPrice, highPrice, page, pageSize int,
	venueService *venue.VenueService, eventService *event.EventService) ([]Ticket, error) {
	offset := (page - 1) * pageSize
	var tickets []Ticket

	// Start Redis transaction by watching the key for sections in the price range

	err := s.redisClient.Watch(ctx, func(tx *redislib.Tx) error {

		// Check if section data is present
		sectionIDs, err := getSectionIDs(ctx.Request.Context(), tx, eventID, lowPrice, highPrice, venueService)
		if err != nil {
			return err
		}

		fmt.Printf("sectionIDs:%+v", sectionIDs)

		count := 0
		priceInfoMap := make(map[int]priceInfo) // Initialize the map to store price info

		// Fetch price blocks and seat availability for each section
		for _, sectionID := range sectionIDs {

			seatBlocks, err := getPriceBlocks(ctx, tx, eventID, sectionID, lowPrice, highPrice, venueService)
			if err != nil {
				return err
			}

			// Calculate max length per price
			for _, priceBlock := range seatBlocks {

				// Get the row condition of this priceBlock
				seatStatuses, err := getConsecutiveSeatBlocks(ctx, tx, eventID, sectionID, venueService, &priceBlock)
				if err != nil {
					return err
				}

				seatStatusesPiece := seatStatuses[priceBlock.StartSeatNumber-1 : priceBlock.EndSeatNumber]

				maxLen := 0
				curLen := 0
				for _, status := range seatStatusesPiece {
					if status == '0' {
						curLen += 1
					} else {
						if curLen > maxLen {
							maxLen = curLen
						}
						curLen = 0
					}
				}

				// Final update
				if curLen > maxLen {
					maxLen = curLen
				}

				// Update the priceIntoMap
				if maxLen > priceInfoMap[priceBlock.Price].Length {
					priceInfoMap[priceBlock.Price] = priceInfo{
						SectionID: sectionID,
						RowID:     priceBlock.RowID,
						Length:    maxLen,
					}
				}
			}
		}

		fmt.Printf("priceInfoMap:%+v", priceInfoMap)

		for price, info := range priceInfoMap {
			if info.Length >= number {
				if count >= offset && len(tickets) < pageSize {
					sectionName, err := venueService.GetSectionNameByID(info.SectionID)
					if err != nil {
						return err
					}
					rowName, err := venueService.GetRowNameByID(info.RowID)
					if err != nil {
						return err
					}

					ticket := Ticket{
						EventID:     eventID,
						SectionID:   info.SectionID,
						SectionName: sectionName,
						RowID:       info.RowID,
						RowName:     rowName,
						Price:       price,
						Length:      info.Length,
					}
					tickets = append(tickets, ticket)
				}
				count++
				if len(tickets) == pageSize {
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return []Ticket{}, err
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

	// Redis keys
	seatsKey := fmt.Sprintf("event:%d:section:%d:rows", msg.EventID, msg.SectionID)
	priceBlocksKey := fmt.Sprintf("event:%d:section:%d:price_blocks", msg.EventID, msg.SectionID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.redisClient.Watch(ctx, func(tx *redislib.Tx) error {
		// Step 1: Fetch and decode row data
		rowData, err := tx.HGet(ctx, seatsKey, fmt.Sprintf("%d", msg.RowID)).Result()
		if err == redislib.Nil {
			return fmt.Errorf("row not found")
		} else if err != nil {
			return fmt.Errorf("failed to get row data: %w", err)
		}

		var rowInfo struct {
			Seats string `json:"seats"`
		}
		if err := json.Unmarshal([]byte(rowData), &rowInfo); err != nil {
			return fmt.Errorf("failed to decode row data: %w", err)
		}

		seats := []rune(rowInfo.Seats) // Eg. ['0', '0', '1', '1', '1']
		seatCount := len(seats)
		consecutiveCount := 0
		reservedSeats := make(map[int]bool)
		startSeatNumber := 0

		// Find and reserve seats
		for i := 0; i < seatCount; i++ {
			if seats[i] == '0' { // Available
				consecutiveCount++
			} else {
				consecutiveCount = 0
			}

			if consecutiveCount == msg.Length {
				// Mark the seats as reserved (moving backwards)
				for j := 0; j < msg.Length; j++ {
					reservedSeats[i-j] = true
					seats[i-j] = '1'
					if j == msg.Length-1 {
						startSeatNumber = i - j + 1
					}
				}
				break
			}
		}

		if len(reservedSeats) == 0 {
			return fmt.Errorf("not enough consecutive seats available")
		}

		// calculate the new row data
		rowInfo.Seats = string(seats)
		updatedRowData, err := json.Marshal(rowInfo)
		if err != nil {
			return fmt.Errorf("failed to encode updated row data: %w", err)
		}

		// Get max consecutive lengths for price blocks
		priceBlocks, err := tx.ZRangeWithScores(ctx, priceBlocksKey, 0, -1).Result()
		if err != nil {
			return fmt.Errorf("failed to get price blocks: %w", err)
		}

		priceMaxConsecutive := map[int]int{}
		for _, block := range priceBlocks {
			member := block.Member.(string)
			price := int(block.Score)

			// Extract start and end seat numbers from the block key

			var rowID, startSeatID, startSeatNum, endSeatID, endSeatNum int
			_, err := fmt.Sscanf(member, "%d:%d:%d:%d:%d", &rowID, &startSeatID, &startSeatNum, &endSeatID, &endSeatNum)
			if err != nil {
				return fmt.Errorf("failed to parse price block: %w", err)
			}

			if rowID != msg.RowID {
				continue
			}

			// Calculate max consecutive available seats in this certain price block
			maxLength, currentLength := 0, 0
			for i := startSeatNum; i <= endSeatNum; i++ {
				if seats[i-1] == '0' {
					currentLength++
				} else {
					if currentLength > maxLength {
						maxLength = currentLength
					}
					currentLength = 0
				}
			}
			if currentLength > maxLength {
				maxLength = currentLength
			}
			priceMaxConsecutive[price] = maxLength
		}

		// Update data to redis : do this in the end to handle checks and preparations before
		_, err = tx.HSet(ctx, seatsKey, fmt.Sprintf("%d", msg.RowID), string(updatedRowData)).Result()
		if err != nil {
			return fmt.Errorf("failed to update seats data: %w", err)
		}

		// Add reservation in redis
		err = setReservation(ctx, tx, msg.SessionID, msg.EventID, msg.SectionID, msg.RowID, startSeatNumber, msg.Length)
		if err != nil {
			return err
		}

		// Notify WebSocket client and broadcast the reservation
		if err := s.NotifyReservation(msg); err != nil {
			log.Printf("failed to notify WebSocket client: %v", err)
		}

		if err := s.broadcastReservation(msg, priceMaxConsecutive); err != nil {
			log.Printf("failed to broadcast reservation: %v", err)
		}

		return err
	}, seatsKey, priceBlocksKey)

	return err
}

func (s *TicketService) NotifyReservation(msg dto.ReservationMsg) error {
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

func (s *TicketService) broadcastReservation(msg dto.ReservationMsg, priceMaxConsecutive map[int]int) error {
	var broadcastMsgs dto.BroadcastMsgs
	for price, length := range priceMaxConsecutive {
		broadcastMsg := dto.BroadcastMsg{
			EventID:   msg.EventID,
			SectionID: msg.SectionID,
			RowID:     msg.RowID,
			Price:     price,
			MaxLength: length,
		}
		broadcastMsgs.Messages = append(broadcastMsgs.Messages, broadcastMsg)
	}

	data, err := json.Marshal(broadcastMsgs)
	if err != nil {
		return fmt.Errorf("error marshaling broadcast message: %w", err)
	}

	return s.connectionManager.BroadcastReservation(data)
}
