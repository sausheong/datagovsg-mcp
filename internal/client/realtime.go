package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const realtimeBaseURL = "https://api.data.gov.sg/v1"

var validEnvironmentTypes = map[string]bool{
	"air-temperature":   true,
	"rainfall":          true,
	"relative-humidity": true,
	"wind-speed":        true,
	"wind-direction":    true,
	"uv-index":          true,
	"psi":               true,
	"pm25":              true,
}

var validForecastTypes = map[string]string{
	"2-hour":  "2-hour-weather-forecast",
	"24-hour": "24-hour-weather-forecast",
	"4-day":   "4-day-weather-forecast",
}

var validTransportTypes = map[string]string{
	"taxi":            "taxi-availability",
	"carpark":         "carpark-availability",
	"traffic-cameras": "traffic-images",
}

type RealtimeClient struct {
	http    *http.Client
	baseURL string
}

func NewRealtimeClient() *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: realtimeBaseURL,
	}
}

func NewRealtimeClientWithConfig(baseURL string, timeout time.Duration) *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: timeout},
		baseURL: baseURL,
	}
}

func (c *RealtimeClient) doGet(url string) ([]byte, error) {
	var (
		body []byte
		err  error
	)
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
		req, reqErr := http.NewRequest(http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("User-Agent", userAgent)

		resp, doErr := c.http.Do(req)
		if doErr != nil {
			err = doErr
			continue
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, apiError(resp.StatusCode, url)
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}
		body = buf.Bytes()
		break
	}
	return body, err
}

func prettyJSON(data []byte) (string, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return "", err
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c *RealtimeClient) GetEnvironmentReading(readingType string) (string, error) {
	if !validEnvironmentTypes[readingType] {
		valid := make([]string, 0, len(validEnvironmentTypes))
		for k := range validEnvironmentTypes {
			valid = append(valid, k)
		}
		return "", fmt.Errorf("invalid environment type %q; valid types: %v", readingType, valid)
	}
	url := fmt.Sprintf("%s/environment/%s", c.baseURL, readingType)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}

func (c *RealtimeClient) GetWeatherForecast(forecastType string) (string, error) {
	endpoint, ok := validForecastTypes[forecastType]
	if !ok {
		return "", fmt.Errorf("invalid forecast type %q; valid types: 2-hour, 24-hour, 4-day", forecastType)
	}
	url := fmt.Sprintf("%s/environment/%s", c.baseURL, endpoint)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}

func (c *RealtimeClient) GetTransportInfo(infoType string) (string, error) {
	endpoint, ok := validTransportTypes[infoType]
	if !ok {
		return "", fmt.Errorf("invalid transport type %q; valid types: taxi, carpark, traffic-cameras", infoType)
	}
	url := fmt.Sprintf("%s/transport/%s", c.baseURL, endpoint)
	data, err := c.doGet(url)
	if err != nil {
		return "", err
	}
	return prettyJSON(data)
}
