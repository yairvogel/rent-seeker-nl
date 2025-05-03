package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// SubscriptionData represents the subscription data sent from the frontend
type SubscriptionData struct {
	Cities     []string `json:"cities"`
	PriceRange []int    `json:"priceRange"`
	LivingArea []int    `json:"livingArea"`
	Email      string   `json:"email"`
	Amount     int      `json:"amount"`
}

// handleCreateSubscription handles the POST request to create a new subscription
func handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	// Parse the JSON data
	var subscriptionData SubscriptionData
	if err := json.Unmarshal(body, &subscriptionData); err != nil {
		http.Error(w, "Error parsing JSON data", http.StatusBadRequest)
		return
	}

	// Validate the data
	if len(subscriptionData.Cities) == 0 {
		http.Error(w, "At least one city must be selected", http.StatusBadRequest)
		return
	}
	if len(subscriptionData.PriceRange) != 2 {
		http.Error(w, "Price range must have min and max values", http.StatusBadRequest)
		return
	}
	if len(subscriptionData.LivingArea) != 2 {
		http.Error(w, "Living area must have min and max values", http.StatusBadRequest)
		return
	}
	if subscriptionData.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// TODO: Save the subscription data to a database or file
	log.Printf("New subscription: %+v", subscriptionData)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"message":   "Subscription created successfully",
		"sessionId": "SampleSessionId",
	})
}

// RunHTTPServer starts a simple HTTP server on the specified port
func RunHTTPServer(port string) {
	// Define the root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!")
	})

	// Define the create-subscription endpoint
	http.HandleFunc("/create-subscription", handleCreateSubscription)

	// Start the server in a goroutine
	log.Printf("Starting HTTP server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
