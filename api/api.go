package api

import (
	"encoding/json"
	"fmt"
	pb "github.com/giedrius-slegeris/proto-definitions/openweathermap-store"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
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

	weatherData := &pb.GetWeatherDataResponse{}
	err = json.Unmarshal(body, &weatherData)
	if err != nil {
		return nil, err
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
