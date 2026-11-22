# Runbook

## Prerequisites

- Go (version in `go.mod`), Docker with Compose, Task (`go-task`)

## Local Setup

1. Copy `.env.example` to `.env` and fill in the values.
2. `task build`, then `task test`.

## Telemetry

Traces and metrics are off until `OTEL_EXPORTER_OTLP_ENDPOINT` is set (see
`.env.example`). To look at them locally, run a collector with a UI, such as Jaeger,
which takes OTLP over HTTP on port 4318:

```sh
docker run --rm -p 16686:16686 -p 4318:4318 jaegertracing/jaeger
OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=http://localhost:4318/v1/traces task run
```

Each MCP call is a span (`tools/call <tool>`) under its HTTP request, with
`mcp.response.size` (bytes) and `mcp.response.tokens_estimate` (bytes / 4). Metrics
go to a collector that takes them: `mcp.server.response.size` and
`mcp.server.operation.duration`, histograms by method and tool. Without a collector,
`<PREFIX>_LOG_LEVEL=debug` logs each call's size and duration.

## Troubleshooting

Add entries as they come up.
