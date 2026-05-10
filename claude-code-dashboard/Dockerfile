# Stage 1: build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: build backend
FROM node:22-alpine AS backend-build
WORKDIR /app/backend
# node-pty requires python3 and build tools
RUN apk add --no-cache python3 make g++
COPY backend/package*.json ./
RUN npm ci
COPY backend/ ./
RUN npm run build

# Stage 3: production image
FROM node:22-alpine

WORKDIR /app/backend
COPY --from=backend-build /app/backend/node_modules ./node_modules
COPY --from=backend-build /app/backend/dist ./dist

# Place frontend dist where server.ts expects it relative to dist/
COPY --from=frontend-build /app/frontend/dist /app/frontend/dist

EXPOSE 9998

CMD ["node", "/app/backend/dist/server.js"]
