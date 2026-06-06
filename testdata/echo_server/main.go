// testdata/echo_server is a simple in-memory MCP echo server used in
// integration tests. It implements one tool ("echo") that echoes back
// whatever message it receives.
package main

import (
	"context"
	"fmt"
	"log"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "echo-server",
		Version: "0.1.0",
	}, nil)

	type EchoInput struct {
		Message string `json:"message" jsonschema:"the message to echo back"`
	}
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "echo",
		Description: "Echo a message back to the caller",
	}, func(_ context.Context, _ *mcpsdk.CallToolRequest, input EchoInput) (
		*mcpsdk.CallToolResult, struct {
			Echo string `json:"echo"`
		}, error,
	) {
		return nil, struct {
			Echo string `json:"echo"`
		}{Echo: input.Message}, nil
	})

	if err := server.Run(context.Background(), &mcpsdk.StdioTransport{}); err != nil {
		log.Fatal(fmt.Errorf("server: %w", err))
	}
}
