package main

import (
	"fmt"
	"giedrius-slegeris/openweathermap-store/api"
	"giedrius-slegeris/openweathermap-store/cron"
	pb "github.com/giedrius-slegeris/proto-definitions/openweathermapstore"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var (
	s            = &server{}
	oneCallCache *oneCallResults
)

type server struct {
	pb.UnimplementedOpenWeatherMapStoreServerServer
}

type oneCallResults struct {
	sync.RWMutex
	resp *pb.GetWeatherDataResponse
}

func main() {
	oneCallCache = new(oneCallResults)
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

	if err := cron.StartTaskAsync(run); err != nil {
		fmt.Printf("Failed to start cron task: %s\n", err)
	}

	// start GRPC server
	listenPort := os.Getenv("GRPC_LISTEN_PORT")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", listenPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	gs := grpc.NewServer()
	pb.RegisterOpenWeatherMapStoreServerServer(gs, s)
	fmt.Println("Server is listening on port " + listenPort)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func updateCache(resp *pb.GetWeatherDataResponse) {
	log.Printf("Updating cache with new API results")
	oneCallCache.Lock()
	defer oneCallCache.Unlock()
	oneCallCache.resp = resp
	oneCallCache.resp.LastUpdated = time.Now().UTC().Unix()
}
