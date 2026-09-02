package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// setEnv points the client at a stub server and fills in the location config.
func setEnv(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("OPEN_WEATHER_MAP_BASE_URL", baseURL)
	t.Setenv("OPEN_WEATHER_MAP_API_KEY", "test-key")
	t.Setenv("OPEN_WEATHER_MAP_LATITUDE", "51.500937")
	t.Setenv("OPEN_WEATHER_MAP_LONGITUDE", "-0.124602")
	t.Setenv("OPEN_WEATHER_MAP_UNITS", "metric")
}

// oneCallBody is a trimmed but structurally faithful One Call 3.0 response.
const oneCallBody = `{
  "lat": 51.5009,
  "lon": -0.1246,
  "timezone": "Europe/London",
  "timezone_offset": 3600,
  "current": {
    "dt": 1726000000,
    "temp": 18.5,
    "feels_like": 18.1,
    "pressure": 1012,
    "humidity": 72,
    "wind_speed": 3.6,
    "wind_deg": 210,
    "weather": [
      {"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}
    ]
  }
}`

func TestOneCallURL(t *testing.T) {
	setEnv(t, "https://api.openweathermap.org/data/3.0/onecall")

	got, err := (&OpenWeatherApi{}).oneCallURL()
	if err != nil {
		t.Fatalf("oneCallURL() returned error: %v", err)
	}

	if got.Scheme != "https" {
		t.Errorf("scheme = %q, want https", got.Scheme)
	}
	if got.Host != "api.openweathermap.org" {
		t.Errorf("host = %q, want api.openweathermap.org", got.Host)
	}
	if got.Path != "/data/3.0/onecall" {
		t.Errorf("path = %q, want /data/3.0/onecall", got.Path)
	}

	wantParams := map[string]string{
		"lat":   "51.500937",
		"lon":   "-0.124602",
		"appid": "test-key",
		"units": "metric",
	}
	q := got.Query()
	for k, want := range wantParams {
		if q.Get(k) != want {
			t.Errorf("query %q = %q, want %q", k, q.Get(k), want)
		}
	}
}

func TestOneCallURLInvalidBaseURL(t *testing.T) {
	setEnv(t, "http://%zz")

	if _, err := (&OpenWeatherApi{}).oneCallURL(); err == nil {
		t.Fatal("oneCallURL() returned nil error for an unparseable base URL")
	}
}

// The API key must reach the upstream as a query parameter, or every call 401s.
func TestGetSendsConfiguredQueryParams(t *testing.T) {
	var gotQuery map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = map[string]string{}
		for k := range r.URL.Query() {
			gotQuery[k] = r.URL.Query().Get(k)
		}
		_, _ = w.Write([]byte(oneCallBody))
	}))
	defer srv.Close()

	setEnv(t, srv.URL)

	if _, err := (&OpenWeatherApi{}).Get(); err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	want := map[string]string{
		"lat":   "51.500937",
		"lon":   "-0.124602",
		"appid": "test-key",
		"units": "metric",
	}
	for k, v := range want {
		if gotQuery[k] != v {
			t.Errorf("upstream received %q = %q, want %q", k, gotQuery[k], v)
		}
	}
}

func TestGetDecodesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(oneCallBody))
	}))
	defer srv.Close()

	setEnv(t, srv.URL)

	got, err := (&OpenWeatherApi{}).Get()
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if got.GetLat() != 51.5009 {
		t.Errorf("Lat = %v, want 51.5009", got.GetLat())
	}
	if got.GetLon() != -0.1246 {
		t.Errorf("Lon = %v, want -0.1246", got.GetLon())
	}
	if got.GetTimezone() != "Europe/London" {
		t.Errorf("Timezone = %q, want Europe/London", got.GetTimezone())
	}
	if got.GetCurrent().GetTemp() != 18.5 {
		t.Errorf("Current.Temp = %v, want 18.5", got.GetCurrent().GetTemp())
	}
	if got.GetCurrent().GetHumidity() != 72 {
		t.Errorf("Current.Humidity = %v, want 72", got.GetCurrent().GetHumidity())
	}
	if w := got.GetCurrent().GetWeather(); len(w) != 1 || w[0].GetMain() != "Clouds" {
		t.Errorf("Current.Weather = %+v, want one entry with Main=Clouds", w)
	}
}

