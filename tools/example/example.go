// Package example is a placeholder tool group showing the shape every group
// follows: typed input and output structs, and an AddTools function that
// server.New calls. Replace it with the server's real tools.
package example

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// EchoInput is the input of echo. Field docs come from the jsonschema tag.
type EchoInput struct {
	Text string `json:"text" jsonschema:"the text to echo back"`
}

// EchoResult is the output of echo.
type EchoResult struct {
	Text string `json:"text"`
}

// AddTools registers the example tools.
func AddTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "echo",
		Description: "Return the given text unchanged. An example tool; replace it.",
		Annotations: &mcp.ToolAnnotations{Title: "Echo", ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in EchoInput) (*mcp.CallToolResult, EchoResult, error) {
		if in.Text == "" {
			// A returned error becomes a tool error the model can read and
			// correct, not a protocol failure.
			return nil, EchoResult{}, errors.New("text is required")
		}
		return nil, EchoResult(in), nil
	})
}
