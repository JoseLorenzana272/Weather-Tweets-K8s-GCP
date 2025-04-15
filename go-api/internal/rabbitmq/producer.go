package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

func Publish(queueName string, message []byte) error {
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")
	if rabbitMQAddr == "" {
		rabbitMQAddr = "amqp://guest:guest@rabbitmq.rabbitmq.svc.cluster.local:5672/"
	}

	conn, err := amqp091.Dial(rabbitMQAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	log.Printf("Published to RabbitMQ: %s", message)
	return nil
}
