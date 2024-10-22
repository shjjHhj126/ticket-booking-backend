package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"ticket-booking-backend/domain/venue"

	redislib "github.com/redis/go-redis/v9"
)

var redisSectionPriceName = "sections_by_price"

func getSectionsByPriceRange(ctx context.Context, tx *redislib.Tx, eventID, lowPrice, highPrice int) ([]int, error) {

	// Retrieve sections with a minPrice between lowPrice and +inf
	sectionData, err := tx.ZRangeByScoreWithScores(ctx, redisSectionPriceName, &redislib.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", float64(highPrice)),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get sections by price range: %w", err)
	}

	log.Printf("getSectionsByPriceRange, sectionData : %+v", sectionData)

	var sectionIDs []int
	for _, z := range sectionData {
		sectionInfo := strings.Split(z.Member.(string), ":")
		if len(sectionInfo) != 3 {
			return nil, fmt.Errorf("invalid section data format")
		}

		// Parse eventID, sectionID, and maxPrice
		fetchedEventID, err := strconv.Atoi(sectionInfo[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse event ID: %w", err)
		}
		if fetchedEventID != eventID {
			continue
		}

		sectionID, err := strconv.Atoi(sectionInfo[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse section ID: %w", err)
		}

		maxPrice, err := strconv.Atoi(sectionInfo[2])
		if err != nil {
			return nil, fmt.Errorf("failed to parse max price: %w", err)
		}

		if maxPrice >= lowPrice {
			sectionIDs = append(sectionIDs, sectionID)
		}
	}

	return sectionIDs, nil
}

// Load all the sections from a venue with price range
func cacheSections(ctx context.Context, tx *redislib.Tx, eventID, lowPrice, highPrice int, venueService *venue.VenueService) ([]int, error) {

	// Fetch sections from DB
	var sectionPriceRangeArray []venue.SectionPriceRange
	sectionPriceRangeArray, err := venueService.GetSectionIds(eventID)
	if err != nil {
		return []int{}, err
	}

	// Use a single Redis sorted set for storing all sections by minPrice
	for _, sectionPriceRange := range sectionPriceRangeArray {

		// Updated member format: {event_id}:{section_id}:{maxPrice}
		member := fmt.Sprintf("%d:%d:%d", eventID, sectionPriceRange.SectionID, sectionPriceRange.MaxPrice)
		if err := tx.ZAdd(ctx, redisSectionPriceName, redislib.Z{
			Score:  float64(sectionPriceRange.MinPrice),
			Member: member,
		}).Err(); err != nil {
			return nil, err
		}
	}

	return getSectionsByPriceRange(ctx, tx, eventID, lowPrice, highPrice)
}

func getSeatPriceBlocks(ctx context.Context, tx *redislib.Tx, eventID, sectionID, lowPrice, highPrice int) ([]venue.SeatPriceBlock, error) {
	redisKey := fmt.Sprintf("event:%d:section:%d:price_blocks", eventID, sectionID)

	priceBlockData, err := tx.ZRangeByScoreWithScores(ctx, redisKey, &redislib.ZRangeBy{
		Min: fmt.Sprintf("%d", lowPrice),
		Max: fmt.Sprintf("%d", highPrice),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get seat price blocks from cache: %w", err)
	}

	log.Printf("priceBlockData:%+v", priceBlockData)

	var seatPriceBlocks []venue.SeatPriceBlock
	for _, z := range priceBlockData {
		blockInfo := strings.Split(z.Member.(string), ":")
		if len(blockInfo) != 5 {
			return nil, fmt.Errorf("invalid seat block data format")
		}

		log.Printf("blockInfo:%+v", blockInfo)

		// Parse the rowID, startSeatID, startSeatNumber, endSeatID, endSeatNumber
		rowID, err := strconv.Atoi(blockInfo[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse row ID: %w", err)
		}
		startSeatID, err := strconv.Atoi(blockInfo[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse start seat id: %w", err)
		}
		startSeatNumber, err := strconv.Atoi(blockInfo[2])
		if err != nil {
			return nil, fmt.Errorf("failed to parse start seat number: %w", err)
		}
		endSeatID, err := strconv.Atoi(blockInfo[3])
		if err != nil {
			return nil, fmt.Errorf("failed to parse end seat id: %w", err)
		}
		endSeatNumber, err := strconv.Atoi(blockInfo[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse end seat number: %w", err)
		}

		// Create a SeatPriceBlock and add it to the result slice
		seatPriceBlock := venue.SeatPriceBlock{
			StartSeatNumber: startSeatNumber,
			StartSeatID:     startSeatID,
			EndSeatNumber:   endSeatNumber,
			EndSeatID:       endSeatID,
			RowID:           rowID,
			Price:           int(z.Score),
		}
		seatPriceBlocks = append(seatPriceBlocks, seatPriceBlock)
	}

	return seatPriceBlocks, nil
}

// Load all price block from a section in a venue of a event
func cacheSeatPriceBlocks(ctx context.Context, tx *redislib.Tx, eventID, sectionID, lowPrice, highPrice int, venueService *venue.VenueService) ([]venue.SeatPriceBlock, error) {
	seatPriceBlocks, err := venueService.GetSeatPriceBlocks(eventID, sectionID)
	if err != nil {
		return []venue.SeatPriceBlock{}, err
	}

	log.Printf("in cache seat price block, seatPriceBlocks:%+v\n", seatPriceBlocks)

	redisKey := fmt.Sprintf("event:%d:section:%d:price_blocks", eventID, sectionID)
	for _, block := range seatPriceBlocks {
		member := fmt.Sprintf("%d:%d:%d:%d:%d", block.RowID, block.StartSeatID, block.StartSeatNumber, block.EndSeatID, block.EndSeatNumber)
		if err := tx.ZAdd(ctx, redisKey, redislib.Z{
			Score:  float64(block.Price),
			Member: member,
		}).Err(); err != nil {
			return []venue.SeatPriceBlock{}, fmt.Errorf("failed to cache seat blocks: %w", err)
		}
	}

	return getSeatPriceBlocks(ctx, tx, eventID, sectionID, lowPrice, highPrice)
}

func getConsecutiveSeats(ctx context.Context,
	tx *redislib.Tx, eventID, sectionID int,
	block *venue.SeatPriceBlock) ([]venue.ConsecutiveSeats, error) {
	redisKey := fmt.Sprintf("event:%d:section:%d:rows", eventID, sectionID)
	rowID := block.RowID
	seatsKey := fmt.Sprintf("%d", rowID) // Field name for the row in the hash

	// Retrieve the seats JSON string from Redis
	seatsData, err := tx.HGet(ctx, redisKey, seatsKey).Result()
	if err != nil {
		if err == redislib.Nil { // not found
			return []venue.ConsecutiveSeats{}, nil
		}
		return nil, fmt.Errorf("error retrieving data from Redis: %w", err)
	}

	var seats map[string]interface{}
	if err := json.Unmarshal([]byte(seatsData), &seats); err != nil {
		return nil, fmt.Errorf("error unmarshaling seats data: %w", err)
	}

	seatStatuses, ok := seats["seats"].(string)
	if !ok {
		return nil, fmt.Errorf("error: seats data is not a string")
	}
	var consecutiveSeats []venue.ConsecutiveSeats

	// Split the seats string into a slice of strings
	availabilityList := strings.Split(seatStatuses, "")

	startSeat := -1
	count := 0
	for seatIndex, availability := range availabilityList {
		status, err := strconv.Atoi(availability)
		if err != nil {
			return nil, fmt.Errorf("error converting seat status to int: %w", err)
		}

		if status == 0 { // Seat is available
			if startSeat == -1 {
				startSeat = seatIndex // Mark the start of a new sequence
			}
			count++
		} else { // Seat is unavailable
			if startSeat != -1 {
				// Save current result
				consecutiveSeats = append(consecutiveSeats, venue.ConsecutiveSeats{
					RowID:   rowID,
					RowName: seats["row_name"].(string),
					Length:  count,
				})
				// reset
				startSeat = -1
				count = 0
			}
		}
	}

	// Handle the last result
	if startSeat != -1 {
		consecutiveSeats = append(consecutiveSeats, venue.ConsecutiveSeats{
			RowID:   rowID,
			RowName: seats["row_name"].(string),
			Length:  count,
		})
	}

	return consecutiveSeats, nil
}

func cacheConsecutiveSeats(ctx context.Context,
	tx *redislib.Tx, eventID, sectionID, rowID int,
	priceBlock *venue.SeatPriceBlock,
	venueService *venue.VenueService) ([]venue.ConsecutiveSeats, error) {
	rowCondition, err := venueService.GetRowConditionByID(rowID, eventID)
	if err != nil {
		return nil, err
	}

	redisKey := fmt.Sprintf("event:%d:section:%d:rows", eventID, sectionID)
	seatsAvailability := ""

	// Initialize all seats to "available" (status 0)
	for seatNumber := 0; seatNumber < len(rowCondition.SeatConditions); seatNumber++ {
		// "0":available, "1":non-available
		var value string
		if rowCondition.SeatConditions[seatNumber].BookedBy == nil {
			value = "0"
		} else {
			value = "1"
		}
		seatsAvailability += value
	}

	// Create the JSON-like structure
	rowData := fmt.Sprintf(`{"row_name": "%s", "seats": "%s"}`, rowCondition.RowName, seatsAvailability)

	// Cache the row's seat availability in Redis
	err = tx.HSet(ctx, redisKey, rowID, rowData).Err()
	if err != nil {
		return nil, err
	}

	log.Print("cache consecutive seat 1")

	return getConsecutiveSeats(ctx, tx, eventID, sectionID, priceBlock)
}

func readLuaScript(scriptPath string) (string, error) {
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Lua script: %w", err)
	}
	return string(content), nil
}
