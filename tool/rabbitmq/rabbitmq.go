package rabbitmq

import (
	"log"
	"os"

	"ticket-booking-backend/tool/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

var (
	defaultRabbitMQURL = os.Getenv("RABBIT_MQ_URL")
	defaultQueueName   = os.Getenv("QUEUE_NAME")
)

func InitRabbitMQ() *RabbitMQ {
	rabbitMQURL := util.GetEnvOrDefault("RABBIT_MQ_URL", defaultRabbitMQURL)
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.Fatal(err)
		return nil
	}

	queueName := util.GetEnvOrDefault("QUEUE_NAME", defaultQueueName)

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		log.Fatal(err)
		return nil
	}

	return &RabbitMQ{
		conn:      conn,
		channel:   ch,
		queueName: queueName,
	}
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQ) PublishMessage(body []byte) error {
	return r.channel.Publish(
		"",          // exchange
		r.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}
