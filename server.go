package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func RunHTTPServer(port, stripeKey string, newRelicApplication *newrelic.Application) {

	// Define the create-subscription endpoint with CORS support
	http.HandleFunc(newrelic.WrapHandleFunc(newRelicApplication, "/create-subscription", enableCORS(createSubscription)))

	// Start the server in a goroutine
	log.Printf("Starting HTTP server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

type VerifyPaymentResponse struct {
	Status     string `json:"status"`
	AccessCode string `json:"accessCode"`
	Error      string `json:"error"`
}

// enableCORS is middleware that adds CORS headers to responses
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}

// SubscriptionData represents the subscription data sent from the frontend
type SubscriptionData struct {
	Cities     []string `json:"cities"`
	PriceRange []int    `json:"priceRange"`
	LivingArea []int    `json:"livingArea"`
	Email      string   `json:"email"`
	Id         string   `json:"id"`
	Name       string   `json:"name"`
}

// createSubscription handles the POST request to create a new subscription
func createSubscription(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
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

	if len(subscriptionData.Cities) > 4 {
		http.Error(w, "Too many cities in selection", http.StatusBadRequest)
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

	filename := "./subscriptions/" + subscriptionData.Id + ".json"
	f, err := os.Create(filename)
	if err != nil {
		log.Printf("[ERROR] failed to create a file at %s. err: %v", filename, err)
		http.Error(w, "Error creating checkout session", http.StatusInternalServerError)
	}

	json.NewEncoder(f).Encode(subscriptionData)

	f.Close()

	accessCode := RandString(6)
	filename = "./accesscodes/" + accessCode + ".txt"
	f, err = os.Create(filename)
	if err != nil {
		log.Printf("[ERROR] failed to create a file at %s. err: %v", filename, err)
		http.Error(w, "Error creating checkout session", http.StatusInternalServerError)
	}

	f.WriteString(subscriptionData.Id)

	f.Close()

	log.Printf("New subscription: %+v", subscriptionData)

	// Return session ID and URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"accessToken": accessCode,
	})
}
