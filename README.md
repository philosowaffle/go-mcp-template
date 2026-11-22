# go-mcp-template

Starting point for Bailey's self-hosted Go MCP servers. It carries the shared
development process; each server adds its own code, CLAUDE.md notes and docs.

## Starting a new repo

1. Create the repo from this template (**Use this template** on GitHub).
2. Fill in `CLAUDE.md` (Project Overview and Project-Specific Notes) and the
   `docs/architecture.md` and `docs/runbook.md` skeletons.
   In `.golangci.yml`, set `local-prefixes` to the new module path, and add the
   server's binary to `.gitignore`.
3. Write this template's current commit SHA to `.template-version`, so the first
   Template sync run doesn't treat the repo as unsynced.
4. While this template is private, add a `TEMPLATE_READ_TOKEN` secret: a fine-grained
   token with Contents: read on this repo.

## What's shared

`.github/template-manifest` lists the shared files:

- **owned** files (process doc, ADR and spec templates, issue and PR templates,
  `.claude/settings.json`, security and Dependabot config, sponsorship, `cliff.toml`,
  the sync workflow) stay identical to the template.
- **seeded** files (`CLAUDE.md`, `pr-checks.yml`, `.golangci.yml`, `.gitignore`) start
  from the template, then belong to the repo.

`docs/process.md` holds the workflow itself: Conventional Commit PR titles, squash
merges, specs for non-trivial features and ADRs for significant decisions.

## Changing the shared process

Change it here, in a PR to this repo. Each Monday, every repo's Template sync workflow
compares its `.template-version` with this repo's `main`. When a shared file has
changed, it files one `template-sync` issue listing the changes and the steps in
`.github/template-sync.md`, ready to be picked up and turned into a PR.
