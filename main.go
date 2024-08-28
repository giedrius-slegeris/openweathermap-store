package main

import (
	"fmt"
	"giedrius-slegeris/openweathermap-store/api"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file. Using default environment variables.")
	}

	owAPI := api.NewOpenWeatherAPI()
	resp, err := owAPI.Get()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("DATA: %+v\n", resp)
}
