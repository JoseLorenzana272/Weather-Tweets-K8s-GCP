package main

import (
	"log"

	"go-api/internal/rabbitmq"
)

func main() {
	log.Println("Starting RabbitMQ consumer...")
	rabbitmq.Consume()
}
