package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func newTestDatasetClientForTools(serverURL string) *client.DatasetClient {
	return client.NewDatasetClientWithConfig(serverURL+"/v2/public/api", serverURL+"/api/action/datastore_search", "", 5*time.Second)
}

func callTool(s *server.MCPServer, name string, args map[string]any) (*mcp.CallToolResult, error) {
	tool := s.GetTool(name)
	if tool == nil {
		return nil, fmt.Errorf("tool %q not found", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	return tool.Handler(context.Background(), req)
}

func TestSearchDatasetsTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"datasets": []map[string]any{
					{"datasetId": "d_abc", "name": "Test Dataset", "format": "CSV"},
				},
				"pages": 1, "totalRowCount": 1,
			},
			"errorMsg": "",
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterDatasetTools(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "search_datasets", map[string]any{"query": "test"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "d_abc") {
		t.Errorf("expected d_abc in output, got: %s", text)
	}
}

func TestSearchDatasetsTool_MissingQuery(t *testing.T) {
	s := server.NewMCPServer("test", "1.0")
	RegisterDatasetTools(s, newTestDatasetClientForTools("http://unused"))

	result, err := callTool(s, "search_datasets", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing query")
	}
}

func TestGetDatasetMetadataTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{
				"datasetId": "d_abc", "name": "Test Dataset",
				"columnMetadata": map[string]any{
					"order": []string{"c_1"},
					"map":   map[string]string{"c_1": "period"},
					"metaMapping": map[string]any{
						"c_1": map[string]any{"name": "period", "columnTitle": "Period", "dataType": "Text", "isCategorical": false},
					},
				},
			},
			"errorMsg": "",
		})
	}))
	defer srv.Close()

	s := server.NewMCPServer("test", "1.0")
	RegisterDatasetTools(s, newTestDatasetClientForTools(srv.URL))

	result, err := callTool(s, "get_dataset_metadata", map[string]any{"dataset_id": "d_abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "period") {
		t.Errorf("expected period column in output, got: %s", text)
	}
}
