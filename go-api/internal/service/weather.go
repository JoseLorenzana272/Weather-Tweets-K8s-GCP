package service

import (
	"context"
	"log"

	pb "go-api/internal/pb"
)

type WeatherServer struct {
	pb.UnimplementedWeatherServiceServer
}

func (s *WeatherServer) ProcessTweet(
	ctx context.Context,
	tweet *pb.WeatherTweet,
) (*pb.WeatherResponse, error) {
	log.Printf("gRPC Server received: %v", tweet)
	return &pb.WeatherResponse{
		Success: true,
		Message: "Tweet processed",
	}, nil
}
