package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"ticket-booking-backend/tool/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn           *amqp.Connection //TCP connection
	bookingChannel *amqp.Channel    //virtual connection, used to send and receive messages.
	paymentChannel *amqp.Channel
}

// queues are buffers "inside" channel

var (
	defaultRabbitMQURL      = os.Getenv("RABBIT_MQ_URL")
	defaultBookingQueueName = os.Getenv("BOOKING_QUEUE_NAME")
	defaultPaymentQueueName = os.Getenv("PAYMENT_QUEUE_NAME")
)

func InitRabbitMQ() *RabbitMQ {
	rabbitMQURL := util.GetEnvOrDefault("RABBIT_MQ_URL", defaultRabbitMQURL)
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Initialize channels and declare queues
	bookingChannel := createChannel(conn, "booking")
	paymentChannel := createChannel(conn, "payment")

	// Declare queues
	bookingQueueName := util.GetEnvOrDefault("BOOKING_QUEUE_NAME", defaultBookingQueueName)
	declareQueue(bookingChannel, bookingQueueName)

	paymentQueueName := util.GetEnvOrDefault("PAYMENT_QUEUE_NAME", defaultPaymentQueueName)
	declareQueue(paymentChannel, paymentQueueName)

	return &RabbitMQ{
		conn:           conn,
		bookingChannel: bookingChannel,
		paymentChannel: paymentChannel,
	}
}

func (r *RabbitMQ) Close() error {
	if r.bookingChannel != nil {
		if err := r.bookingChannel.Close(); err != nil {
			return err
		}
	}
	if r.paymentChannel != nil {
		if err := r.paymentChannel.Close(); err != nil {
			return err
		}
	}

	// Close the connection if it exists
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (r *RabbitMQ) PublishMessage(action string, body []byte) error {
	var channel *amqp.Channel
	var queueName string

	switch action {
	case "book":
		channel = r.bookingChannel
		queueName = "booking-queue"
	case "pay":
		channel = r.paymentChannel
		queueName = "payment-queue"
	default:
		log.Printf(`Unknown action : %s, should be either "book" or "pay" `, action)
		return fmt.Errorf("unknown action name in publish message : %s", action)
	}

	err := channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish message to queue %s: %v", queueName, err)
		return err
	}

	return nil
}

func (r *RabbitMQ) ConsumeMessages(action string, handler func([]byte) error) error {
	var channel *amqp.Channel
	var queueName string

	switch action {
	case "book":
		channel = r.bookingChannel
		queueName = "booking-queue"
	case "pay":
		channel = r.paymentChannel
		queueName = "payment-queue"

	default:
		log.Printf(`Unknown action : %s, should be either "book" or "pay" `, action)
		return fmt.Errorf("unknown action name in publish message : %s", action)
	}

	msgs, err := channel.Consume( //set the consumer of the channel and consume the messages in msgs
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (manual ack for reliability)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("Failed to process message: %v", err)
				if err := msg.Nack(false, true); err != nil { // second param in msg.Nack() means requeue
					log.Printf("Failed to NACK message: %v", err)
				}
			} else {
				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to ACK message: %v", err)
				}
			}
		}
	}()

	return nil
}

func createChannel(conn *amqp.Connection, channelName string) *amqp.Channel {
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open %s channel: %v", channelName, err)
	}
	return channel
}

// declareQueue declares a queue with the given name on the specified channel
func declareQueue(channel *amqp.Channel, queueName string) {
	_, err := channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare %s queue: %v", queueName, err)
	}
}
