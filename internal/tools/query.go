package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterQueryTool(s *server.MCPServer, c *client.DatasetClient) {
	s.AddTool(mcp.NewTool("query_dataset",
		mcp.WithDescription("Query actual records from a data.gov.sg dataset. Use get_dataset_metadata first to learn column names for filters. Supports column-level exact-match filters. Returns records, field types, and total count."),
		mcp.WithString("dataset_id", mcp.Required(), mcp.Description("Dataset ID, e.g. 'd_ca32584c91ee07d091a4ce75fa868414'")),
		mcp.WithObject("filters", mcp.Description("Optional column filters as key/value pairs. Key is column name, value is exact match string. Example: {\"residential_status\": \"overall\"}")),
		mcp.WithNumber("limit", mcp.Description("Max records to return, max 100 (default: 20)")),
		mcp.WithNumber("offset", mcp.Description("Number of records to skip for pagination (default: 0)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		datasetID, err := req.RequireString("dataset_id")
		if err != nil || datasetID == "" {
			return mcp.NewToolResultError("dataset_id is required"), nil
		}

		limit := req.GetInt("limit", 20)
		if limit < 1 {
			limit = 20
		}
		if limit > 100 {
			limit = 100
		}

		offset := req.GetInt("offset", 0)
		if offset < 0 {
			offset = 0
		}

		var filters map[string]string
		args := req.GetArguments()
		if f, ok := args["filters"].(map[string]any); ok {
			filters = make(map[string]string, len(f))
			for k, v := range f {
				if s, ok := v.(string); ok {
					filters[k] = s
				}
			}
		}

		result, err := c.QueryDataset(datasetID, filters, limit, offset)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("query failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}
