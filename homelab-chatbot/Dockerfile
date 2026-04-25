# syntax=docker/dockerfile:1.7

# Stage 1: frontend build
FROM node:24-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build
# Next.js output:"export" writes static assets to out/.

# Stage 2: backend dependency build
FROM python:3.13-slim-bookworm AS backend-deps
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
RUN apt-get update \
    && apt-get install -y --no-install-recommends git curl ca-certificates build-essential \
    && rm -rf /var/lib/apt/lists/*
# Install uv (Astral) for deterministic, fast installs.
COPY --from=ghcr.io/astral-sh/uv:0.4.27 /uv /usr/local/bin/uv
WORKDIR /app
COPY backend/pyproject.toml backend/uv.lock ./
RUN uv venv /opt/venv && \
    VIRTUAL_ENV=/opt/venv uv sync --frozen --no-install-project --python /opt/venv/bin/python

# Stage 3: final runtime
FROM python:3.13-slim-bookworm
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 \
    PATH="/opt/venv/bin:${PATH}" \
    VIRTUAL_ENV=/opt/venv \
    HLCB_DATA_DIR=/data \
    HLCB_CONFIG_PATH=/config/config.yaml
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl git ca-certificates tini tesseract-ocr \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 1000 --home-dir /app --shell /usr/sbin/nologin app \
    && mkdir -p /data /config /app/static \
    && chown -R app:app /data /config /app
COPY --from=backend-deps /opt/venv /opt/venv
WORKDIR /app
COPY --chown=app:app backend/ /app/
COPY --chown=app:app --from=frontend-builder /src/frontend/out /app/static
USER app
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8000/api/health || exit 1
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["uvicorn", "app.main:create_app", "--factory", "--host", "0.0.0.0", "--port", "8000"]
