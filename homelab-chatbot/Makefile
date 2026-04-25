# Makefile for homelab-chatbot

IMAGE       ?= adeotek/homelab-chatbot
TAG         ?= dev
BACKEND_DIR := backend
FRONTEND_DIR:= frontend

.PHONY: help all build build-backend build-frontend docker-build \
        run dev test test-backend test-frontend test-integration \
        lint lint-backend lint-frontend fmt fmt-backend fmt-frontend \
        deps deps-backend deps-frontend clean

help:
	@echo "Targets:"
	@echo "  build           - Build frontend static assets (out/) and sync backend deps"
	@echo "  docker-build    - Build the Docker image ($(IMAGE):$(TAG))"
	@echo "  run             - docker compose up (foreground)"
	@echo "  dev             - Print dev-server launch commands"
	@echo "  test            - Run backend + frontend unit tests"
	@echo "  test-integration- Run integration tests (requires RUN_INTEGRATION_TESTS=1)"
	@echo "  lint            - Lint backend (ruff) and frontend (eslint)"
	@echo "  fmt             - Format backend (ruff format) and frontend (prettier)"
	@echo "  clean           - Remove build artifacts"

all: build

build: build-frontend build-backend

build-backend:
	cd $(BACKEND_DIR) && uv sync

build-frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

docker-build:
	docker build -t $(IMAGE):$(TAG) .

run:
	docker compose up --build

dev:
	@echo "Start backend:  cd $(BACKEND_DIR) && uv run uvicorn app.main:create_app --factory --reload"
	@echo "Start frontend: cd $(FRONTEND_DIR) && npm run dev"

test: test-backend test-frontend

test-backend:
	cd $(BACKEND_DIR) && uv run pytest -q

test-frontend:
	cd $(FRONTEND_DIR) && npm run test -- --run

test-integration:
	cd $(BACKEND_DIR) && RUN_INTEGRATION_TESTS=1 uv run pytest -v tests/integration

lint: lint-backend lint-frontend

lint-backend:
	cd $(BACKEND_DIR) && uv run ruff check .

lint-frontend:
	cd $(FRONTEND_DIR) && npm run lint

fmt: fmt-backend fmt-frontend

fmt-backend:
	cd $(BACKEND_DIR) && uv run ruff format .

fmt-frontend:
	cd $(FRONTEND_DIR) && npm run format

deps: deps-backend deps-frontend

deps-backend:
	cd $(BACKEND_DIR) && uv sync

deps-frontend:
	cd $(FRONTEND_DIR) && npm ci

clean:
	rm -rf $(BACKEND_DIR)/.pytest_cache $(BACKEND_DIR)/.ruff_cache
	rm -rf $(FRONTEND_DIR)/.next $(FRONTEND_DIR)/out $(FRONTEND_DIR)/node_modules
