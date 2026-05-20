#!/usr/bin/env bash
set -euo pipefail

echo "==> Auth Service setup"

# Check deps
for cmd in go docker docker-compose; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "ERROR: $cmd not found. Please install it."
    exit 1
  fi
done

# Create .env from example if not present
if [ ! -f .env ]; then
  cp .env.example .env
  echo "Created .env from .env.example — edit secrets before running!"
fi

# Start infra
echo "==> Starting Postgres + Redis..."
docker-compose up -d postgres redis

# Wait for postgres
echo "==> Waiting for PostgreSQL..."
until docker-compose exec -T postgres pg_isready -U postgres; do sleep 1; done

echo "==> Building Go binary..."
go build -o bin/auth-service ./cmd/server

echo ""
echo "Setup complete! Run: make run"
echo "Then open: frontend/public/index.html"
