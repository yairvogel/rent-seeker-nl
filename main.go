package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	log.Println("Bot started successfully. Press Ctrl+C to exit.")
	
	// Keep the program running
	select {}
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
