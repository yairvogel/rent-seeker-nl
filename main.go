package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
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
	properties, err := fetchProperties(url)
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

func fetchProperties(url string) ([]Property, error) {
	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
	}

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var properties []Property

	// Find all property listings
	doc.Find("section").Each(func(i int, s *goquery.Selection) {
		// Extract property information
		titleElem := s.Find("h2")
		title := strings.TrimSpace(titleElem.Text())

		// Skip if not a property listing
		if title == "" {
			return
		}

		// Get property URL
		url, _ := titleElem.Find("a").Attr("href")
		if url != "" {
			url = "https://www.pararius.nl" + url
		}

		// Get address
		address := strings.TrimSpace(s.Find("div.listing-search-item__location").Text())

		// Get price
		price := strings.TrimSpace(s.Find("div.listing-search-item__price").Text())
		priceValue := extractPriceValue(price)

		// Get property details
		var size, rooms string
		s.Find("li.illustrated-features__item").Each(func(i int, feat *goquery.Selection) {
			text := strings.TrimSpace(feat.Text())
			if strings.Contains(text, "m²") {
				size = text
			} else if strings.Contains(text, "kamer") {
				rooms = text
			}
		})

		// Create hash of URL for deduplication
		urlHash := generateHash(url)

		// Create property object
		property := Property{
			Title:      title,
			Address:    address,
			PriceValue: priceValue,
			Size:       size,
			Rooms:      rooms,
			URL:        url,
			Hash:       urlHash,
		}

		properties = append(properties, property)
	})

	return properties, nil
}

// extractPriceValue extracts the numeric price value from a price string
func extractPriceValue(priceStr string) int32 {
	// Remove non-numeric characters except for decimal point
	re := regexp.MustCompile(`[^0-9,.]`)
	numStr := re.ReplaceAllString(priceStr, "")

	numStr = strings.Replace(numStr, ".", "", -1)

	// Parse the string to a int32
	value, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		return 0
	}

	return int32(value)
}

// generateHash creates a SHA-256 hash of the input string
func generateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
