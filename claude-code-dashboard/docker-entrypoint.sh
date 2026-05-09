#!/bin/sh
set -e

# Start Fastify backend in background
node /app/backend/dist/server.js &

# Start Nginx in foreground
nginx -g "daemon off;"
