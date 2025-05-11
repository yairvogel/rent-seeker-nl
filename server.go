package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/price"
)

func RunHTTPServer(port, stripeKey string, newRelicApplication *newrelic.Application) {
	// stripe.Key = stripeKey

	// Define the create-subscription endpoint with CORS support
	http.HandleFunc(newrelic.WrapHandleFunc(newRelicApplication, "/create-checkout-session", enableCORS(createCheckoutSession)))
	http.HandleFunc(newrelic.WrapHandleFunc(newRelicApplication, "/payment/{sessionId}", enableCORS(verifyPayment)))

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
}

func verifyPayment(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodGet && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionId := r.PathValue("sessionId")

	// Validate session ID
	if sessionId == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	// Retrieve the checkout session from Stripe
	s, err := session.Get(sessionId, nil)
	if err != nil {
		log.Printf("[VerifyPayment][%s]Error retrieving session: %v", sessionId, err)
		encoder.Encode(VerifyPaymentResponse{
			Status: "failed",
			Error:  "Error retrieving session data",
		})
		return
	}

	// Check if payment was successful
	if s.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		log.Printf("[VerifyPayment][%s] Payment not completed", sessionId)
		encoder.Encode(VerifyPaymentResponse{
			Status: "failed",
			Error:  "payment noot completed",
		})
		return
	}

	// Extract subscription data from metadata
	subscription := s.Subscription
	if subscription == nil {
		log.Printf("[VerifyPayment][%s] No subscription found", sessionId)
		encoder.Encode(VerifyPaymentResponse{
			Status: "failed",
			Error:  "No subscription found",
		})
		return
	}

	// Return success response with subscription details
	encoder.Encode(VerifyPaymentResponse{
		Status:     "success",
		AccessCode: "some-random-stuff",
	})
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

	log.Printf("New subscription: %+v", subscriptionData)
	var lookupKey string
	switch len(subscriptionData.Cities) {
	case 1:
		lookupKey = "single_city"
	case 2:
		lookupKey = "two_cities"
	case 3:
		lookupKey = "three_cities"
	case 4:
		lookupKey = "four_cities"
	default:
		return
	}

	domain := "http://localhost:8080"

	priceListParams := &stripe.PriceListParams{
		LookupKeys: stripe.StringSlice([]string{lookupKey}),
	}

	i := price.List(priceListParams)
	var price *stripe.Price
	for i.Next() {
		p := i.Price()
		price = p
	}

	// Create stripe checkout session
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(domain + "/payment-success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(domain + "/cancel"),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{{Price: stripe.String(price.ID),
			Quantity: stripe.Int64(1),
		}},
		CustomerEmail:    stripe.String(subscriptionData.Email),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		http.Error(w, "Error creating checkout session", http.StatusInternalServerError)
		return
	}

	// Return session ID and URL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"redirectUrl": s.URL,
	})
}