// TestGetDecodesSnakeCaseFields guards a regression that silently zeroed data.
//
// OpenWeatherMap sends snake_case field names, but the generated structs carry
// PascalCase encoding/json tags (`json:"TimezoneOffset"`). encoding/json matches
// case-insensitively yet does not ignore underscores, so decoding with it left
// every snake_case field at its zero value with no error returned. Eight fields
// were affected: timezone_offset, feels_like, dew_point, wind_speed, wind_deg,
// wind_gust, moon_phase, sender_name.
//
// api.go decodes with protojson, which honours the protobuf json names. Swapping
// back to encoding/json will fail this test.
func TestGetDecodesSnakeCaseFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(oneCallBody))
	}))
	defer srv.Close()

	setEnv(t, srv.URL)

	got, err := (&OpenWeatherApi{}).Get()
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if got.GetTimezoneOffset() != 3600 {
		t.Errorf("TimezoneOffset = %v, want 3600", got.GetTimezoneOffset())
	}
	if got.GetCurrent().GetFeelsLike() != 18.1 {
		t.Errorf("Current.FeelsLike = %v, want 18.1", got.GetCurrent().GetFeelsLike())
	}
	if got.GetCurrent().GetWindSpeed() != 3.6 {
		t.Errorf("Current.WindSpeed = %v, want 3.6", got.GetCurrent().GetWindSpeed())
	}
	if got.GetCurrent().GetWindDeg() != 210 {
		t.Errorf("Current.WindDeg = %v, want 210", got.GetCurrent().GetWindDeg())
	}
}

// fullOneCallBody mirrors a complete One Call 3.0 response, including shapes
// protojson has to tolerate: fields the proto does not declare at all
// (current.rain and hourly.rain arrive as objects, daily.summary as a string),
// alongside daily.rain which the proto does declare, as a number.
const fullOneCallBody = `{
  "lat": 51.5009, "lon": -0.1246,
  "timezone": "Europe/London", "timezone_offset": 3600,
  "current": {
    "dt": 1726000000, "sunrise": 1725950000, "sunset": 1726000000,
    "temp": 18.5, "feels_like": 18.1, "pressure": 1012, "humidity": 72,
    "dew_point": 13.2, "uvi": 4.2, "clouds": 75, "visibility": 10000,
    "wind_speed": 3.6, "wind_deg": 210, "wind_gust": 7.2,
    "weather": [{"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}],
    "rain": {"1h": 0.25}
  },
  "minutely": [{"dt": 1726000060, "precipitation": 0.1}],
  "hourly": [{
    "dt": 1726000000, "temp": 18.5, "feels_like": 18.1, "pressure": 1012,
    "humidity": 72, "dew_point": 13.2, "uvi": 4.2, "clouds": 75,
    "visibility": 10000, "wind_speed": 3.6, "wind_deg": 210, "wind_gust": 7.2,
    "weather": [{"id": 803, "main": "Clouds", "description": "broken clouds", "icon": "04d"}],
    "pop": 0.32, "rain": {"1h": 0.25}
  }],
  "daily": [{
    "dt": 1726000000, "sunrise": 1725950000, "sunset": 1726000000,
    "moonrise": 1725960000, "moonset": 1725990000, "moon_phase": 0.75,
    "summary": "Expect a day of partly cloudy with rain",
    "temp": {"day": 18.5, "min": 12.1, "max": 20.3, "night": 13.0, "eve": 17.2, "morn": 12.8},
    "feels_like": {"day": 18.1, "night": 12.6, "eve": 16.9, "morn": 12.3},
    "pressure": 1012, "humidity": 72, "dew_point": 13.2,
    "wind_speed": 3.6, "wind_deg": 210, "wind_gust": 7.2,
    "weather": [{"id": 500, "main": "Rain", "description": "light rain", "icon": "10d"}],
    "clouds": 75, "pop": 0.32, "rain": 1.24, "uvi": 4.2
  }],
  "alerts": [{
    "sender_name": "Met Office", "event": "Yellow Wind Warning",
    "start": 1726000000, "end": 1726100000,
    "description": "Strong winds expected", "tags": ["Wind"]
  }]
}`

