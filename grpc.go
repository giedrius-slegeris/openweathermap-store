package main

import (
	"context"
	pb "github.com/giedrius-slegeris/proto-definitions/openweathermap-store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) GetWeatherData(_ context.Context, _ *pb.GetWeatherDataRequest) (*pb.GetWeatherDataResponse, error) {
	if oneCallCache == nil || oneCallCache.resp == nil {
		return nil, status.Errorf(codes.Unavailable, "Weather data unavailable")
	}
	return oneCallCache.resp, nil
}
