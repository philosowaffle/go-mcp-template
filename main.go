// Command go-mcp-template is an MCP server that serves Streamable HTTP for
// self-hosting. Describe what it wraps and what its tools are for here.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/philosowaffle/go-mcp-template/internal/auth"
	"github.com/philosowaffle/go-mcp-template/internal/health"
	"github.com/philosowaffle/go-mcp-template/internal/logging"
	"github.com/philosowaffle/go-mcp-template/internal/server"
	"github.com/philosowaffle/go-mcp-template/internal/telemetry"
)

// listenAddr defaults to loopback so a bare local run without auth is
// private. The container image sets <PREFIX>_ADDR=:8080.
func listenAddr() string {
	if addr := os.Getenv(server.EnvPrefix + "_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:8080"
}

// allowUnauthEnv opts in to serving the MCP endpoint without auth on a
// non-loopback address.
const allowUnauthEnv = server.EnvPrefix + "_ALLOW_UNAUTHENTICATED"

// authConfig reads the <PREFIX>_OIDC_* settings.
func authConfig() auth.Config {
	return auth.ConfigFromEnv(server.Name, server.EnvPrefix)
}

// isLoopback reports whether addr (host:port) listens only on loopback. An
// empty host binds every interface, so it doesn't count.
func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil || host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// checkExposure refuses to serve an unauthenticated MCP endpoint beyond
// loopback unless the operator opted in, since anyone who can reach it can
// use every tool.
func checkExposure(addr string, authOn bool) error {
	if authOn || isLoopback(addr) {
		return nil
	}
	if allow, _ := strconv.ParseBool(os.Getenv(allowUnauthEnv)); allow {
		slog.Warn("auth disabled on a non-loopback address; anyone who can reach it can use every tool",
			"addr", addr, allowUnauthEnv, true)
		return nil
	}
	return fmt.Errorf("refusing to listen on %q without auth: set %s_OIDC_ISSUER to require OAuth, "+
		"listen on loopback (e.g. 127.0.0.1:8080), or set %s=true if the network is trusted", addr, server.EnvPrefix, allowUnauthEnv)
}

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe the running server's health endpoint and exit (used by the container HEALTHCHECK)")
	version := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *version {
		fmt.Println(server.Name, server.Version, server.Commit)
		return
	}

	logging.Setup(server.EnvPrefix)
	addr := listenAddr()

	if *healthcheck {
		if err := health.Check(addr); err != nil {
			slog.Error("healthcheck failed", "err", err)
			os.Exit(1)
		}
		return
	}

	if err := run(addr); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(addr string) error {
	if err := checkExposure(addr, authConfig().Enabled()); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Off unless an OTLP endpoint is set (OTEL_EXPORTER_OTLP_ENDPOINT).
	shutdownTelemetry, err := telemetry.Setup(ctx, server.Name, server.Version)
	if err != nil {
		return err
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTelemetry(sctx); err != nil {
			slog.Warn("telemetry shutdown", "err", err)
		}
	}()

	// Build the clients the tools need here (an API client, a CLI wrapper)
	// and pass them to newHandler and server.New.

	handler, err := newHandler(ctx)
	if err != nil {
		return err
	}

	// No ReadTimeout or WriteTimeout: MCP responses can be long-lived SSE
	// streams, and either deadline would cut them off. The SDK caps request
	// bodies (4 MiB) itself.
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           requestLog(handler),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "service", server.Name, "version", server.Version, "commit", server.Commit, "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	}

	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(sctx)
}

// newHandler assembles the HTTP routes:
//
//	/healthz                      liveness probe (no auth)
//	/.well-known/...              protected-resource metadata when auth is on,
//	                              404 otherwise (no auth, per RFC 9728)
//	/                             MCP Streamable HTTP (bearer auth when on;
//	                              cross-origin browser requests rejected;
//	                              traced when telemetry is on)
func newHandler(ctx context.Context) (http.Handler, error) {
	srv := server.New()
	telemetry.Instrument(srv)
	var mcpHandler http.Handler = mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{
		// Stateless: the tools keep no per-session state and never call back
		// to the client. go-sdk v1.8.0 rejects, on stateful servers, any call
		// whose _meta carries io.modelcontextprotocol/protocolVersion (SEP-2575),
		// even at 2025-11-25, which Claude's client now sends on every call.
		Stateless: true,
	})

	mux := http.NewServeMux()
	mux.HandleFunc(health.Path, health.Handler)

	// Discovery paths must be reachable without auth. Anything under
	// /.well-known/ other than the protected-resource document gets a plain
	// 404, so a client probing this host for authorization-server metadata
	// moves on to the advertised issuer instead of hitting a 401 from the MCP
	// handler. The exact metadata path registered below takes precedence.
	mux.HandleFunc("/.well-known/", http.NotFound)

	if cfg := authConfig(); cfg.Enabled() {
		authn, err := auth.New(ctx, cfg)
		if err != nil {
			return nil, err
		}
		mux.Handle(auth.MetadataPath, authn.MetadataHandler())
		mcpHandler = authn.Middleware()(mcpHandler)

		md := authn.Metadata()
		slog.Info("auth enabled",
			"resource", md.Resource,
			"authorization_servers", md.AuthorizationServers,
			"scopes_supported", md.ScopesSupported,
		)
	} else {
		slog.Warn("auth disabled: the MCP endpoint is open to anyone who can reach it; set " + server.EnvPrefix + "_OIDC_ISSUER to require OAuth")
	}

	// The SDK's DNS-rebinding check only covers loopback. This rejects
	// cross-origin browser requests on any address, so a web page can't drive
	// the tools from the user's browser. Clients that send neither Origin nor
	// Sec-Fetch-Site (Claude Code, claude.ai's servers) are unaffected.
	mux.Handle("/", telemetry.HTTPHandler(http.NewCrossOriginProtection().Handler(mcpHandler), server.Name))
	return mux, nil
}

// requestLog logs each request at TRACE, excluding health probes. Only the
// method, path, status, size and duration are logged; query strings are left
// out.
// httpsnoop keeps http.Flusher working for the MCP SSE stream.
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == health.Path {
			next.ServeHTTP(w, r)
			return
		}
		m := httpsnoop.CaptureMetrics(next, w, r)
		logging.Trace("request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"status", m.Code,
			"bytes", m.Written,
			"dur", m.Duration,
		)
	})
}
