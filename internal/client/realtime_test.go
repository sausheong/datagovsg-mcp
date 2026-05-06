package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestRealtimeClient(serverURL string) *RealtimeClient {
	return &RealtimeClient{
		http:    &http.Client{Timeout: 5 * time.Second},
		baseURL: serverURL + "/v1",
	}
}

func TestGetEnvironmentReading_AirTemperature(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/environment/air-temperature" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metadata": map[string]interface{}{
				"stations": []map[string]interface{}{
					{"id": "S107", "name": "East Coast Parkway", "location": map[string]float64{"latitude": 1.31, "longitude": 103.96}},
				},
				"reading_type": "DBT 1M F",
				"reading_unit": "deg C",
			},
			"items": []map[string]interface{}{
				{"timestamp": "2026-05-06T14:00:00+08:00", "readings": []map[string]interface{}{
					{"station_id": "S107", "value": 29.5},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetEnvironmentReading("air-temperature")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "S107") {
		t.Errorf("expected S107 in result, got: %s", result)
	}
}

func TestGetEnvironmentReading_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetEnvironmentReading("invalid-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestGetWeatherForecast_2Hour(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/environment/2-hour-weather-forecast" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"area_metadata": []map[string]interface{}{
				{"name": "Ang Mo Kio", "label_location": map[string]float64{"latitude": 1.375, "longitude": 103.839}},
			},
			"items": []map[string]interface{}{
				{"update_timestamp": "2026-05-06T14:00:00+08:00", "forecasts": []map[string]interface{}{
					{"area": "Ang Mo Kio", "forecast": "Partly Cloudy"},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetWeatherForecast("2-hour")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "Ang Mo Kio") {
		t.Errorf("expected Ang Mo Kio in result, got: %s", result)
	}
}

func TestGetWeatherForecast_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetWeatherForecast("bad-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestGetTransportInfo_Carpark(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transport/carpark-availability" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"timestamp": "2026-05-06T14:00:00+08:00", "carpark_data": []map[string]interface{}{
					{"carpark_number": "HE12", "carpark_info": []map[string]interface{}{
						{"lot_type": "C", "lots_available": "101", "total_lots": "105"},
					}},
				}},
			},
		})
	}))
	defer srv.Close()

	c := newTestRealtimeClient(srv.URL)
	result, err := c.GetTransportInfo("carpark")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "HE12") {
		t.Errorf("expected HE12 in result, got: %s", result)
	}
}

func TestGetTransportInfo_InvalidType(t *testing.T) {
	c := &RealtimeClient{http: &http.Client{}, baseURL: "http://unused"}
	_, err := c.GetTransportInfo("bad-type")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}
