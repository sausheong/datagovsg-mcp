package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func newTestRealtimeClientForTools(serverURL string) *client.RealtimeClient {
	return client.NewRealtimeClientWithConfig(serverURL+"/v1", 5*time.Second)
}

func TestGetEnvironmentReadingTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"metadata": map[string]any{"reading_unit": "deg C", "stations": []any{}},
			"items":    []map[string]any{{"timestamp": "2026-05-06T14:00:00+08:00", "readings": []any{}}},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_environment_reading", map[string]any{"type": "air-temperature"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
}

func TestGetEnvironmentReadingTool_InvalidType(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools("http://unused"))

	result, err := callTool(s, "get_environment_reading", map[string]any{"type": "bad-type"})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for invalid type")
	}
}

func TestGetWeatherForecastTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"update_timestamp": "2026-05-06T14:00:00+08:00", "forecasts": []map[string]any{
					{"area": "Ang Mo Kio", "forecast": "Partly Cloudy"},
				}},
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_weather_forecast", map[string]any{"type": "2-hour"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Partly Cloudy") {
		t.Errorf("expected forecast in output, got: %s", text)
	}
}

func TestGetTransportInfoTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"timestamp": "2026-05-06T14:00:00+08:00", "carpark_data": []map[string]any{
					{"carpark_number": "HE12"},
				}},
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterRealtimeTools(s, newTestRealtimeClientForTools(srv.URL))

	result, err := callTool(s, "get_transport_info", map[string]any{"type": "carpark"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "HE12") {
		t.Errorf("expected HE12 in output, got: %s", text)
	}
}
