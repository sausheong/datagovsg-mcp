package main

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
	"github.com/sausheong/datagovsg-mcp/internal/tools"
)

func main() {
	datasetClient := client.NewDatasetClient()
	realtimeClient := client.NewRealtimeClient()

	s := server.NewMCPServer(
		"datagovsg-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tools.RegisterDatasetTools(s, datasetClient)
	tools.RegisterQueryTool(s, datasetClient)
	tools.RegisterRealtimeTools(s, realtimeClient)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
