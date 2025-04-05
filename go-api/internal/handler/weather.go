package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WeatherTweet struct {
	Description string `json:"description"`
	Country     string `json:"country"`
	Weather     string `json:"weather"`
}

func HandleWeatherTweet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var tweet WeatherTweet
	if err := json.NewDecoder(r.Body).Decode(&tweet); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid JSON: %v", err)
		return
	}

	// Log received tweet (will forward via gRPC later)
	fmt.Printf("Received tweet: %+v\n", tweet)

	// TODO: Call gRPC service (next step)
	// grpcClient.PublishToKafka(tweet)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tweet)
}
