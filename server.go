package main

import (
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go/v82"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/price"
	"github.com/stripe/stripe-go/v82/webhook"
	"io"
	"log"
	"net/http"
	"os"
)

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
}

func createPortalSession(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello, world!")
}

// createCheckoutSession handles the POST request to create a new subscription
func createCheckoutSession(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("New subscription: %+v", subscriptionData)

	// Create stripe subscription
	var checkoutParams *stripe.CheckoutSessionParams

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
	stripe.Key = os.Getenv("STRIPE_KEY")

	// Define the create-subscription endpoint with CORS support
	http.HandleFunc("/create-checkout-session", enableCORS(createCheckoutSession))
	http.HandleFunc("/create-portal-session", enableCORS(createPortalSession))

	// Start the server in a goroutine
	log.Printf("Starting HTTP server on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
