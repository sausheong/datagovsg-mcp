package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterDatasetTools(s *server.MCPServer, c *client.DatasetClient) {
	s.AddTool(mcp.NewTool("search_datasets",
		mcp.WithDescription("Search data.gov.sg datasets by keyword. Returns dataset IDs, names, descriptions, formats, and coverage dates. Use this first to find relevant datasets before querying their data."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Keyword or topic to search for, e.g. 'unemployment', 'rainfall', 'HDB'")),
		mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page, max 100 (default: 10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil || query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}
		page := req.GetInt("page", 1)
		if page < 1 {
			page = 1
		}
		limit := req.GetInt("limit", 10)
		if limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
		result, err := c.SearchDatasets(query, page, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("get_dataset_metadata",
		mcp.WithDescription("Get the full schema for a dataset: column names, data types, units, and categorical flags. Call this after search_datasets to understand a dataset's structure before querying it."),
		mcp.WithString("dataset_id", mcp.Required(), mcp.Description("Dataset ID from search_datasets, e.g. 'd_ca32584c91ee07d091a4ce75fa868414'")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("dataset_id")
		if err != nil || id == "" {
			return mcp.NewToolResultError("dataset_id is required"), nil
		}
		meta, err := c.GetDatasetMetadata(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("metadata fetch failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(meta, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("list_collections",
		mcp.WithDescription("Browse thematic collections of related datasets on data.gov.sg. Each collection groups datasets from the same agency or topic. Returns collection IDs and their child dataset IDs."),
		mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
		mcp.WithNumber("limit", mcp.Description("Results per page, max 100 (default: 10)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		page := req.GetInt("page", 1)
		if page < 1 {
			page = 1
		}
		limit := req.GetInt("limit", 10)
		if limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100
		}
		result, err := c.ListCollections(page, limit)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list_collections failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})

	s.AddTool(mcp.NewTool("get_collection_info",
		mcp.WithDescription("Get full details for a specific collection, including its child dataset IDs. Use after list_collections to see which datasets belong to a collection."),
		mcp.WithString("collection_id", mcp.Required(), mcp.Description("Collection ID from list_collections, e.g. '471'")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("collection_id")
		if err != nil || id == "" {
			return mcp.NewToolResultError("collection_id is required"), nil
		}
		info, err := c.GetCollectionInfo(id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("get_collection_info failed: %v", err)), nil
		}
		out, _ := json.MarshalIndent(info, "", "  ")
		return mcp.NewToolResultText(string(out)), nil
	})
}
