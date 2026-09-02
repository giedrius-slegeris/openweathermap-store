package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	pb "github.com/giedrius-slegeris/proto-definitions-go/openweathermapstore"
	"google.golang.org/protobuf/encoding/protojson"
)

type OpenWeatherApi struct{}

func NewOpenWeatherAPI() *OpenWeatherApi {
	return &OpenWeatherApi{}
}

func (o *OpenWeatherApi) Get() (*pb.GetWeatherDataResponse, error) {
	apiURL, err := o.oneCallURL()
	if err != nil {
		return nil, err
	}

	// construct HTTP Client with timeouts
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	// Make the HTTP request
	resp, err := client.Get(apiURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch weather data: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Decode with protojson rather than encoding/json: the generated structs carry
	// PascalCase encoding/json tags, so snake_case fields from OpenWeatherMap
	// (timezone_offset, feels_like, wind_speed, ...) would silently stay at zero.
	// protojson honours the protobuf json names instead, and DiscardUnknown lets
	// the upstream add fields the proto does not declare without breaking us.
	weatherData := &pb.GetWeatherDataResponse{}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(body, weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return weatherData, nil
}

func (o *OpenWeatherApi) oneCallURL() (*url.URL, error) {
	// Create the full URL with query parameters
	requestURL, err := url.Parse(os.Getenv("OPEN_WEATHER_MAP_BASE_URL"))
	if err != nil {
		return nil, err
	}

	// Add query parameters
	params := url.Values{}
	params.Add("lat", os.Getenv("OPEN_WEATHER_MAP_LATITUDE"))
	params.Add("lon", os.Getenv("OPEN_WEATHER_MAP_LONGITUDE"))
	params.Add("appid", os.Getenv("OPEN_WEATHER_MAP_API_KEY"))
	params.Add("units", os.Getenv("OPEN_WEATHER_MAP_UNITS"))
	requestURL.RawQuery = params.Encode()

	return requestURL, nil
}
