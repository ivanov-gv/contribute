# Devcontainer Configuration Design

**Date:** 2026-03-15
**Status:** Approved

## Overview

Replace the existing single VS Code devcontainer with three purpose-built configurations:

1. **`claude-code`** — JetBrains GoLand + Claude Code CLI, suitable for `--dangerously-skip-permissions` mode
2. **`default`** — JetBrains GoLand, minimal Go development environment
3. **`claude-session`** — Standalone Docker image for isolated Claude Code terminal sessions

---

## File Structure

```
.devcontainer/
  claude-code/
    devcontainer.json     ← JetBrains + Claude Code
    Dockerfile
    install-deps.sh       ← system deps, no Node.js
  default/
    devcontainer.json     ← JetBrains minimal (rewritten)
  shared/
    setup.sh              ← moved from root
    init-firewall.sh      ← moved from root

deploy/
  claude-session/
    Dockerfile            ← same image as claude-code
    docker-compose.yml    ← usage entry point
```

The root `.devcontainer/devcontainer.json` and `.devcontainer/Dockerfile` are removed and replaced by the named configs above.

---

## Variant 1: `claude-code`

### Purpose

Full development environment for JetBrains GoLand with Claude Code CLI. Designed to be safe for `--dangerously-skip-permissions` mode via an outbound firewall allowlist.

### Dockerfile

- **Base image:** `golang:latest` (Debian Bookworm, always latest Go)
- **No Node.js, no npm** — remove the Node.js install from `install-deps.sh`, remove the npm global dir setup (`/usr/local/share/npm-global`, `chown`, `NPM_CONFIG_PREFIX`, `PATH` extension for npm), and replace `npm install -g @anthropic-ai/claude-code` with the native binary install below
- **Claude Code native binary installation:**
  1. Detect arch: `ARCH=$(dpkg --print-architecture)` → maps to `amd64` or `arm64`
  2. Map to Claude Code release arch: `amd64` → `x64`, `arm64` → `arm64`
  3. Fetch latest release tag from GitHub API: `curl -s https://api.github.com/repos/anthropics/claude-code/releases/latest | jq -r '.tag_name'`
  4. Download binary: `https://github.com/anthropics/claude-code/releases/download/<tag>/claude-linux-<arch>`
  5. Install to `/usr/local/bin/claude`, `chmod +x`
  6. Verify: `claude --version`
  - Note: exact release asset naming must be confirmed against the GitHub releases page at implementation time
- **gopls** installed via `go install golang.org/x/tools/gopls@latest` (run as `dev` user so it lands in `$GOPATH/bin`)
- **Dev user:** non-root `dev` user with zsh as default shell
- **Shell:** zsh + Powerline10k theme (via zsh-in-docker), fzf key bindings
- **Tools:** git, git-delta, Docker CLI (no daemon), docker-compose-plugin, iptables, ipset, iproute2, dnsutils, aggregate, jq, gh, nano, vim, sudo
- **Firewall scripts** copied to `/usr/local/bin/` from `.devcontainer/shared/` — updated `COPY` paths:
  - `COPY .devcontainer/claude-code/install-deps.sh /tmp/install-deps.sh`
  - `COPY .devcontainer/shared/init-firewall.sh /usr/local/bin/`
  - `COPY .devcontainer/shared/setup.sh /usr/local/bin/`
- **Passwordless sudo** for `dev` on firewall scripts only
- **History persistence:** `/commandhistory` directory owned by `dev`
- **`NODE_OPTIONS` removed** — intentionally dropped along with Node.js

### install-deps.sh

Same as current but with the Node.js section (`nodesource` setup, `nodejs` apt install) removed entirely.

### devcontainer.json

| Field | Value |
|---|---|
| `name` | `gh-contribute (claude-code)` |
| `build.dockerfile` | `Dockerfile` |
| `build.context` | `../..` |
| `build.args` | `TZ`, `CLAUDE_CODE_VERSION`, `GIT_DELTA_VERSION`, `ZSH_IN_DOCKER_VERSION` |
| `runArgs` | `--cap-add=NET_ADMIN`, `--cap-add=NET_RAW`, Docker socket bind mount |
| `mounts` | history volume + `~/.claude` bind mount (see below) |
| `remoteUser` | `dev` |
| `workspaceMount` | bind, delegated consistency |
| `workspaceFolder` | `/workspace` |
| `postStartCommand` | `sudo /usr/local/bin/setup.sh` |
| `waitFor` | `postStartCommand` |

**JetBrains customization:**
```json
"customizations": {
  "jetbrains": {
    "backend": "GoLand"
  }
}
```

