package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FetchProperties fetches and parses property listings from the given URL
func FetchProperties(url string) ([]Property, error) {
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
