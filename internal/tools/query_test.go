package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestQueryDatasetTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result": map[string]any{
				"resource_id": "d_abc",
				"fields":      []map[string]any{{"id": "period", "type": "text"}},
				"records":     []map[string]any{{"period": "1992-03", "_id": 1}},
				"total":       1,
			},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "query_dataset", map[string]any{
		"dataset_id": "d_abc",
		"limit":      5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "1992-03") {
		t.Errorf("expected 1992-03 in output, got: %s", text)
	}
}

func TestQueryDatasetTool_WithFilters(t *testing.T) {
	var receivedFilters string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedFilters = r.URL.Query().Get("filters")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"result":  map[string]any{"fields": []any{}, "records": []any{}, "total": 0},
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools(srv.URL))

	_, err := callTool(s, "query_dataset", map[string]any{
		"dataset_id": "d_abc",
		"filters":    map[string]any{"residential_status": "overall"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if receivedFilters == "" {
		t.Error("expected filters to be passed to upstream API")
	}
}

func TestQueryDatasetTool_MissingDatasetID(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	RegisterQueryTool(s, newTestDatasetClientForTools("http://unused"))

	result, err := callTool(s, "query_dataset", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing dataset_id")
	}
}
