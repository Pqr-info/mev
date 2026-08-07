#!/usr/bin/env bash
# run-local-sos.sh — Orchestrates the Sovereign OS stack locally inside Git Bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "==============================================="
echo "   Sovereign OS — Local Stack Launcher"
echo "==============================================="

# 1. Verify Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "Error: Docker daemon is not running. Please start Docker Desktop first."
    exit 1
fi

# 2. Check for .env file
if [ ! -f .env ]; then
    echo "Warning: .env file not found. Copying .env.example..."
    cp .env.example .env
fi

# 3. Handle action arguments
ACTION="${1:-up}"

if [ "$ACTION" = "up" ]; then
    echo "[*] Spinning up all services (mev-engine, mev-node, prometheus, grafana)..."
    docker compose -f docker/docker-compose.yml up -d --build
    echo "[+] Services started successfully!"
    echo "    - Grafana Dashboard: http://localhost:3000 (admin / mev_admin)"
    echo "    - Prometheus Server: http://localhost:9090"
    echo "    - MEV node Metrics:  http://localhost:9091/metrics"
    echo ""
    echo "    To follow logs, run: ./run-local-sos.sh logs"
elif [ "$ACTION" = "down" ]; then
    echo "[*] Tearing down the local stack..."
    docker compose -f docker/docker-compose.yml down
    echo "[+] Stack removed."
elif [ "$ACTION" = "logs" ]; then
    docker compose -f docker/docker-compose.yml logs -f
elif [ "$ACTION" = "build" ]; then
    echo "[*] Building stack containers..."
    docker compose -f docker/docker-compose.yml build
else
    echo "Usage: ./run-local-sos.sh [up|down|logs|build]"
    exit 1
fi
