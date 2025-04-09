package service

import (
	"context"
	"encoding/json"
	"go-api/internal/kafka"
	pb "go-api/internal/pb"
	"go-api/internal/rabbitmq"
	"log"
)

type WeatherServer struct {
	pb.UnimplementedWeatherServiceServer
}

func (s *WeatherServer) ProcessTweet(ctx context.Context, tweet *pb.WeatherTweet) (*pb.WeatherResponse, error) {

	jsonData, _ := json.Marshal(tweet)

	// Publish to Kafka
	if err := kafka.Publish("weather-topic", jsonData); err != nil {
		log.Printf("Kafka error: %v", err)
	}

	// Publish to RabbitMQ
	if err := rabbitmq.Publish("weather-queue", jsonData); err != nil {
		log.Printf("RabbitMQ error: %v", err)
	}

	return &pb.WeatherResponse{Success: true}, nil
}
