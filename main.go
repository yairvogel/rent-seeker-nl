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
	flag.Parse()

	if *outputDir == "" {
		log.Fatal("Please provide an output directory using the -output flag")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// URL to fetch
	url := "https://www.pararius.nl/huurwoningen/utrecht/0-2500"

	// Fetch the page
	properties, err := FetchProperties(url)
	if err != nil {
		log.Fatalf("Error fetching properties: %v", err)
	}

	// Print the results and save new files
	fmt.Printf("Found %d properties in Utrecht under €2500\n", len(properties))

	newCount := 0
	for _, property := range properties {
		// Check if property already exists
		filename := filepath.Join(*outputDir, property.Hash+".json")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// This is a new property
			newCount++
			fmt.Printf("\nNEW PROPERTY: %s\n", property.Title)
			fmt.Printf("   Address: %s\n", property.Address)
			fmt.Printf("   Price: %d\n", property.PriceValue)
			fmt.Printf("   Size: %s\n", property.Size)
			fmt.Printf("   Rooms: %s\n", property.Rooms)
			fmt.Printf("   URL: %s\n", property.URL)
			fmt.Printf("   Hash: %s\n", property.Hash)

			// Save property to JSON file
			savePropertyToFile(property, *outputDir)
		}
	}

	fmt.Printf("\nSummary: %d new properties found out of %d total listings\n", newCount, len(properties))
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