// A complete upstream payload must decode without error. protojson is stricter
// than encoding/json, so this guards against a real response being rejected
// outright rather than merely losing a field.
func TestGetDecodesFullPayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(fullOneCallBody))
	}))
	defer srv.Close()

	setEnv(t, srv.URL)

	got, err := (&OpenWeatherApi{}).Get()
	if err != nil {
		t.Fatalf("Get() rejected a full upstream payload: %v", err)
	}

	if got.GetTimezoneOffset() != 3600 {
		t.Errorf("TimezoneOffset = %v, want 3600", got.GetTimezoneOffset())
	}
	if got.GetCurrent().GetDewPoint() != 13.2 {
		t.Errorf("Current.DewPoint = %v, want 13.2", got.GetCurrent().GetDewPoint())
	}
	if got.GetCurrent().GetWindGust() != 7.2 {
		t.Errorf("Current.WindGust = %v, want 7.2", got.GetCurrent().GetWindGust())
	}
	if n := len(got.GetMinutely()); n != 1 {
		t.Errorf("len(Minutely) = %d, want 1", n)
	}
	if n := len(got.GetHourly()); n != 1 {
		t.Errorf("len(Hourly) = %d, want 1", n)
	}

	daily := got.GetDaily()
	if len(daily) != 1 {
		t.Fatalf("len(Daily) = %d, want 1", len(daily))
	}
	if daily[0].GetMoonPhase() != 0.75 {
		t.Errorf("Daily[0].MoonPhase = %v, want 0.75", daily[0].GetMoonPhase())
	}
	if daily[0].GetRain() != 1.24 {
		t.Errorf("Daily[0].Rain = %v, want 1.24", daily[0].GetRain())
	}
	if daily[0].GetTemp().GetMax() != 20.3 {
		t.Errorf("Daily[0].Temp.Max = %v, want 20.3", daily[0].GetTemp().GetMax())
	}
	if daily[0].GetFeelsLike().GetNight() != 12.6 {
		t.Errorf("Daily[0].FeelsLike.Night = %v, want 12.6", daily[0].GetFeelsLike().GetNight())
	}

	alerts := got.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("len(Alerts) = %d, want 1", len(alerts))
	}
	if alerts[0].GetSenderName() != "Met Office" {
		t.Errorf("Alerts[0].SenderName = %q, want Met Office", alerts[0].GetSenderName())
	}
}

func TestGetNon200(t *testing.T) {
	for _, tc := range []struct {
		name string
		code int
		want string
	}{
		{"unauthorized", http.StatusUnauthorized, "401"},
		{"too many requests", http.StatusTooManyRequests, "429"},
		{"server error", http.StatusInternalServerError, "500"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.code)
			}))
			defer srv.Close()

			setEnv(t, srv.URL)

			got, err := (&OpenWeatherApi{}).Get()
			if err == nil {
				t.Fatalf("Get() returned nil error for HTTP %d", tc.code)
			}
			if got != nil {
				t.Errorf("Get() returned %+v alongside an error, want nil", got)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to mention %q", err, tc.want)
			}
		})
	}
}

func TestGetMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"lat": not-json`))
	}))
	defer srv.Close()

	setEnv(t, srv.URL)

	if _, err := (&OpenWeatherApi{}).Get(); err == nil {
		t.Fatal("Get() returned nil error for a malformed body")
	} else if _, ok := err.(*json.SyntaxError); !ok {
		t.Logf("error was %v (%T)", err, err)
	}
}

func TestGetUnreachableUpstream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close() // nothing is listening now

	setEnv(t, url)

	if _, err := (&OpenWeatherApi{}).Get(); err == nil {
		t.Fatal("Get() returned nil error when the upstream was unreachable")
	}
}
