FROM debian:bookworm-slim

LABEL description="Claude Code autonomous agent container"

ENV DEBIAN_FRONTEND=noninteractive

# ─── Layer 1: System packages + Python 3 ─────────────────────────────────────
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    git \
    ca-certificates \
    gnupg \
    build-essential \
    libssl-dev \
    unzip \
    jq \
    procps \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    libicu-dev \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 2: Node.js 24 LTS ─────────────────────────────────────────────────
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 3: .NET SDK 9 ─────────────────────────────────────────────────────
ENV DOTNET_ROOT=/usr/local/dotnet
ENV PATH="${PATH}:/usr/local/dotnet:/usr/local/dotnet/tools:/root/.local/bin:/root/.dotnet/tools"

RUN curl -fsSL https://dot.net/v1/dotnet-install.sh \
    | bash -s -- --channel 9.0 --install-dir "$DOTNET_ROOT" --no-path

# ─── Layer 4: Python uv ──────────────────────────────────────────────────────

RUN curl -fsSL https://astral.sh/uv/install.sh | UV_VERSION="0.11.8" bash \
    && uv --version

# ─── Layer 5: GitHub CLI ─────────────────────────────────────────────────────
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 6: Language servers & dev tooling ─────────────────────────────────
RUN npm install -g \
    typescript \
    typescript-language-server \
    pyright \
    @angular/cli \
    && npm cache clean --force

RUN dotnet tool install --global csharp-ls --version 0.20.0

# ─── Layer 7: Claude Code (native binary) ────────────────────────────────────
RUN curl -fsSL https://claude.ai/install.sh | bash \
    && claude --version \
    && rm -f /root/.claude.json

# ─── Layer 8: Config + entrypoint ────────────────────────────────────────────
WORKDIR /workspace

COPY config/settings.json /root/.claude/settings.json
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
