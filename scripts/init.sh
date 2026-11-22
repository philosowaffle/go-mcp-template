#!/usr/bin/env bash
# Turns a fresh copy of go-mcp-template into a named server. Run it once from
# the repo root, right after creating the repo from the template:
#
#   scripts/init.sh github.com/philosowaffle/foo-mcp          # name: foo-mcp
#   scripts/init.sh github.com/philosowaffle/kroger-cart kroger-mcp
#
# It sets the module path, the server name (binary, image, compose service)
# and the env prefix (FOO_MCP_*), records the template commit in
# .template-version, removes template-only files, then deletes itself.
set -euo pipefail

template_repo=philosowaffle/go-mcp-template
old_module=github.com/$template_repo

module="${1:?usage: scripts/init.sh <module path> [name]}"
name="${2:-${module##*/}}"
if ! [[ "$name" =~ ^[a-z][a-z0-9-]*$ ]]; then
  echo "name must be lowercase letters, digits and dashes: $name" >&2
  exit 1
fi
prefix="$(printf '%s' "$name" | tr 'a-z-' 'A-Z_')"

cd "$(git rev-parse --show-toplevel)"
if [ "$(go list -m)" != "$old_module" ]; then
  echo "already initialised as $(go list -m)" >&2
  exit 1
fi

# Only the files that carry the names. The workflows and docs name the
# template itself and stay as they are.
# perl -pi works the same on Linux, macOS and Git Bash, unlike sed -i.
git ls-files -z '*.go' go.mod Taskfile.yml Dockerfile docker-compose.yml \
  .env.example .golangci.yml .gitignore |
  MODULE="$module" NAME="$name" PREFIX="$prefix" OLD="$old_module" xargs -0 perl -pi -e '
    s/\Q$ENV{OLD}\E/$ENV{MODULE}/g;
    s/GO_MCP_TEMPLATE/$ENV{PREFIX}/g;
    s/go-mcp-template/$ENV{NAME}/g;
  '

# Template-only files: its own specs and this README.
find docs/specs -name '[0-9]*.md' ! -name '000-template.md' -delete
printf '# %s\n\nAn MCP server. See docs/ and CLAUDE.md.\n' "$name" > README.md

# The commit the repo starts from, so the first Template sync run only
# reports later changes.
if sha="$(gh api "repos/$template_repo/commits/main" --jq .sha 2>/dev/null)" && [ -n "$sha" ]; then
  echo "$sha" > .template-version
else
  echo "warning: couldn't read $template_repo's main; write its commit SHA to .template-version by hand" >&2
fi

gofmt -w .
go mod tidy
go build ./...

rm -- "$0"
rmdir scripts 2>/dev/null || true

cat <<MSG
Initialised $name ($module), env prefix ${prefix}_.
Next: fill in CLAUDE.md, docs/architecture.md and docs/runbook.md, replace
tools/example, check LICENSE, then run task ci and commit.
MSG
