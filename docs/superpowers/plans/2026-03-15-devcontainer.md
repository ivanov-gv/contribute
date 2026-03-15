# Devcontainer Configuration Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the existing single VS Code devcontainer with three named configurations: `claude-code` (JetBrains + Claude Code native binary), `default` (JetBrains minimal), and `claude-session` (standalone Docker image).

**Architecture:** All three variants live under `.devcontainer/<name>/` or `deploy/claude-session/`. Shared scripts (`setup.sh`, `init-firewall.sh`) move to `.devcontainer/shared/` and are referenced by both Dockerfiles. `claude-code` and `claude-session` use an identical Dockerfile; the latter is a plain copy.

**Tech Stack:** Docker, devcontainer spec, Go, gopls, zsh, Powerline10k, git-delta, Claude Code native binary (GitHub releases)

---

## Chunk 1: Shared Scripts Migration

**Files:**
- Create: `.devcontainer/shared/setup.sh` (moved from root)
- Create: `.devcontainer/shared/init-firewall.sh` (moved from root, minus npm entry)
- Delete: `.devcontainer/setup.sh`
- Delete: `.devcontainer/init-firewall.sh`

---

### Task 1: Move `setup.sh` to `shared/`

**Files:**
- Create: `.devcontainer/shared/setup.sh`

- [ ] **Step 1: Create the shared directory and move setup.sh**

```bash
mkdir -p .devcontainer/shared
git mv .devcontainer/setup.sh .devcontainer/shared/setup.sh
```

- [ ] **Step 2: Verify the move**

```bash
cat .devcontainer/shared/setup.sh
```

Expected: full contents of the original `setup.sh` (Docker network creation, socket permissions, firewall call).

---

### Task 2: Move `init-firewall.sh` to `shared/` and remove npm allowlist entry

**Files:**
- Create: `.devcontainer/shared/init-firewall.sh`
- Delete: `.devcontainer/init-firewall.sh`

- [ ] **Step 1: Move init-firewall.sh**

```bash
git mv .devcontainer/init-firewall.sh .devcontainer/shared/init-firewall.sh
```

- [ ] **Step 2: Remove the npm registry domain entry**

