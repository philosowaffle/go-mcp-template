# Spec: Reusable Go MCP server scaffold

**Issue:** #2
**Status:** Complete

## Problem

A repo made from the template started with the dev process but no Go code, so each MCP
server copied its skeleton from another, and the copies drifted. `internal/auth`,
`internal/health` and `internal/logging` in kroger-mcp, speediance-mcp and vault-mcp
differ only in names.

## Acceptance Criteria

- [x] A repo made from the template passes `task ci` and builds its image after
      `scripts/init.sh` (checked by hand on a copy; CI runs both on this repo)
- [x] The example tool is reachable over Streamable HTTP and covered by `TestListTools`
      and `TestMCPTools` (`TestListTools`, `TestEcho`, `integration.TestMCPTools`)
- [x] Auth is off by default and turns on through `<PREFIX>_OIDC_*`, as in kroger-mcp
      and speediance-mcp (`TestRoutesWithAuth`, `TestRoutesWithoutAuth`,
      `TestConfigFromEnv`, `TestErrorsNameThePrefixedVariables`)
- [x] `release.yml` names nothing repo-specific: names come from the repository and
      `go list -m`
- [x] Every shared file is in `.github/template-manifest` as owned or seeded, and the
      README covers setup
- [x] `docs/process.md` and the README describe the scaffold

## Notes

- Built from speediance-mcp, the newest copy. Its CLI wrapper and tools stay there.
- The owned packages take the server's name and env prefix as arguments
  (`auth.ConfigFromEnv(name, prefix)`, `logging.Setup(prefix)`) instead of naming
  them, so one copy serves every repo and template sync can keep them identical.
  `internal/server` holds `Name` and `EnvPrefix`.
- `main.go`, the Taskfile, Dockerfile, compose, `.env.example` and `release.yml` are
  seeded: repos extend them (vault-mcp runs a multi-service `deploy/` stack, kroger-mcp
  publishes as `kroger-mcp`). `internal/server`, `tools/` and `integration/` aren't
  shared at all.
- `scripts/init.sh` renames the module, name and prefix in one go, since they appear
  in a dozen files, then deletes itself.
- The image's binary is `/server`, so the Dockerfile needs no name.
- The template's own release workflow publishes nothing.
