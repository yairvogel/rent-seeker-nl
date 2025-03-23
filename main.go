package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Property represents a rental property listing
type Property struct {
	Title       string
	Address     string
	PriceValue  int32 // Price as a numeric value
	Size        string
	Rooms       string
	Type        string
	URL         string
	Description string
	Hash        string // Hash of URL for deduplication
}

func main() {
	// Parse command line arguments
	outputDir := flag.String("output", "", "Directory to save property JSON files")
	telegramToken := flag.String("token", "", "Telegram Bot API token")
	flag.Parse()

	if *outputDir == "" {
		log.Fatal("Please provide an output directory using the -output flag")
	}

	if *telegramToken == "" {
		log.Fatal("Please provide a Telegram Bot API token using the -token flag")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Initialize Telegram bot
	bot, err := NewTelegramBot(*telegramToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
	}

	// Start the bot in a goroutine
	go bot.Start()

	log.Println("Bot started successfully. Property checks will run every 5 minutes.")

	// Define search URLs
	searchURLs := []string{
		"https://www.pararius.nl/huurwoningen/utrecht/1000-2500/50m2",
		"https://www.pararius.nl/huurwoningen/haarlem/1000-2500/50m2",
		"https://www.pararius.nl/huurwoningen/leiden/1000-2500/50m2",
	}

	// Start periodic property checks
	go runPeriodicPropertyChecks(searchURLs, *outputDir, bot)

	// Keep the program running
	select {}
}

// runPeriodicPropertyChecks runs property checks every 10 minutes
func runPeriodicPropertyChecks(urls []string, outputDir string, bot *TelegramBot) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Run once immediately
	for _, url := range urls {
		checkForNewProperties(url, outputDir, bot)
	}

	// Then run on ticker
	for range ticker.C {
		for _, url := range urls {
			checkForNewProperties(url, outputDir, bot)
		}
	}
}

// checkForNewProperties checks for new properties and notifies subscribers
func checkForNewProperties(url, outputDir string, bot *TelegramBot) {
	// Extract city name from URL for logging
	cityName := "unknown"
	if city := extractCityFromURL(url); city != "" {
		cityName = city
	}
	log.Printf("Checking for new properties in %s...", cityName)

	// Fetch properties
	properties, err := FetchProperties(url)
	if err != nil {
		log.Printf("Error fetching properties: %v", err)
		return
	}

	log.Printf("Found %d properties in %s", len(properties), cityName)

	// Check for new properties
	var newProperties []Property
	for _, property := range properties {
		filename := filepath.Join(outputDir, property.Hash+".json")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// This is a new property
			newProperties = append(newProperties, property)

			// Save property to file
			savePropertyToFile(property, outputDir)
		}
	}

	// Notify subscribers about new properties
	if len(newProperties) > 0 {
		log.Printf("Found %d new properties", len(newProperties))

		// Only notify if there are subscribers
		if bot.HasSubscribers() {

			// Send individual property details
			for _, prop := range newProperties {
				message := formatPropertyMessage(prop)
				bot.NotifySubscribers(message)
			}
		} else {
			log.Println("No subscribers to notify")
		}
	} else {
		log.Println("No new properties found")
	}
}

// formatPropertyMessage formats a property as a message for Telegram
func formatPropertyMessage(property Property) string {
	var priceStr string
	if property.PriceValue > 0 {
		priceStr = fmt.Sprintf("€%d", property.PriceValue)
	} else {
		priceStr = "Price unknown"
	}

	message := fmt.Sprintf("📍*%s*\n", property.Title)
	if property.Address != "" {
		message += fmt.Sprintf(" %s\n", property.Address)
	}
	message += fmt.Sprintf("💰 %s\n", priceStr)
	if property.Size != "" {
		message += fmt.Sprintf("📏 %s\n", property.Size)
	}
	if property.Rooms != "" {
		message += fmt.Sprintf("🚪 %s\n", property.Rooms)
	}
	message += fmt.Sprintf("🔗 [View on Pararius](%s)", property.URL)

	return message
}

// extractCityFromURL extracts the city name from a Pararius URL
func extractCityFromURL(url string) string {
	// Simple extraction based on URL format
	// Example: https://www.pararius.nl/huurwoningen/utrecht/1000-2500/50m2
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "huurwoningen" && i+1 < len(parts) {
			return strings.Title(parts[i+1])
		}
	}
	return ""
}

// savePropertyToFile saves a property as a JSON file
func savePropertyToFile(property Property, outputDir string) {
	// Create JSON data
	jsonData, err := json.MarshalIndent(property, "", "  ")
	if err != nil {
		log.Printf("Error creating JSON for property %s: %v", property.Title, err)
		return
	}

	// Create filename using hash to ensure uniqueness
	filename := filepath.Join(outputDir, property.Hash+".json")

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		log.Printf("Error writing property to file %s: %v", filename, err)
		return
	}

	fmt.Printf("   Saved to %s\n", filename)
}
