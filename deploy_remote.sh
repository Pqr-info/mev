#!/usr/bin/env bash
set -euo pipefail

REMOTE_OPT="/opt/sos"
REMOTE_ETC="/etc/sos"

wait_for_container() {
  local container_name="$1"
  local timeout_seconds="${2:-180}"
  local deadline=$((SECONDS + timeout_seconds))

  echo "Waiting for ${container_name} to become healthy..."
  while [ $SECONDS -lt $deadline ]; do
    status=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$container_name" 2>/dev/null || true)
    if [ "$status" = "healthy" ]; then
      echo "${container_name} is healthy"
      return 0
    fi
    sleep 5
  done

  echo "Timed out waiting for ${container_name} to become healthy" >&2
  docker compose -f docker-compose.prod.yml ps
  return 1
}

echo "=== [PQRL0] Setting up directories ==="
mkdir -p "$REMOTE_OPT"/mev
mkdir -p "$REMOTE_OPT"/jetweb-time-machine
mkdir -p "$REMOTE_OPT"/substrate-node-template
mkdir -p "$REMOTE_ETC"
mkdir -p /var/sos

echo "=== [PQRL1] Cleaning and Extracting archives ==="
rm -rf "$REMOTE_OPT"/mev "$REMOTE_OPT"/jetweb-time-machine "$REMOTE_OPT"/substrate-node-template
mkdir -p "$REMOTE_OPT"/mev "$REMOTE_OPT"/jetweb-time-machine "$REMOTE_OPT"/substrate-node-template
tar -xzf /tmp/mev.tar.gz -C "$REMOTE_OPT"/
tar -xzf /tmp/jetweb-time-machine.tar.gz -C "$REMOTE_OPT"/
tar -xzf /tmp/substrate-node-template.tar.gz -C "$REMOTE_OPT"/

# Move runlevels to /etc/sos
cp /tmp/runlevels.toml "$REMOTE_ETC"/runlevels.toml

echo "=== [PQRL5] Configuring environment vars ==="
cd "$REMOTE_OPT"/mev
cp .env.example .env
sed -i 's|HETZNER_API_KEY=.*|HETZNER_API_KEY=NWp5z8GM8t5ufitsknndPyrsk62W7uIn27cIiv5cURKWAGxzTkmpw9LfpyR9S2ww|' .env

echo "=== [PQRL7] Starting core services in Runlevel 7 ==="
# Build and run consensus and backend
docker compose -f docker-compose.prod.yml up -d --build substrate-node mev-engine
wait_for_container "substrate-node"
wait_for_container "mev-engine"

# Build and run node, adapters, and sidecars
docker compose -f docker-compose.prod.yml up -d --build bcpd gemini-agentd mev-node time-machine-go mesh-adapter markets
wait_for_container "bcpd"
wait_for_container "gemini-agentd"
wait_for_container "mev-node"
wait_for_container "time-machine-go"
wait_for_container "mesh-adapter"
wait_for_container "markets"

echo "=== [PQRL9] Starting telemetry and ready flag ==="
docker compose -f docker-compose.prod.yml up -d prometheus grafana

echo "SOVEREIGN_READY" > /var/sos/SOVEREIGN_READY
echo "VERIFICATION: Triggering healer recovery test on adapter..."
sleep 5
curl -s -X POST http://localhost:8080/healer/trigger-recovery
echo ""

echo "Verifying health endpoints..."
if ! curl -sf http://localhost:8082/ping >/dev/null; then
  echo "Error: bcpd is unreachable or degraded" >&2
  exit 1
fi

if ! curl -sf http://localhost:8080/ping >/dev/null; then
  echo "Error: mesh-adapter is unreachable" >&2
  exit 1
fi

STATE_JSON=$(curl -sf http://localhost:8082/state || echo "")
if [ -n "$STATE_JSON" ]; then
  GEMINI_STATUS=$(echo "$STATE_JSON" | grep -o '"status": "[^"]*"' | head -n 1 | cut -d'"' -f4 || echo "offline")
  if [ "$GEMINI_STATUS" != "online" ]; then
    echo "Error: Gemini Agent status is: $GEMINI_STATUS (expected online)" >&2
    exit 1
  fi
fi

echo "SOS stack started successfully!"
