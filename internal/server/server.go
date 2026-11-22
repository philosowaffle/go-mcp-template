// Package server builds the MCP server and is the one place tools are
// registered. To add a tool, create a package under tools/ and call its
// registration function from New.
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/philosowaffle/go-mcp-template/tools/example"
)

// Name identifies this server to MCP clients, names the binary and image,
// and is advertised as the protected-resource name when auth is on.
const Name = "go-mcp-template"

// EnvPrefix starts every variable the server reads: <EnvPrefix>_ADDR,
// <EnvPrefix>_LOG_LEVEL, <EnvPrefix>_OIDC_*.
const EnvPrefix = "GO_MCP_TEMPLATE"

// Version and Commit are set at build time with
//
//	-ldflags "-X <module>/internal/server.Version=v1.2.3
//	          -X <module>/internal/server.Commit=abc1234"
//
// by the Taskfile, Dockerfile and release workflow. Version is reported to MCP
// clients as serverInfo.version.
var (
	Version = "dev"
	Commit  = "unknown"
)

// instructions tells the client what the tools are for and how to use them
// together. Replace it with this server's own.
const instructions = `An example server. echo returns the text it is given.`

// New builds the MCP server with every tool registered.
func New() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: Name, Version: Version}, &mcp.ServerOptions{
		Instructions: instructions,
	})

	example.AddTools(s)

	return s
}
