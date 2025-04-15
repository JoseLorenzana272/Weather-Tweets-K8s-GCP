package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func Consume() {
	rabbitMQAddr := os.Getenv("RABBITMQ_ADDR")
	if rabbitMQAddr == "" {
		rabbitMQAddr = "amqp://guest:guest@rabbitmq.rabbitmq.svc.cluster.local:5672/"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey-primary.default.svc.cluster.local:6379"
	}

	valkeyPass := os.Getenv("VALKEY_PASSWORD")
	if valkeyPass == "" {
		valkeyPass = "valkey-pass"
	}

	conn, err := amqp091.Dial(rabbitMQAddr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	queueName := "weather-queue"
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // autoAck
		false,     // exclusive
		false,     // noLocal
		false,     // noWait
		nil,       // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     valkeyAddr,
		Password: valkeyPass,
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}

	log.Printf("Connected to RabbitMQ and Valkey successfully")
	for msg := range msgs {
		message := string(msg.Body)
		log.Printf("Received message: %s", message)
		key := fmt.Sprintf("weather:%d", time.Now().UnixNano())
		err = rdb.Set(ctx, key, message, 0).Err()
		if err != nil {
			log.Printf("Failed to store in Valkey: %v", err)
		} else {
			log.Printf("Stored in Valkey with key: %s", key)
		}
	}
}
