# go-mcp-template

Starting point for Bailey's self-hosted Go MCP servers. It carries a working
Streamable HTTP server with one example tool, opt-in OAuth on the MCP endpoint, a
Docker image, a release workflow, and the shared development process. Each server
adds its own tools, CLAUDE.md notes and docs.

## Starting a new repo

1. Create the repo from this template (**Use this template** on GitHub) and clone it.
2. Run the init script with the module path, and the server name if it isn't the
   repo name:

   ```
   scripts/init.sh github.com/philosowaffle/foo-mcp
   ```

   It sets the module path, the name (binary, image, compose service) and the env
   prefix (`FOO_MCP_*`), writes this template's commit to `.template-version`, and
   deletes itself. It needs Go, Perl and, for `.template-version`, the `gh` CLI.
3. Replace `tools/example` with the server's tools, and fill in `CLAUDE.md`
   (Project Overview and Project-Specific Notes), `docs/architecture.md` and
   `docs/runbook.md`. Check `LICENSE` and the `LICENSE` value in `release.yml`.
4. While this template is private, add a `TEMPLATE_READ_TOKEN` secret: a fine-grained
   token with Contents: read on this repo.

`task run` starts the server on `127.0.0.1:8080`; `task --list` shows the rest.

## What you get

- `main.go`: health probe, protected-resource metadata, the MCP endpoint, and a guard
  that refuses to listen beyond loopback without auth.
- `internal/server`: the tool registry, plus the server's `Name` and `EnvPrefix`.
- `internal/auth`, `internal/health`, `internal/logging`: shared packages.
- `tools/example`: one `echo` tool showing the shape of a tool group.
- Tests: tools end to end over the SDK's in-memory transport (`TestListTools`), and
  `integration/` over real Streamable HTTP (`TestMCPTools`).
- `Taskfile.yml`, a multi-stage distroless `Dockerfile`, a sample
  `docker-compose.yml` and `.env.example`.
- Workflows: PR Checks, Release (multi-arch image to GHCR, binaries and changelog on a
  tag), Security and Template sync.

## What's shared

`.github/template-manifest` lists the files repos share with the template:

- **owned** files stay identical to the template: the process doc, ADR and spec
  templates, issue and PR templates, `.claude/settings.json`, security and Dependabot
  config, sponsorship, `cliff.toml`, the sync workflow, and `internal/auth`,
  `internal/health` and `internal/logging`.
- **seeded** files start from the template, then belong to the repo: `CLAUDE.md`,
  `main.go`, the Taskfile, Dockerfile, compose, `.env.example`, `.golangci.yml`,
  `.gitignore`, and the PR Checks and Release workflows.

The server's own code (`internal/server`, `tools/`, `integration/`) isn't shared.
`docs/process.md` holds the workflow itself.

## Changing the shared process

Change it here, in a PR to this repo. Each Monday, every repo's Template sync workflow
compares its `.template-version` with this repo's `main`. When a shared file has
changed, it files one `template-sync` issue listing the changes and the steps in
`.github/template-sync.md`, ready to be picked up and turned into a PR.
