package main

import (
	"flag"
	"github.com/newrelic/go-agent/v3/newrelic"
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
	disableJob := flag.Bool("disable-job", false, "Disable periodic property checks")
	newrelicKey := flag.String("newrelic-key", "", "newrelic key")
	flag.Parse()

	newRelicApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName("earlybird-back"),
		newrelic.ConfigLicense(*newrelicKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)

	if err != nil {
		log.Fatalf("failed initializing newrelic app: %v", err)
	}

	// Only require outputDir if job is enabled
	if !*disableJob && *outputDir == "" {
		log.Fatal("Please provide an output directory using the -output flag")
	}

	if !*disableJob && *telegramToken == "" {
		log.Fatal("Please provide a Telegram Bot API token using the -token flag")
	}

	// Create output directory if it doesn't exist and job is enabled
	if !*disableJob {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
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

	// Start the HTTP server
	go RunHTTPServer(*httpPort, "", newRelicApp)
	log.Printf("HTTP server started on port %s", *httpPort)

	// Start periodic property checks if not disabled
	if !*disableJob {
		log.Println("Starting periodic property checks...")
		go RunPeriodicPropertyChecks(searchUrls, *outputDir, bot)
	} else {
		log.Println("Periodic property checks are disabled")
	}

	// Keep the program running
	select {}
}
