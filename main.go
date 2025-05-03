package main

import (
	"flag"
	"log"
	"os"
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
	httpPort := flag.String("port", "8080", "HTTP server port")
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
	searchUrls := map[string][]string{
		"Utrecht":   {"https://www.pararius.nl/huurwoningen/utrecht"},
		"Amsterdam": {"https://www.pararius.nl/huurwoningen/amsterdam"},
		"Den-haag":  {"https://www.pararius.nl/huurwoningen/den-haag"},
		"Rotterdam": {"https://www.pararius.nl/huurwoningen/rotterdam"},
	}

	// Start periodic property checks
	go RunPeriodicPropertyChecks(searchUrls, *outputDir, bot)
	
	// Start HTTP server
	StartHTTPServer(*httpPort)
	
	log.Printf("HTTP server started on port %s", *httpPort)

	// Keep the program running
	select {}
}
