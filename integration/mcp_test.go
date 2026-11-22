//go:build integration

// Package integration runs the server in-process and calls every tool over a
// real Streamable HTTP client. Point it at the real dependencies the tools
// wrap (a sandbox API, a pinned CLI, a compose service), never at fakes.
//
// Run with: task test:integration
package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/philosowaffle/go-mcp-template/internal/server"
)

func ctx(t *testing.T) context.Context {
	t.Helper()
	c, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	return c
}

// mcpSession serves server.New over Streamable HTTP, as main does, and
// connects a client to it.
func mcpSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	srv := server.New()
	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return srv }, &mcp.StreamableHTTPOptions{Stateless: true})
	ts := httptest.NewServer(h)
	t.Cleanup(ts.Close)
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "integration", Version: "0"}, nil).
		Connect(ctx(t), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// TestMCPTools calls every tool once. It fails for any tool without a case,
// so a new tool can't skip integration coverage.
func TestMCPTools(t *testing.T) {
	cs := mcpSession(t)

	res, err := cs.ListTools(ctx(t), nil)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]map[string]any{
		"echo": {"text": "integration"},
	}
	for _, tool := range res.Tools {
		if _, ok := cases[tool.Name]; !ok {
			t.Errorf("no TestMCPTools case for %s", tool.Name)
		}
	}

	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			out, err := cs.CallTool(ctx(t), &mcp.CallToolParams{Name: name, Arguments: args})
			if err != nil {
				t.Fatalf("CallTool: %v", err)
			}
			if out.IsError {
				t.Fatalf("tool error: %+v", out.Content)
			}
		})
	}
}
