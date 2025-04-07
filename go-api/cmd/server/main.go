package main

import (
	"log"
	"net"
	"net/http"

	"go-api/internal/handler"
	pb "go-api/internal/pb"
	"go-api/internal/service"

	"google.golang.org/grpc"
)

func main() {
	// Start gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterWeatherServiceServer(grpcServer, &service.WeatherServer{})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("gRPC failed to listen: %v", err)
	}
	go func() {
		log.Println("gRPC server running on :50051")
		grpcServer.Serve(lis)
	}()

	// Start HTTP server
	http.HandleFunc("/process", handler.HandleWeatherTweet)
	log.Println("HTTP server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
