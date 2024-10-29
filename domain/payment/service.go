package payment

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"ticket-booking-backend/dto"
	"ticket-booking-backend/tool/rabbitmq"
	"time"

	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/paymentintent"

	redislib "github.com/redis/go-redis/v9"
)

type PaymentService struct {
	mq           *rabbitmq.RabbitMQ
	redisClient  *redislib.Client
	StripeAPIKey string
	repo         *PaymentRepository
}

func NewPaymentService(redisClient *redislib.Client, rmq *rabbitmq.RabbitMQ, db *sql.DB) *PaymentService {
	return &PaymentService{
		mq:           rmq,
		redisClient:  redisClient,
		StripeAPIKey: os.Getenv("STRIPE_API_KEY"),
		repo:         NewPaymentRepository(db),
	}
}

func (s *PaymentService) CreatePaymentIntent(ctx context.Context, tx *redislib.Tx, amount int, sessionID, reservationID string) (string, error) {
	stripe.Key = s.StripeAPIKey

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount) * 100),
		Currency: stripe.String(string(stripe.CurrencyTWD)),
		PaymentMethodTypes: []*string{
			stripe.String("card"),
		},
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create payment intent: %w", err)
	}

	fmt.Printf("client secret:%s\n", intent.ClientSecret)
	fmt.Printf("paymentintent_id:%s\n", intent.ID)

	reservationInfo, err := tx.Get(ctx, "reservation:"+reservationID).Result()
	if err != nil {
		if err == redislib.Nil { // not found
			return "", fmt.Errorf("session / reservation not found: %w", err)
		}
		return "", fmt.Errorf("error retrieving data from Redis: %w", err)
	}
	reservationInfo = intent.ID + ":" + intent.ClientSecret + ":" + reservationInfo

	// add the paymentintent info into the record
	// Todo: 5*time.Minute -> changeable when develop
	if err := tx.Set(ctx, "reservation:"+reservationID, reservationInfo, 5*time.Minute).Err(); err != nil {
		return "", fmt.Errorf("failed to set paymentintent in reservation: %w", err)
	}

	return intent.ClientSecret, nil
}

func (s *PaymentService) CancelPaymentIntent(paymentIntentID string) error {
	stripe.Key = s.StripeAPIKey

	params := &stripe.PaymentIntentCancelParams{}
	result, err := paymentintent.Cancel(paymentIntentID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel payment intent: %w", err)
	}

	fmt.Printf("Payment Intent %s has been canceled: %v\n", result.ID, result.Status)
	return nil
}

func (s *PaymentService) GetPaymentIntentStatus(paymentIntentID string) (stripe.PaymentIntentStatus, error) {
	stripe.Key = s.StripeAPIKey

	params := &stripe.PaymentIntentParams{}
	intent, err := paymentintent.Get(paymentIntentID, params)
	if err != nil {
		return "", fmt.Errorf("failed to get payment intent: %w", err)
	}

	fmt.Printf("Payment Intent %s has status: %v\n", intent.ID, intent.Status)
	return intent.Status, nil
}

func (s *PaymentService) PaymentIntentIsSucceeded(status stripe.PaymentIntentStatus) bool {
	return status == stripe.PaymentIntentStatusSucceeded
}

func (s *PaymentService) HandleBookedMessage(data []byte) error {
	bookInfo := dto.DbBookDTO{}
	if err := json.Unmarshal(data, &bookInfo); err != nil {
		return fmt.Errorf("error unmarshaling db book info:%w", err)
	}
	//Todo: find record in redis, expire it
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// update redis record
	if _, err := s.redisClient.Del(ctx, "reservation:"+bookInfo.ReservationID).Result(); err != nil {
		return fmt.Errorf("error deleting redis record:%w", err)
	}

	return s.repo.PipelineInsertData(bookInfo)
}
