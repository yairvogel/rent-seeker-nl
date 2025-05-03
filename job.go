package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// runPeriodicPropertyChecks runs property checks every 10 minutes
func RunPeriodicPropertyChecks(cities map[string][]string, outputDir string, bot *TelegramBot) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Run once immediately
	for cityName, urls := range cities {
		checkForNewPropertiesInCity(cityName, urls, outputDir, bot)
	}

	// Then run on ticker
	for range ticker.C {
		for cityName, urls := range cities {
			checkForNewPropertiesInCity(cityName, urls, outputDir, bot)
		}
	}
}

func checkForNewPropertiesInCity(cityName string, urls []string, outputDir string, bot *TelegramBot) {
	log.Printf("Checking for new properties in %s...", cityName)
	for _, url := range urls {
		checkForNewProperties(cityName, url, outputDir, bot)
	}
}

// checkForNewProperties checks for new properties and notifies subscribers
func checkForNewProperties(cityName, url, outputDir string, bot *TelegramBot) {
	log.Printf("\tChecking for new properties in %s...", url)

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

	for _, prop := range newProperties {
		bot.NotifySubscribers(prop)
	}
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
