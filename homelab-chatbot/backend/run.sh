#!/bin/bash

if [ "$1" == "-i" ]; then
  uv sync
fi

HLCB_CONFIG_PATH=../config/config.dev.yaml \
  uv run --env-file ../.env \
  uvicorn app.main:create_app --factory --reload --port 8000
