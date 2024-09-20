package main

import (
	"context"
	pb "github.com/giedrius-slegeris/proto-definitions/openweathermap-store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) GetWeatherData(_ context.Context, _ *pb.GetWeatherDataRequest) (*pb.GetWeatherDataResponse, error) {
	oneCallCache.RLock()
	defer oneCallCache.RUnlock()
	if oneCallCache == nil || oneCallCache.resp == nil {
		return nil, status.Errorf(codes.Unavailable, "Weather data unavailable")
	}
	return oneCallCache.resp, nil
}
