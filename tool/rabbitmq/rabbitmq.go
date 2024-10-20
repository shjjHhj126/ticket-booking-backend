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

	// Create separate channels for booking and payment queues
	bookingChannel, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open booking channel:", err)
	}

	paymentChannel, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open payment channel:", err)
	}

	// Declare booking queue on its own channel
	bookingQueueName := util.GetEnvOrDefault("BOOKING_QUEUE_NAME", defaultBookingQueueName)
	_, err = bookingChannel.QueueDeclare(
		bookingQueueName, true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare booking queue:", err)
	}

	// Declare payment queue on its own channel
	paymentQueueName := util.GetEnvOrDefault("PAYMENT_QUEUE_NAME", defaultPaymentQueueName)
	_, err = paymentChannel.QueueDeclare(
		paymentQueueName, true, false, false, false, nil,
	)
	if err != nil {
		log.Fatal("Failed to declare payment queue:", err)
	}

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

func (r *RabbitMQ) PublishMessage(queueName string, body []byte) error {
	var channel *amqp.Channel

	if queueName == util.GetEnvOrDefault("BOOKING_QUEUE_NAME", defaultBookingQueueName) {
		channel = r.bookingChannel
	} else if queueName == util.GetEnvOrDefault("PAYMENT_QUEUE_NAME", defaultPaymentQueueName) {
		channel = r.paymentChannel
	} else {
		log.Printf("Unknown queue name: %s", queueName)
		return fmt.Errorf("unknown queue name: %s", queueName)
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

func (r *RabbitMQ) ConsumeMessages(queueName string, handler func([]byte) error) error {
	var channel *amqp.Channel

	if queueName == util.GetEnvOrDefault("BOOKING_QUEUE_NAME", defaultBookingQueueName) {
		channel = r.bookingChannel
	} else if queueName == util.GetEnvOrDefault("PAYMENT_QUEUE_NAME", defaultPaymentQueueName) {
		channel = r.paymentChannel
	} else {
		log.Printf("Unknown queue name: %s", queueName)
		return fmt.Errorf("unknown queue name: %s", queueName)
	}

	msgs, err := channel.Consume(
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
				if err := msg.Nack(false, true); err != nil {
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
