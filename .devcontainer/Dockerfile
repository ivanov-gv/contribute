FROM node:20-bookworm

ARG TZ
ENV TZ="$TZ"

ARG CLAUDE_CODE_VERSION=latest
ARG GO_VERSION=1.24.1
ARG GIT_DELTA_VERSION=0.18.2
ARG ZSH_IN_DOCKER_VERSION=1.2.0

# Install system packages: Go build tools, Docker CLI, iptables for firewall, gh CLI, and dev tools
RUN apt-get update && apt-get install -y --no-install-recommends \
  less \
  git \
  procps \
  sudo \
  fzf \
  zsh \
  man-db \
  unzip \
  gnupg2 \
  curl \
  wget \
  iptables \
  ipset \
  iproute2 \
  dnsutils \
  aggregate \
  jq \
  nano \
  vim \
  gh \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Docker CLI (no daemon — uses host socket via mount)
RUN install -m 0755 -d /etc/apt/keyrings \
  && curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc \
  && chmod a+r /etc/apt/keyrings/docker.asc \
  && echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian bookworm stable" \
     > /etc/apt/sources.list.d/docker.list \
  && apt-get update \
  && apt-get install -y --no-install-recommends docker-ce-cli \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Install Go
RUN ARCH=$(dpkg --print-architecture | sed 's/amd64/amd64/;s/arm64/arm64/') && \
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" | tar -C /usr/local -xz
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/home/node/go
ENV PATH=$PATH:/home/node/go/bin

# Ensure node user has access to npm global dir
RUN mkdir -p /usr/local/share/npm-global && \
  chown -R node:node /usr/local/share

# Add node user to the docker group (created at runtime via socket mount)
RUN groupadd -f docker && usermod -aG docker node

ARG USERNAME=node

# Persist bash history
RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  && mkdir /commandhistory \
  && touch /commandhistory/.bash_history \
  && chown -R $USERNAME /commandhistory

# Signal that this is a devcontainer
ENV DEVCONTAINER=true

# Create workspace and config directories
RUN mkdir -p /workspace /home/node/.claude /home/node/go && \
  chown -R node:node /workspace /home/node/.claude /home/node/go

WORKDIR /workspace

# Install git-delta for improved diffs
RUN ARCH=$(dpkg --print-architecture) && \
  wget -q "https://github.com/dandavison/delta/releases/download/${GIT_DELTA_VERSION}/git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
  dpkg -i "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
  rm "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb"

# Build gh-contribute binary from source and install it system-wide
COPY . /build/gh-contribute
RUN cd /build/gh-contribute && \
  /usr/local/go/bin/go build -o /usr/local/bin/gh-contribute ./cmd/gh-contribute && \
  rm -rf /build/gh-contribute

USER node

# Configure npm global path
ENV NPM_CONFIG_PREFIX=/usr/local/share/npm-global
ENV PATH=$PATH:/usr/local/share/npm-global/bin
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

# Install Claude Code globally
RUN npm install -g @anthropic-ai/claude-code@${CLAUDE_CODE_VERSION}

# Copy setup scripts and grant passwordless sudo for them
COPY init-firewall.sh /usr/local/bin/
COPY setup.sh /usr/local/bin/
USER root
RUN chmod +x /usr/local/bin/init-firewall.sh /usr/local/bin/setup.sh && \
  echo "node ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh, /usr/local/bin/setup.sh" \
    > /etc/sudoers.d/node-scripts && \
  chmod 0440 /etc/sudoers.d/node-scripts
USER node
