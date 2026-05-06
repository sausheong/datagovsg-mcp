package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
)

func RegisterRealtimeTools(s *server.MCPServer, c *client.RealtimeClient) {
	s.AddTool(mcp.NewTool("get_environment_reading",
		mcp.WithDescription("Get current environmental sensor readings from NEA stations across Singapore. Each reading includes timestamp, station names, GPS coordinates, and measured values."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Type of reading. One of: air-temperature (°C), rainfall (mm), relative-humidity (%), wind-speed (km/h), wind-direction (degrees), uv-index, psi (Pollutant Standards Index), pm25 (PM2.5 by region)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		readingType, err := req.RequireString("type")
		if err != nil || readingType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetEnvironmentReading(readingType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("environment reading failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	s.AddTool(mcp.NewTool("get_weather_forecast",
		mcp.WithDescription("Get weather forecasts for Singapore. The 2-hour forecast gives area-level nowcasts; 24-hour gives regional conditions for today; 4-day gives an island-wide outlook with temperature ranges."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Forecast window. One of: 2-hour (area-level nowcast), 24-hour (regional forecast for today), 4-day (island-wide outlook)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		forecastType, err := req.RequireString("type")
		if err != nil || forecastType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetWeatherForecast(forecastType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("weather forecast failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	s.AddTool(mcp.NewTool("get_transport_info",
		mcp.WithDescription("Get live transport data from Singapore. Taxi returns GPS coordinates of available taxis. Carpark returns lot availability for 2,001 HDB carparks. Traffic-cameras returns image URLs and locations for traffic cameras."),
		mcp.WithString("type", mcp.Required(), mcp.Description(
			"Transport data type. One of: taxi (live taxi GPS locations), carpark (HDB carpark lot availability), traffic-cameras (live traffic camera image URLs and locations)",
		)),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		infoType, err := req.RequireString("type")
		if err != nil || infoType == "" {
			return mcp.NewToolResultError("type is required"), nil
		}
		result, err := c.GetTransportInfo(infoType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("transport info failed: %v", err)), nil
		}
		return mcp.NewToolResultText(result), nil
	})
}
