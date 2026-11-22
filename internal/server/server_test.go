package server

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newSession connects an MCP client to New() over the SDK's in-memory
// transport. Give New the fakes its tools need here (a fake API, a fake CLI)
// rather than mocking the tools.
func newSession(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	clientT, serverT := mcp.NewInMemoryTransports()
	ss, err := New().Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = ss.Close() })

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// call invokes a tool and decodes its structured output into out. It returns
// the tool's error text, or "" on success.
func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any, out any) string {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		if len(res.Content) > 0 {
			if tc, ok := res.Content[0].(*mcp.TextContent); ok {
				return tc.Text
			}
		}
		return "tool error"
	}
	if out != nil {
		b, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(b, out); err != nil {
			t.Fatalf("%s: decode %s: %v", name, b, err)
		}
	}
	return ""
}

// TestListTools pins the registered tool set. Add each new tool here.
func TestListTools(t *testing.T) {
	cs := newSession(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tool := range res.Tools {
		got = append(got, tool.Name)
		if tool.Description == "" {
			t.Errorf("%s has no description", tool.Name)
		}
	}
	sort.Strings(got)
	want := []string{"echo"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("tools = %v, want %v", got, want)
	}
}

func TestServerInfo(t *testing.T) {
	cs := newSession(t)
	info := cs.InitializeResult().ServerInfo
	if info.Name != Name || info.Version != Version {
		t.Errorf("serverInfo = %+v, want %s %s", info, Name, Version)
	}
}

func TestEcho(t *testing.T) {
	cs := newSession(t)
	tests := []struct {
		name    string
		args    map[string]any
		want    string
		wantErr string
	}{
		{"echoes text", map[string]any{"text": "hello"}, "hello", ""},
		{"empty text is a tool error", map[string]any{"text": ""}, "", "text is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out struct {
				Text string `json:"text"`
			}
			errText := call(t, cs, "echo", tt.args, &out)
			if tt.wantErr != "" {
				if !strings.Contains(errText, tt.wantErr) {
					t.Fatalf("error = %q, want %q", errText, tt.wantErr)
				}
				return
			}
			if errText != "" {
				t.Fatalf("error: %s", errText)
			}
			if out.Text != tt.want {
				t.Errorf("text = %q, want %q", out.Text, tt.want)
			}
		})
	}
}
