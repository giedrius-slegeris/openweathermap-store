package main

import (
	"fmt"
	"giedrius-slegeris/openweathermap-store/api"
	"giedrius-slegeris/openweathermap-store/cron"
	"github.com/joho/godotenv"
	"log"
	"sync"
	"time"
)

var (
	oneCallCache *oneCallResults
)

type oneCallResults struct {
	sync.Mutex
	resp           *api.OneCallResp
	lastUpdatedUTC time.Time
}

func main() {
	oneCallCache = new(oneCallResults)

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file. Using default environment variables.")
	}

	owAPI := api.NewOpenWeatherAPI()

	// wrap API call with a callback function to update cache, this is by design to enforce separation of concerns
	// in functions and testing
	run := func() {
		resp, err := owAPI.Get()
		if err != nil {
			log.Printf("Failed to fetch one call results, %s", err)
			return
		}
		updateCache(resp)
	}

	if err = cron.StartTaskAsync(run); err != nil {
		fmt.Printf("Failed to start cron task: %s\n", err)
	}

	// temporary measure to keep the server alive
	for {
		time.Sleep(5 * time.Minute)
	}
}

func updateCache(resp *api.OneCallResp) {
	log.Printf("Updating cache with new API results")
	oneCallCache.Lock()
	defer oneCallCache.Unlock()
	oneCallCache.resp = resp
	oneCallCache.lastUpdatedUTC = time.Now().UTC()
}
