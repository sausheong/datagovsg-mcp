package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sausheong/datagovsg-mcp/internal/client"
	"github.com/sausheong/datagovsg-mcp/internal/tools"
)

func main() {
	addr := flag.String("http", "", "Start streamable HTTP server on this address (e.g. :8080). Omit for STDIO mode.")
	flag.Parse()

	apiKey := os.Getenv("DATAGOVSG_API_KEY")
	datasetClient := client.NewDatasetClient(apiKey)
	realtimeClient := client.NewRealtimeClient(apiKey)

	s := server.NewMCPServer(
		"datagovsg-mcp",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	tools.RegisterDatasetTools(s, datasetClient)
	tools.RegisterQueryTool(s, datasetClient)
	tools.RegisterRealtimeTools(s, realtimeClient)

	if *addr != "" {
		fmt.Fprintf(os.Stderr, "starting streamable HTTP server on %s/mcp\n", *addr)
		httpServer := server.NewStreamableHTTPServer(s)
		if err := httpServer.Start(*addr); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}
}
