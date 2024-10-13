package payment

// import (
// 	"encoding/json"
// 	"log"

// 	"github.com/jmoiron/sqlx"
// 	amqp "github.com/rabbitmq/amqp091-go"
// 	"github.com/redis/go-redis/v9"
// )

// // StartWorker initializes the RabbitMQ worker for processing payment confirmations.
// func StartWorker(rmq *amqp.Connection, rdb *redis.Client, db *sqlx.DB) {
// 	ch, err := rmq.Channel()
// 	if err != nil {
// 		log.Fatalf("Failed to open a channel: %v", err)
// 	}
// 	defer ch.Close()

// 	msgs, err := ch.Consume("payment_confirmations", "", false, false, false, false, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to register a consumer: %v", err)
// 	}

// 	for d := range msgs {
// 		var paymentConfirmation PaymentConfirmation // Define this struct according to your needs
// 		if err := json.Unmarshal(d.Body, &paymentConfirmation); err != nil {
// 			log.Printf("Error decoding payment confirmation message: %v", err)
// 			d.Nack(false, false) // Not acknowledging the message means it will be requeued
// 			continue
// 		}

// 		// Process the payment confirmation
// 		err = processPaymentConfirmation(paymentConfirmation, rdb, db)
// 		if err != nil {
// 			log.Printf("Error processing payment confirmation: %v", err)
// 			d.Nack(false, true) // Requeue the message if there was an error
// 		} else {
// 			d.Ack(false) // Acknowledge the message if processed successfully
// 		}
// 	}
// }

// // processPaymentConfirmation processes the payment confirmation logic.
// // You might want to confirm the ticket and cancel any temporary reservations.
// func processPaymentConfirmation(paymentConfirmation PaymentConfirmation, rdb *redis.Client, db *sqlx.DB) error {
// 	// Implement logic to confirm the ticket and handle the reservation cancellation if needed
// 	ticketID := paymentConfirmation.TicketID
// 	userID := paymentConfirmation.UserID

// 	// Confirm ticket in the database
// 	// Remove temporary reservation from Redis
// 	return nil
// }

// // Define the PaymentConfirmation struct based on the expected message structure
// type PaymentConfirmation struct {
// 	UserID   string `json:"user_id"`
// 	TicketID string `json:"ticket_id"`
// 	// Add more fields as needed
// }
