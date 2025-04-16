package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
)

type WeatherMessage struct {
	Description string `json:"description"`
	Country     string `json:"country"`
	Weather     string `json:"weather"`
}

func Consume() {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	if bootstrapServers == "" {
		bootstrapServers = "weather-kafka-kafka-bootstrap.kafka.svc.cluster.local:9092"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis.redis.svc.cluster.local:6379"
	}

	redisPass := os.Getenv("REDIS_PASSWORD")
	if redisPass == "" {
		redisPass = "my-redis-pass"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": bootstrapServers,
		"group.id":          "weather-consumer-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer c.Close()

	err = c.SubscribeTopics([]string{"weather-topic"}, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Kafka and Redis successfully")
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			var weather WeatherMessage
			if err := json.Unmarshal(msg.Value, &weather); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			message := string(msg.Value)
			log.Printf("Received message: %s", message)

			// Store individual message
			key := fmt.Sprintf("weather:%d", time.Now().UnixNano())
			if err := rdb.Set(ctx, key, message, 0).Err(); err != nil {
				log.Printf("Failed to store in Redis: %v", err)
			} else {
				log.Printf("Stored in Redis with key: %s", key)
			}

			// Update country counter
			countryKey := "country:counts"
			if err := rdb.HIncrBy(ctx, countryKey, weather.Country, 1).Err(); err != nil {
				log.Printf("Failed to update country counter: %v", err)
			}

			// Update total messages
			totalKey := "messages:total"
			if err := rdb.Incr(ctx, totalKey).Err(); err != nil {
				log.Printf("Failed to update total messages: %v", err)
			}

			log.Printf("Updated country: %s, total messages: %s", weather.Country, totalKey)
		} else {
			log.Printf("Consumer error: %v (%v)", err, msg)
		}
	}
}
