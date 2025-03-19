package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Property represents a rental property listing
type Property struct {
	Title       string
	Address     string
	Price       string
	PriceValue  float64 // Price as a numeric value
	Size        string
	Rooms       string
	Type        string
	URL         string
	Description string
}

func main() {
	// URL to fetch
	url := "https://www.pararius.nl/huurwoningen/utrecht/1000-2500/50m2"

	// Fetch the page
	properties, err := fetchProperties(url)
	if err != nil {
		log.Fatalf("Error fetching properties: %v", err)
	}

	// Print the results
	fmt.Printf("Found %d properties in Utrecht under €2500\n\n", len(properties))
	for i, property := range properties {
		fmt.Printf("%d. %s\n", i+1, property.Title)
		fmt.Printf("   Address: %s\n", property.Address)
		fmt.Printf("   Price: %s (€%.2f)\n", property.Price, property.PriceValue)
		fmt.Printf("   Size: %s\n", property.Size)
		fmt.Printf("   Rooms: %s\n", property.Rooms)
		fmt.Printf("   URL: %s\n\n", property.URL)
	}
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

		// Create property object
		property := Property{
			Title:      title,
			Address:    address,
			Price:      price,
			PriceValue: priceValue,
			Size:       size,
			Rooms:      rooms,
			URL:        url,
		}

		properties = append(properties, property)
	})

	return properties, nil
}

// extractPriceValue extracts the numeric price value from a price string
func extractPriceValue(priceStr string) float64 {
	// Remove non-numeric characters except for decimal point
	re := regexp.MustCompile(`[^0-9,.]`)
	numStr := re.ReplaceAllString(priceStr, "")
	
	// Replace comma with dot for decimal point (European format)
	numStr = strings.Replace(numStr, ",", ".", -1)
	
	// Parse the string to a float
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0.0
	}
	
	return value
}
