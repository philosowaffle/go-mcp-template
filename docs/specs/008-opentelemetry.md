# Spec: OpenTelemetry traces and response-size metrics

**Issue:** #8
**Status:** Complete

## Problem

Several MCP servers come from this template, and none say how big their responses are.
That hides where caching would pay off, and which tools should return a short summary
before fetching full detail. There are no traces either, so no flame graph of where a
tool call spends its time.

## Acceptance Criteria

- [x] With no OTLP endpoint set, nothing is exported and no provider is installed
  (`TestSetup_OffByDefault`, `TestEnabled`)
- [x] An OTLP endpoint installs tracer and meter providers
  (`TestSetup_EnabledWithEndpoint`)
- [x] A `grpc` protocol setting fails at startup with a clear error
  (`TestSetup_RejectsGRPC`)
- [x] Each MCP call gets a span with method, tool name, request and response sizes and a
  token estimate (`TestInstrument_ToolSpanHasSizes`)
- [x] A tool error marks its span as an error (`TestInstrument_ToolErrorMarksSpan`)
- [x] Response size and duration are histograms by method and tool
  (`TestInstrument_RecordsHistograms`)
- [x] Over Streamable HTTP, a tool span is a child of its HTTP request span
  (`TestHTTPHandler_ToolSpanIsChildOfRequest`)
- [x] At debug level, each call logs its size and duration
  (`TestInstrument_LogsSizeAtDebug`)
- [x] `internal/telemetry` is owned in the manifest; the runbook, `.env.example` and the
  sync steps cover it

## Notes

- `internal/telemetry` is owned, like `internal/auth`. `main.go` (seeded) wires it in
  three places: `Setup` and its shutdown in `run`, `Instrument` on the MCP server, and
  `HTTPHandler` around the MCP endpoint. `internal/server` and `tools/` don't change.
- Configuration is the standard `OTEL_*` variables rather than `<PREFIX>_*`, so any
  collector's docs apply as written. That is the one exception to "every variable
  starts with `EnvPrefix`" in `docs/process.md`.
- Export is OTLP over HTTP/protobuf only. gRPC would pull the gRPC client into every
  server for no gain on a home network; asking for it fails at startup rather than
  exporting nowhere.
- Span and attribute names follow the OpenTelemetry MCP conventions where they exist
  (`tools/call <tool>`, `mcp.method.name`, `gen_ai.tool.name`, `error.type`, the
  `mcp.server.operation.duration` metric). Sizes are this template's own:
  `mcp.request.size`, `mcp.response.size`, `mcp.response.tokens_estimate` on spans, and
  the `mcp.server.response.size` histogram.
- Response size is the result marshalled to JSON once more. That's skipped when no span
  is recording, no meter provider is installed and debug logging is off.
- The token estimate is bytes / 4, rounded up: good for ranking tools by cost, not a
  tokenizer count.
- `/healthz` isn't traced: only the MCP endpoint is wrapped.
