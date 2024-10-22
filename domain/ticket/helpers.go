package ticket

import (
	"context"
	"ticket-booking-backend/domain/venue"

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
