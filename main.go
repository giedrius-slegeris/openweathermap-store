package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file. Using default environment variables.")
	}

	// Create the full URL with query parameters
	requestURL, err := url.Parse(os.Getenv("OPEN_WEATHER_MAP_BASE_URL"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add query parameters
	params := url.Values{}
	params.Add("lat", os.Getenv("OPEN_WEATHER_MAP_LATITUDE"))
	params.Add("lon", os.Getenv("OPEN_WEATHER_MAP_LONGITUDE"))
	params.Add("appid", os.Getenv("OPEN_WEATHER_MAP_API_KEY"))
	params.Add("units", os.Getenv("OPEN_WEATHER_MAP_UNITS"))
	requestURL.RawQuery = params.Encode()

	// Make the HTTP request
	resp, err := http.Get(requestURL.String())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("io.ReadAll failed with %s\n", err)
		return
	}

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch weather data: %s\n", resp.Status)
	}

	// Parse the JSON response
	weatherData := &oneCallResp{}
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		fmt.Println(err)
	}

	// Print the weather data
	fmt.Printf("DATA: %+v\n", weatherData)
}
