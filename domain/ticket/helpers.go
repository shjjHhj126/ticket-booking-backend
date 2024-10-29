package ticket

import (
	"context"
	"fmt"
	"strconv"
	"ticket-booking-backend/domain/venue"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

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

func getConsecutiveSeatBlocks(ctx context.Context, tx *redislib.Tx, eventID, sectionID int, venueService *venue.VenueService, priceBlock *venue.SeatPriceBlock) (string, string, error) {
	consecutiveSeats, rowName, err := getConsecutiveSeats(ctx, tx, eventID, sectionID, priceBlock.RowID)
	if err != nil {
		return "", "", err
	}

	if len(consecutiveSeats) == 0 {
		consecutiveSeats, rowName, err = cacheConsecutiveSeats(ctx, tx, eventID, sectionID, priceBlock.RowID, priceBlock, venueService)
		if err != nil {
			return "", "", err
		}
	}
	return consecutiveSeats, rowName, nil
}

func setReservation(ctx context.Context, tx *redislib.Tx, sessionID string, eventID, sectionID, rowID, startSeatNumber, length, price int, reservationID string) error {
	key := fmt.Sprintf("reservation:%s", reservationID)
	value := fmt.Sprintf("%s:%d:%d:%d:%d:%d:%d", sessionID, eventID, sectionID, rowID, startSeatNumber, length, price)

	// Set the reservation with 5 minutes time out
	if err := tx.Set(ctx, key, value, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to set reservation: %w", err)
	}

	return nil
}

func ConvertReservationRecordAtoi(eventIDStr, sectionIDStr, rowIDStr, startSeatNumberStr, lengthStr, priceStr string) (int, int, int, int, int, int, error) {
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid eventID form in redis, error:%w", err)
	}
	sectionID, err := strconv.Atoi(sectionIDStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid sectionID form in redis, error:%w", err)
	}
	rowID, err := strconv.Atoi(rowIDStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid rowID form in redis, error:%w", err)
	}
	startSeatNumber, err := strconv.Atoi(startSeatNumberStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid startSeatNumber form in redis, error:%w", err)
	}
	Length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid length form in redis, error:%w", err)
	}
	Price, err := strconv.Atoi(priceStr)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("invalid price form in redis, error:%w", err)
	}
	return eventID, sectionID, rowID, startSeatNumber, Length, Price, nil
}
