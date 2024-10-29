package redisexpirescanner

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"ticket-booking-backend/domain/payment"

	redislib "github.com/redis/go-redis/v9"
)

type ExpirationScanner struct {
	redisClient    *redislib.Client
	scanInterval   time.Duration
	paymentService *payment.PaymentService
}

func NewExpirationScanner(redisClient *redislib.Client, scanInterval time.Duration, paymentService *payment.PaymentService) *ExpirationScanner {
	return &ExpirationScanner{
		redisClient:    redisClient,
		scanInterval:   scanInterval,
		paymentService: paymentService,
	}
}

func (s *ExpirationScanner) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.scanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.scanAndProcess(ctx); err != nil {
				log.Printf("Error scanning keys: %v", err)
			}
		}
	}
}

func (s *ExpirationScanner) scanAndProcess(ctx context.Context) error {
	var cursor uint64

	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, "reservation:*", 100).Result()
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if err := s.processKeys(ctx, keys); err != nil {
			log.Printf("Error processing keys: %v", err)
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return nil
}

// processKeys checks TTL for each key and processes expired ones.
func (s *ExpirationScanner) processKeys(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if s.isExpired(ctx, key) {
			if err := s.processKey(ctx, key); err != nil {
				log.Printf("Error processing key %s: %v", key, err)
			}
		}
	}
	return nil
}

func (s *ExpirationScanner) isExpired(ctx context.Context, key string) bool {
	ttl, err := s.redisClient.TTL(ctx, key).Result()
	if err != nil {
		log.Printf("Error checking TTL for %s: %v", key, err)
		return false
	}
	return ttl < 0
}

// processKey handles expired keys, adding them to the processing set, updating seat availability, and deleting the reservation.
func (s *ExpirationScanner) processKey(ctx context.Context, key string) error {
	return s.redisClient.Watch(ctx, func(tx *redislib.Tx) error {
		// Check if the key is already being processed
		if exists, err := tx.SIsMember(ctx, "processing_keys", key).Result(); err != nil {
			log.Printf("Error checking processing set: %v", err)
			return err
		} else if exists {
			return nil // Skip already processed keys
		}

		// Add to processing set
		if err := s.addToProcessingSet(ctx, tx, key); err != nil {
			return err
		}

		// Cancel paymentintent
		if err := s.cancelPaymentIntent(ctx, tx, key); err != nil {
			return err
		}

		// Update seat availability within the transaction
		if err := s.updateSeatAvailability(ctx, tx, key); err != nil {
			log.Printf("Error updating seat availability for key %s: %v", key, err)
			s.removeFromProcessingSet(ctx, tx, key) // Clean up processing set on failure
			return err
		}

		// Delete the reservation key within the transaction
		if err := tx.Del(ctx, key).Err(); err != nil {
			log.Printf("Error deleting reservation key %s: %v", key, err)
			return fmt.Errorf("error deleting reservation data: %w", err)
		}

		return nil
	}, key)
}

// addToProcessingSet adds the key to the processing set to avoid duplicate processing.
func (s *ExpirationScanner) addToProcessingSet(ctx context.Context, tx *redislib.Tx, key string) error {
	if err := tx.SAdd(ctx, "processing_keys", key).Err(); err != nil {
		return fmt.Errorf("error adding to processing set: %w", err)
	}

	// Set expiration on the processing set (done once per day)
	if err := tx.Expire(ctx, "processing_keys", 24*time.Hour).Err(); err != nil {
		log.Printf("Failed to set expiration on processing set: %v", err)
	}
	return nil
}

func (s *ExpirationScanner) removeFromProcessingSet(ctx context.Context, tx *redislib.Tx, key string) {
	if err := tx.SRem(ctx, "processing_keys", key).Err(); err != nil {
		log.Printf("Failed to remove key from processing set: %v", err)
	}
}

// updateSeatAvailability updates seat availability based on the reservation information.
func (s *ExpirationScanner) updateSeatAvailability(ctx context.Context, tx *redislib.Tx, key string) error {
	reservationInfo, err := tx.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("error retrieving reservation info: %w", err)
	}

	// Parse reservation info
	infoArr := strings.Split(reservationInfo, ":")
	if len(infoArr) != 9 {
		return fmt.Errorf("invalid reservation format for key %s", key)
	}

	// Extract required details
	startSeatNum, err := strconv.Atoi(infoArr[6])
	if err != nil {
		return fmt.Errorf("cannot convert string to int: %w", err)
	}

	length, err := strconv.Atoi(infoArr[7])
	if err != nil {
		return fmt.Errorf("cannot convert string to int: %w", err)
	}

	eventID, sectionID, rowID := infoArr[3], infoArr[4], infoArr[5]

	// Update row seats within the same transaction
	return s.updateRowSeats(ctx, tx, eventID, sectionID, rowID, startSeatNum, length)
}

// updateRowSeats updates the seat statuses for a specific row in the event section.
func (s *ExpirationScanner) updateRowSeats(ctx context.Context, tx *redislib.Tx, eventID, sectionID, rowID string, startSeatNum, length int) error {
	rowKey := fmt.Sprintf("event:%s:section:%s:rows", eventID, sectionID)

	// Get the row data within the transaction
	rowData, err := tx.HGet(ctx, rowKey, rowID).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil // Row not found, nothing to update
		}
		return fmt.Errorf("error retrieving row data from Redis: %w", err)
	}

	// Parse row information
	var rowInfo struct {
		RowName string `json:"row_name"`
		Seats   string `json:"seats"`
	}
	if err := json.Unmarshal([]byte(rowData), &rowInfo); err != nil {
		return fmt.Errorf("failed to decode row data: %w", err)
	}

	// Modify the seat statuses
	newStatuses := []rune(rowInfo.Seats)
	for i := startSeatNum - 1; i < startSeatNum+length-1; i++ {
		newStatuses[i] = '0'
	}
	rowInfo.Seats = string(newStatuses)

	// Serialize the updated row information
	updatedRowData, err := json.Marshal(rowInfo)
	if err != nil {
		return fmt.Errorf("failed to encode updated row data: %w", err)
	}

	// Set the updated row data back within the transaction
	if _, err := tx.HSet(ctx, rowKey, rowID, string(updatedRowData)).Result(); err != nil {
		return fmt.Errorf("failed to update seats data: %w", err)
	}
	fmt.Printf("new row data in scanner:%s", string(updatedRowData))

	return nil
}

func (s *ExpirationScanner) cancelPaymentIntent(ctx context.Context, tx *redislib.Tx, key string) error {
	// Get the reservation info within the transaction
	reservationInfo, err := tx.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("error retrieving reservation info: %w", err)
	}

	// Parse reservation info
	infoArr := strings.Split(reservationInfo, ":")
	if len(infoArr) != 9 {
		return fmt.Errorf("invalid reservation format for key %s", key)
	}

	// Extract the payment intent ID
	paymentIntentID := infoArr[0]

	// Call the payment service to cancel the payment intent
	err = s.paymentService.CancelPaymentIntent(paymentIntentID)
	if err != nil {
		return fmt.Errorf("error cancelling payment intent %s: %w", paymentIntentID, err)
	}

	log.Printf("Successfully cancelled payment intent: %s", paymentIntentID)
	return nil
}
