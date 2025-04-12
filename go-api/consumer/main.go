package main

import (
	"log"

	"go-api/internal/kafka" // Replace with your actual module name, e.g., github.com/yourusername/project2/go-api
)

func main() {
	log.Println("Starting Kafka consumer...")
	kafka.Consume()
}