**Mounts:**
- History volume: `gh-contribute-bashhistory-${devcontainerId}` → `/commandhistory`
- `~/.claude` bind mount: `${localEnv:HOME}/.claude` → `/home/dev/.claude` — imports local Claude Code plugins, settings, and auth into the container

**containerEnv:**
- `CLAUDE_CONFIG_DIR=/home/dev/.claude`
- `ANTHROPIC_API_KEY=${localEnv:ANTHROPIC_API_KEY}`
- `GH_CONTRIBUTE_TOKEN=${localEnv:GH_CONTRIBUTE_TOKEN}`
- `GITHUB_TOKEN=${localEnv:GITHUB_TOKEN}`
- `GH_TOKEN=${localEnv:GH_TOKEN}`
- `DOCKER_NETWORK=gh-contribute-${devcontainerId}`
- `POWERLEVEL9K_DISABLE_GITSTATUS=true`

---

## Variant 2: `default`

### Purpose

Minimal Go development environment for JetBrains GoLand. No Claude Code, no firewall, no custom Dockerfile.

### devcontainer.json

| Field | Value |
|---|---|
| `name` | `gh-contribute (default)` |
| `image` | `golang:latest` |
| `remoteUser` | `root` |
| `postCreateCommand` | `go install golang.org/x/tools/gopls@latest` |

**JetBrains customization:**
```json
"customizations": {
  "jetbrains": {
    "backend": "GoLand"
  }
}
```

**containerEnv:**
- `GOPATH=/root/go`
- `PATH` extended with `/root/go/bin`

No mounts, no runArgs, no firewall, no extra tools. Git is already present in the base image.

`remoteUser: root` is intentional — avoids user management overhead in a minimal setup. `GOPATH=/root/go` and the corresponding `PATH` extension follow from this.

---

## Variant 3: `claude-session`

### Purpose

Standalone Docker image for isolated Claude Code terminal sessions. Not a devcontainer — started manually via Docker Compose, used interactively via `docker compose run`.

### Dockerfile

Identical content to `claude-code/Dockerfile`. Same base, same tools, same Claude Code native binary, same `dev` user and zsh setup. Firewall scripts present and sudoable, but not auto-run.

### docker-compose.yml

```yaml
services:
  claude-session:
    build:
      context: ../..
      dockerfile: deploy/claude-session/Dockerfile
    stdin_open: true
    tty: true
    user: dev
    command: /bin/zsh
    cap_add:
      - NET_ADMIN
      - NET_RAW
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
```

- `build.context: ../..` sets the build context to the workspace root — this means the `claude-session/Dockerfile` is **identical** to `claude-code/Dockerfile` (same `COPY .devcontainer/shared/...` paths work unchanged). No separate copy with adjusted paths needed.
- `user: dev` ensures the session starts as the non-root `dev` user explicitly.
- `command: /bin/zsh` overrides the `golang:latest` default (`bash`) to drop into zsh.

**Usage:**
```bash
docker compose run --rm claude-session
```

Drops into zsh as `dev`. Clone a repo, run `sudo init-firewall.sh` for outbound protection, then run `claude`.

No workspace bind mount (fresh clone inside container), no `~/.claude` bind mount (paths differ), no API key passthrough (authenticate inside container via `claude auth login`).

---

## Shared Scripts

`setup.sh` and `init-firewall.sh` move from `.devcontainer/` root to `.devcontainer/shared/`. The `claude-code` Dockerfile copies them from there. The `claude-session` Dockerfile does the same. The `default` variant does not use them.

`init-firewall.sh` allowlist stays unchanged:
- GitHub IP ranges
- `api.anthropic.com`
- Go module proxy / checksum DB
- npm registry (removed — Node.js and npm are not installed; no longer applicable)
- Docker Hub, GHCR
- Sentry, Statsig, VS Code Marketplace (Claude Code internals)
- DNS, SSH, localhost, host Docker network

---

## Migration

| Old path | New path |
|---|---|
| `.devcontainer/devcontainer.json` | `.devcontainer/claude-code/devcontainer.json` |
| `.devcontainer/Dockerfile` | `.devcontainer/claude-code/Dockerfile` |
| `.devcontainer/install-deps.sh` | `.devcontainer/claude-code/install-deps.sh` |
| `.devcontainer/setup.sh` | `.devcontainer/shared/setup.sh` |
| `.devcontainer/init-firewall.sh` | `.devcontainer/shared/init-firewall.sh` |
| `.devcontainer/default/devcontainer.json` | `.devcontainer/default/devcontainer.json` (rewritten) |
| _(new)_ | `deploy/claude-session/Dockerfile` |
| _(new)_ | `deploy/claude-session/docker-compose.yml` |
