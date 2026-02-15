package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type nominatimResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

type Location struct {
	Lat     float64
	Lon     float64
	Country string
}

func GeolocatePin(pincode string) (Location, error) {
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/search?postalcode=%s&format=json&limit=1",
		pincode,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Location{}, err
	}
	req.Header.Set("User-Agent", "starminal/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Location{}, fmt.Errorf("geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return Location{}, fmt.Errorf("failed to parse geocode response: %w", err)
	}

	if len(results) == 0 {
		return Location{}, fmt.Errorf("no location found for pincode %q", pincode)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return Location{}, fmt.Errorf("invalid latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return Location{}, fmt.Errorf("invalid longitude: %w", err)
	}

	country := ""
	parts := strings.Split(results[0].DisplayName, ", ")
	if len(parts) > 0 {
		country = parts[len(parts)-1]
	}

	return Location{Lat: lat, Lon: lon, Country: country}, nil
}
