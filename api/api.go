package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type OpenWeatherApi struct{}

func NewOpenWeatherAPI() *OpenWeatherApi {
	return &OpenWeatherApi{}
}

func (o *OpenWeatherApi) Get() (*OneCallResp, error) {
	apiURL, err := o.oneCallURL()
	if err != nil {
		return nil, err
	}

	// Make the HTTP request
	resp, err := http.Get(apiURL.String())
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

	weatherData := &OneCallResp{}
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
