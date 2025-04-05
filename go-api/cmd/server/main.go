package main

import (
	"log"
	"net/http"

	"go-api/internal/handler"
)

func main() {
	// HTTP server
	http.HandleFunc("/process", handler.HandleWeatherTweet)
	log.Println("Go API running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