Open `.devcontainer/shared/init-firewall.sh`. Find and delete this line (inside the `for domain in \` block, roughly line 76):
```
    "registry.npmjs.org" \
```

- [ ] **Step 3: Remove the stale npm comment from the header**

In the same file, find the header comment block near the top (around line 10) and delete this line:
```
#   - npm registry
```

- [ ] **Step 4: Verify npm is fully gone and surrounding lines are intact**

```bash
grep -n "npmjs\|npm registry" .devcontainer/shared/init-firewall.sh
```

Expected: no matches.

```bash
grep -n "anthropic\|proxy.golang" .devcontainer/shared/init-firewall.sh
```

Expected: both `api.anthropic.com` and `proxy.golang.org` are present with trailing backslashes, confirming the domain list is syntactically intact after removing the npm entry.

- [ ] **Step 5: Commit**

```bash
git add .devcontainer/shared/
git commit -m "refactor: move shared devcontainer scripts to shared/, remove npm from firewall allowlist"
```

---

## Chunk 2: claude-code Variant

**Files:**
- Create: `.devcontainer/claude-code/install-deps.sh` (moved from root, Node.js section removed)
- Create: `.devcontainer/claude-code/Dockerfile` (moved from root, rewritten)
- Create: `.devcontainer/claude-code/devcontainer.json` (moved from root, rewritten for JetBrains)
- Delete: `.devcontainer/install-deps.sh`
- Delete: `.devcontainer/Dockerfile`
- Delete: `.devcontainer/devcontainer.json`

---

### Task 3: Move and update `install-deps.sh`

**Files:**
- Create: `.devcontainer/claude-code/install-deps.sh`

- [ ] **Step 1: Create the claude-code directory and move install-deps.sh**

```bash
mkdir -p .devcontainer/claude-code
git mv .devcontainer/install-deps.sh .devcontainer/claude-code/install-deps.sh
```

- [ ] **Step 2: Remove the Node.js section**

Open `.devcontainer/claude-code/install-deps.sh`. Delete the entire Node.js block (comment + code):

```bash
# ── Node.js ────────────────────────────────────────────────────────────────────
# Required for Claude Code. Claude Code itself is installed later as the dev user
# via `npm install -g`, using a user-owned npm global dir.
NODE_VERSION="${NODE_VERSION:-20}"
curl -fsSL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | bash -
apt-get install -y --no-install-recommends nodejs
apt-get clean && rm -rf /var/lib/apt/lists/*
```

- [ ] **Step 3: Verify no Node.js references remain**

```bash
grep -in "node\|npm" .devcontainer/claude-code/install-deps.sh
```

Expected: no matches.

- [ ] **Step 4: Commit**

```bash
git add .devcontainer/claude-code/install-deps.sh
git commit -m "refactor: move install-deps.sh to claude-code/, remove Node.js install"
```

---

### Task 4: Rewrite the `claude-code` Dockerfile

**Files:**
- Create: `.devcontainer/claude-code/Dockerfile`

- [ ] **Step 1: Move the Dockerfile**

```bash
git mv .devcontainer/Dockerfile .devcontainer/claude-code/Dockerfile
```

- [ ] **Step 2: Rewrite the file**

Replace the entire contents of `.devcontainer/claude-code/Dockerfile` with:

```dockerfile
FROM golang:latest

ARG TZ
ENV TZ="$TZ"

ARG CLAUDE_CODE_VERSION=latest
ARG GIT_DELTA_VERSION=0.18.2
ARG ZSH_IN_DOCKER_VERSION=1.2.0

# Install all system-level tools via the dedicated script.
# To add a new tool, edit install-deps.sh — no Dockerfile changes needed.
COPY .devcontainer/claude-code/install-deps.sh /tmp/install-deps.sh
RUN chmod +x /tmp/install-deps.sh && \
  GIT_DELTA_VERSION=${GIT_DELTA_VERSION} \
  /tmp/install-deps.sh && \
  rm /tmp/install-deps.sh

# Create non-root dev user.
# adduser (Debian wrapper) picks a free UID automatically — avoids the
# "no matching entries in passwd file" error that occurs when a hard-coded
# UID like 1000 is already taken by a package installed above.
RUN adduser --disabled-password --gecos "" --shell /usr/bin/zsh dev

# Allow dev user to use docker group (GID reconciled at runtime in setup.sh)
RUN groupadd -f docker && usermod -aG docker dev

# Persist bash history across rebuilds
RUN mkdir /commandhistory && touch /commandhistory/.bash_history && chown -R dev /commandhistory

# Signal that this is a devcontainer
ENV DEVCONTAINER=true

# GOPATH in the dev user home; Go binary already on PATH via the base image
ENV GOPATH=/home/dev/go
ENV PATH=$PATH:/home/dev/go/bin

# Create workspace and per-user config directories
RUN mkdir -p /workspace /home/dev/.claude /home/dev/go && \
  chown -R dev:dev /workspace /home/dev/.claude /home/dev/go

WORKDIR /workspace

# Build the gh-contribute binary from project source and install system-wide
COPY . /build/gh-contribute
RUN cd /build/gh-contribute && \
  go build -o /usr/local/bin/gh-contribute ./cmd/gh-contribute && \
  rm -rf /build/gh-contribute

RUN chown -R dev:dev /workspace /home/dev/.claude /home/dev/go

# Install Claude Code as a native binary (no Node.js or npm required).
# CLAUDE_CODE_VERSION=latest fetches the newest release from GitHub.
# Set CLAUDE_CODE_VERSION to a specific version tag (e.g. "1.2.3") to pin.
# NOTE: confirm release asset naming at https://github.com/anthropics/claude-code/releases
# if the binary URL changes in a future release.
RUN set -eux; \
    ARCH=$(dpkg --print-architecture); \
    case "$ARCH" in \
        amd64) CC_ARCH="x64" ;; \
        arm64) CC_ARCH="arm64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac; \
    if [ "$CLAUDE_CODE_VERSION" = "latest" ]; then \
        TAG=$(curl -sf https://api.github.com/repos/anthropics/claude-code/releases/latest \
              | jq -r '.tag_name'); \
    else \
        TAG="v${CLAUDE_CODE_VERSION}"; \
    fi; \
    curl -fsSL \
        "https://github.com/anthropics/claude-code/releases/download/${TAG}/claude-linux-${CC_ARCH}" \
        -o /usr/local/bin/claude; \
    chmod +x /usr/local/bin/claude; \
    claude --version

# Copy setup scripts and grant passwordless sudo for them
COPY .devcontainer/shared/init-firewall.sh /usr/local/bin/
COPY .devcontainer/shared/setup.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/init-firewall.sh /usr/local/bin/setup.sh && \
  echo "dev ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh, /usr/local/bin/setup.sh" \
    > /etc/sudoers.d/dev-scripts && \
  chmod 0440 /etc/sudoers.d/dev-scripts

USER dev

ENV SHELL=/bin/zsh
ENV EDITOR=nano
ENV VISUAL=nano

# Install zsh with Powerline10k theme
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
  -p git \
  -p fzf \
  -a "source /usr/share/doc/fzf/examples/key-bindings.zsh" \
  -a "source /usr/share/doc/fzf/examples/completion.zsh" \
  -a "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  -x

# Install gopls for Go language server support
RUN go install golang.org/x/tools/gopls@latest
```

- [ ] **Step 3: Verify no npm/Node.js references remain**

```bash
grep -n "npm\|node\|NODE" .devcontainer/claude-code/Dockerfile
```

Expected: no matches.

- [ ] **Step 4: Verify COPY paths reference the correct new locations**

```bash
grep -n "COPY" .devcontainer/claude-code/Dockerfile
```

Expected output (3 COPY lines):
```
COPY .devcontainer/claude-code/install-deps.sh /tmp/install-deps.sh
COPY . /build/gh-contribute
COPY .devcontainer/shared/init-firewall.sh /usr/local/bin/
COPY .devcontainer/shared/setup.sh /usr/local/bin/
```

- [ ] **Step 5: Commit**

```bash
git add .devcontainer/claude-code/Dockerfile
git commit -m "refactor: move Dockerfile to claude-code/, replace npm Claude Code install with native binary, add gopls"
```

---

### Task 5: Rewrite `claude-code/devcontainer.json`

**Files:**
- Create: `.devcontainer/claude-code/devcontainer.json`

- [ ] **Step 1: Move the existing devcontainer.json**

```bash
git mv .devcontainer/devcontainer.json .devcontainer/claude-code/devcontainer.json
```

- [ ] **Step 2: Replace the entire file contents**

```json
{
  "name": "gh-contribute (claude-code)",
  "build": {
    "dockerfile": "Dockerfile",
    "context": "../..",
    "args": {
      "TZ": "${localEnv:TZ:Europe/Berlin}",
      "CLAUDE_CODE_VERSION": "latest",
      "GIT_DELTA_VERSION": "0.18.2",
      "ZSH_IN_DOCKER_VERSION": "1.2.0"
    }
  },
  "runArgs": [
    "--cap-add=NET_ADMIN",
    "--cap-add=NET_RAW",
    // Mount host Docker socket so Claude Code can run containers for testing.
    // Containers are placed on an isolated per-devcontainer bridge network (see setup.sh).
    "--mount=type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock"
  ],
  "customizations": {
    "jetbrains": {
      "backend": "GoLand"
    }
  },
  "remoteUser": "dev",
  "mounts": [
    // Persist shell history across rebuilds
    "source=gh-contribute-bashhistory-${devcontainerId},target=/commandhistory,type=volume",
    // Mount host ~/.claude so local plugins, settings, and auth are available inside the container
    "source=${localEnv:HOME}/.claude,target=/home/dev/.claude,type=bind,consistency=cached"
  ],
  "containerEnv": {
    // Claude Code — ANTHROPIC_API_KEY alone is enough for headless/API-key auth.
    // If you use OAuth (`claude auth login` on the host), the ~/.claude bind mount above
    // carries the session into the container automatically.
    "CLAUDE_CONFIG_DIR": "/home/dev/.claude",
    "ANTHROPIC_API_KEY": "${localEnv:ANTHROPIC_API_KEY}",

    // GitHub auth for gh-contribute.
    // Priority (from config/token.go): GH_CONTRIBUTE_TOKEN env var → ~/.config/gh-contribute/token file.
    // Set GH_CONTRIBUTE_TOKEN on the host for env-var auth, or run `gh-contribute auth login`
    // on the host once — the token file is shared via the bind mount above.
    "GH_CONTRIBUTE_TOKEN": "${localEnv:GH_CONTRIBUTE_TOKEN}",

    // Standard GitHub tokens forwarded for gh CLI and any other tooling.
    "GITHUB_TOKEN": "${localEnv:GITHUB_TOKEN}",
    "GH_TOKEN": "${localEnv:GH_TOKEN}",

    // Docker network name for test containers — set by setup.sh on start.
    // Use: docker run --network "$DOCKER_NETWORK" <image>
    // Containers on this network can reach each other but are isolated from
    // other devcontainers running on the same host Docker daemon.
    "DOCKER_NETWORK": "gh-contribute-${devcontainerId}",

    "POWERLEVEL9K_DISABLE_GITSTATUS": "true"
  },
  "workspaceMount": "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=delegated",
  "workspaceFolder": "/workspace",

  // setup.sh: creates the isolated Docker network, then locks down outbound traffic via iptables.
  // Running with sudo because iptables and Docker network creation require root.
  "postStartCommand": "sudo /usr/local/bin/setup.sh",
  "waitFor": "postStartCommand"
}
```

- [ ] **Step 3: Verify no VS Code or NODE_OPTIONS references remain**

```bash
grep -n "vscode\|NODE_OPTIONS\|extensions\|golang.go\|gitlens" .devcontainer/claude-code/devcontainer.json
```

Expected: no matches.

- [ ] **Step 4: Verify JetBrains block is present**

```bash
grep -n "jetbrains\|GoLand" .devcontainer/claude-code/devcontainer.json
```

Expected: both present.

- [ ] **Step 5: Verify ~/.claude mount is present**

```bash
grep -n "\.claude" .devcontainer/claude-code/devcontainer.json
```

Expected: the `source=${localEnv:HOME}/.claude` mount line.

- [ ] **Step 6: Commit**

```bash
git add .devcontainer/claude-code/devcontainer.json
git commit -m "refactor: move devcontainer.json to claude-code/, switch to JetBrains GoLand, add ~/.claude mount"
```

---

## Chunk 3: Default Variant, claude-session, and Cleanup

**Files:**
- Modify: `.devcontainer/default/devcontainer.json` (full rewrite)
- Create: `deploy/claude-session/Dockerfile` (copy of claude-code Dockerfile)
- Create: `deploy/claude-session/docker-compose.yml`

---

### Task 6: Rewrite `default/devcontainer.json`

**Files:**
- Modify: `.devcontainer/default/devcontainer.json`

- [ ] **Step 1: Replace the entire file contents**

```json
{
  "name": "gh-contribute (default)",
  "image": "golang:latest",
  "customizations": {
    "jetbrains": {
      "backend": "GoLand"
    }
  },
  "remoteUser": "root",
  "containerEnv": {
    "GOPATH": "/root/go",
    // Extend the base image PATH to include the user's Go bin directory.
    // golang:latest already includes /usr/local/go/bin; this adds /root/go/bin for
    // tools installed via `go install` (e.g. gopls).
    "PATH": "${containerEnv:PATH}:/root/go/bin"
  },
  // Install gopls on first create. Subsequent rebuilds skip this since the volume is cached.
  "postCreateCommand": "go install golang.org/x/tools/gopls@latest"
}
```

- [ ] **Step 2: Verify the file parses as valid JSON5/devcontainer spec**

```bash
cat .devcontainer/default/devcontainer.json
```

Check visually: correct braces, no stray commas, all fields present.

- [ ] **Step 3: Verify GoLand backend is set**

```bash
grep -n "GoLand\|jetbrains" .devcontainer/default/devcontainer.json
```

Expected: both present.

- [ ] **Step 4: Commit**

```bash
git add .devcontainer/default/devcontainer.json
git commit -m "refactor: rewrite default devcontainer.json for JetBrains GoLand with minimal Go setup"
```

---

### Task 7: Create `deploy/claude-session/` image

**Files:**
- Create: `deploy/claude-session/Dockerfile`
- Create: `deploy/claude-session/docker-compose.yml`

- [ ] **Step 1: Create the deploy directory**

```bash
mkdir -p deploy/claude-session
```

- [ ] **Step 2: Copy the Dockerfile from claude-code (identical — shared build context makes COPY paths work)**

```bash
cp .devcontainer/claude-code/Dockerfile deploy/claude-session/Dockerfile
git add deploy/claude-session/Dockerfile
```

- [ ] **Step 3: Verify the Dockerfile is truly identical**

```bash
diff .devcontainer/claude-code/Dockerfile deploy/claude-session/Dockerfile
```

Expected: no output (files are identical).

- [ ] **Step 4: Create `docker-compose.yml`**

```yaml
# Claude Code session image — isolated terminal environment.
#
# Usage:
#   docker compose run --rm claude-session
#
# Drops into zsh as the non-root `dev` user. From there:
#   - Clone a repo: git clone <url>
#   - Lock down outbound traffic: sudo init-firewall.sh
#   - Start Claude Code: claude
#
# The build context is set to the repo root so the Dockerfile can share
# COPY paths with the devcontainer Dockerfile without modification.
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

- [ ] **Step 5: Commit**

```bash
git add deploy/claude-session/
git commit -m "feat: add claude-session Docker image for isolated Claude Code terminal sessions"
```

---

### Task 8: Final cleanup verification

- [ ] **Step 1: Confirm the old root-level files are gone**

```bash
ls .devcontainer/
```

Expected: only `claude-code/`, `default/`, `shared/` — no `devcontainer.json`, `Dockerfile`, `install-deps.sh`, `setup.sh`, or `init-firewall.sh` at the root level.

- [ ] **Step 2: Confirm all expected files exist**

```bash
find .devcontainer deploy/claude-session -type f | sort
```

Expected:
```
.devcontainer/claude-code/Dockerfile
.devcontainer/claude-code/devcontainer.json
.devcontainer/claude-code/install-deps.sh
.devcontainer/default/devcontainer.json
.devcontainer/shared/init-firewall.sh
.devcontainer/shared/setup.sh
deploy/claude-session/Dockerfile
deploy/claude-session/docker-compose.yml
```

- [ ] **Step 3: Confirm no npm/Node.js references exist anywhere in the devcontainer tree**

```bash
grep -rn "npm\|nodejs\|NODE_VERSION\|NPM_CONFIG" .devcontainer/ deploy/claude-session/
```

Expected: no matches.

- [ ] **Step 4: Confirm the native binary install block is present in both Dockerfiles**

```bash
grep -n "claude-linux" .devcontainer/claude-code/Dockerfile deploy/claude-session/Dockerfile
```

Expected: both files show the binary download URL line.

- [ ] **Step 5: Confirm gopls is installed in both Dockerfiles**

```bash
grep -n "gopls" .devcontainer/claude-code/Dockerfile deploy/claude-session/Dockerfile
```

Expected: `go install golang.org/x/tools/gopls@latest` present in both.

- [ ] **Step 6: Verify all commits from this plan are present**

```bash
git log --oneline -8
```

Expected (most-recent first, exact messages may vary slightly):
```
feat: add claude-session Docker image for isolated Claude Code terminal sessions
refactor: rewrite default devcontainer.json for JetBrains GoLand with minimal Go setup
refactor: move devcontainer.json to claude-code/, switch to JetBrains GoLand, add ~/.claude mount
refactor: move Dockerfile to claude-code/, replace npm Claude Code install with native binary, add gopls
refactor: move install-deps.sh to claude-code/, remove Node.js install
refactor: move shared devcontainer scripts to shared/, remove npm from firewall allowlist
```

---

## Implementation Notes

### Claude Code native binary — release asset name

The plan uses `claude-linux-x64` / `claude-linux-arm64` as the asset name. **Verify this against the actual GitHub releases page before building:** `https://github.com/anthropics/claude-code/releases`. If the naming differs (e.g. `claude_linux_amd64`), update the `CC_ARCH` mapping in the Dockerfile accordingly.

### First devcontainer build

The first build will be slow: it fetches the Claude Code binary from GitHub, installs gopls, and sets up zsh. Subsequent builds use the Docker layer cache.

### `~/.claude` bind mount

The `claude-code` devcontainer mounts the host `~/.claude` into the container. If this directory doesn't exist on your host, create it first: `mkdir -p ~/.claude`. The mount will fail silently or cause errors if the source path doesn't exist.

### `docker compose run` vs `docker compose up`

The `claude-session` service is interactive — use `docker compose run --rm claude-session`, not `docker compose up`. `up` would start the container and immediately exit (no persistent foreground process).

### Smoke-testing the Claude Code binary

The Dockerfile runs `claude --version` as a build-time verification step. If the release asset name changes in a future Claude Code release (e.g. `claude_linux_amd64` instead of `claude-linux-x64`), the build will fail at that step with a "not found" or "permission denied" error. Check the releases page and update the `CC_ARCH` mapping in the Dockerfile accordingly.
