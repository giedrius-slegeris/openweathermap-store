package main

import (
	"context"
	"sync"
	"testing"
	"time"

	pb "github.com/giedrius-slegeris/proto-definitions-go/openweathermapstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// resetCache puts the package-level cache into the state main() would leave it
// in just after startup: allocated, but not yet filled by the scheduler.
func resetCache(t *testing.T) {
	t.Helper()
	oneCallCache = new(oneCallResults)
}

func TestGetWeatherDataUnavailableBeforeFirstFetch(t *testing.T) {
	resetCache(t)

	got, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err == nil {
		t.Fatal("GetWeatherData() returned nil error with an empty cache")
	}
	if got != nil {
		t.Errorf("GetWeatherData() returned %+v alongside an error, want nil", got)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error %v is not a gRPC status", err)
	}
	if st.Code() != codes.Unavailable {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unavailable)
	}
}

// The nil-cache guard has to be checked before RLock, or it panics instead of
// returning a status. This test only passes if that ordering holds.
func TestGetWeatherDataUnavailableWhenCacheNil(t *testing.T) {
	t.Cleanup(func() { oneCallCache = new(oneCallResults) })
	oneCallCache = nil

	got, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err == nil {
		t.Fatal("GetWeatherData() returned nil error with a nil cache")
	}
	if got != nil {
		t.Errorf("GetWeatherData() returned %+v alongside an error, want nil", got)
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("error %v is not a gRPC status", err)
	}
	if st.Code() != codes.Unavailable {
		t.Errorf("status code = %v, want %v", st.Code(), codes.Unavailable)
	}
}

func TestGetWeatherDataReturnsCachedResponse(t *testing.T) {
	resetCache(t)

	want := &pb.GetWeatherDataResponse{
		Lat:      51.5009,
		Lon:      -0.1246,
		Timezone: "Europe/London",
		Current:  &pb.Current{Temp: 18.5},
	}
	updateCache(want)

	got, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err != nil {
		t.Fatalf("GetWeatherData() returned error: %v", err)
	}
	if got.GetTimezone() != "Europe/London" {
		t.Errorf("Timezone = %q, want Europe/London", got.GetTimezone())
	}
	if got.GetCurrent().GetTemp() != 18.5 {
		t.Errorf("Current.Temp = %v, want 18.5", got.GetCurrent().GetTemp())
	}
}

func TestUpdateCacheStampsLastUpdated(t *testing.T) {
	resetCache(t)

	before := time.Now().UTC().Unix()
	updateCache(&pb.GetWeatherDataResponse{Timezone: "Europe/London"})
	after := time.Now().UTC().Unix()

	got, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err != nil {
		t.Fatalf("GetWeatherData() returned error: %v", err)
	}

	if ts := got.GetLastUpdated(); ts < before || ts > after {
		t.Errorf("LastUpdated = %d, want within [%d, %d]", ts, before, after)
	}
}

func TestUpdateCacheOverwritesPreviousResult(t *testing.T) {
	resetCache(t)

	updateCache(&pb.GetWeatherDataResponse{Timezone: "Europe/London"})
	updateCache(&pb.GetWeatherDataResponse{Timezone: "America/New_York"})

	got, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err != nil {
		t.Fatalf("GetWeatherData() returned error: %v", err)
	}
	if got.GetTimezone() != "America/New_York" {
		t.Errorf("Timezone = %q, want America/New_York", got.GetTimezone())
	}
}

// A response handed to a caller must not be mutated by a later refresh, since
// gRPC marshals it after the handler has released the read lock.
func TestReturnedResponseNotMutatedByLaterUpdate(t *testing.T) {
	resetCache(t)

	updateCache(&pb.GetWeatherDataResponse{Timezone: "Europe/London"})
	first, err := (&server{}).GetWeatherData(context.Background(), &pb.GetWeatherDataRequest{})
	if err != nil {
		t.Fatalf("GetWeatherData() returned error: %v", err)
	}
	firstStamp := first.GetLastUpdated()

	updateCache(&pb.GetWeatherDataResponse{Timezone: "America/New_York"})

	if first.GetTimezone() != "Europe/London" {
		t.Errorf("previously returned Timezone = %q, want it unchanged at Europe/London",
			first.GetTimezone())
	}
	if first.GetLastUpdated() != firstStamp {
		t.Errorf("previously returned LastUpdated changed from %d to %d",
			firstStamp, first.GetLastUpdated())
	}
}

// Run under -race: the scheduler writes while gRPC handlers read concurrently.
func TestCacheConcurrentReadWrite(t *testing.T) {
	resetCache(t)
	updateCache(&pb.GetWeatherDataResponse{Timezone: "Europe/London"})

	const (
		writers = 4
		readers = 16
		rounds  = 50
	)

	var wg sync.WaitGroup
	for range writers {
		wg.Go(func() {
			for range rounds {
				updateCache(&pb.GetWeatherDataResponse{Timezone: "Europe/London"})
			}
		})
	}
	for range readers {
		wg.Go(func() {
			for range rounds {
				if _, err := (&server{}).GetWeatherData(
					context.Background(), &pb.GetWeatherDataRequest{}); err != nil {
					t.Errorf("GetWeatherData() returned error: %v", err)
					return
				}
			}
		})
	}
	wg.Wait()
}
