#!/usr/bin/env bash
# Spark-OS (SOS) Automated Runlevels Deployment Script
set -euo pipefail

# Configuration
VPS_IP="46.224.219.174"
REMOTE_OPT="/opt/sos"
REMOTE_ETC="/etc/sos"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOS_WSL_SOURCE_ROOT="${SOS_WSL_SOURCE_ROOT:-/home/sos/sources/pqr.info}"

"$SCRIPT_DIR/scripts/prepare_sos_wsl_source.sh" "$SOS_WSL_SOURCE_ROOT"

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

echo "=== [PQRL0] Preparing Server Directory Topology ==="
ssh -T -o StrictHostKeyChecking=no root@"$VPS_IP" << 'EOF'
  mkdir -p /opt/sos/mev
  mkdir -p /opt/sos/jetweb-time-machine
  mkdir -p /opt/sos/substrate-node-template
  mkdir -p /etc/sos
  mkdir -p /var/sos
EOF

echo "=== [PQRL1] Uploading and Extracting Codebase Tarballs ==="
rm -f /tmp/mev.tar.gz /tmp/jetweb-time-machine.tar.gz /tmp/substrate-node-template.tar.gz

tar -czf /tmp/mev.tar.gz -C "$SOS_WSL_SOURCE_ROOT" mev
if [ -d "$SOS_WSL_SOURCE_ROOT/jetweb-time-machine" ]; then
  tar -czf /tmp/jetweb-time-machine.tar.gz -C "$SOS_WSL_SOURCE_ROOT" jetweb-time-machine
fi
if [ -d "$SOS_WSL_SOURCE_ROOT/substrate-node-template" ]; then
  tar -czf /tmp/substrate-node-template.tar.gz -C "$SOS_WSL_SOURCE_ROOT" substrate-node-template
fi

scp -o StrictHostKeyChecking=no /tmp/mev.tar.gz root@"$VPS_IP":/tmp/
if [ -f /tmp/jetweb-time-machine.tar.gz ]; then
  scp -o StrictHostKeyChecking=no /tmp/jetweb-time-machine.tar.gz root@"$VPS_IP":/tmp/
fi
if [ -f /tmp/substrate-node-template.tar.gz ]; then
  scp -o StrictHostKeyChecking=no /tmp/substrate-node-template.tar.gz root@"$VPS_IP":/tmp/
fi
scp -o StrictHostKeyChecking=no "$SOS_WSL_SOURCE_ROOT/runlevels.toml" root@"$VPS_IP":"$REMOTE_ETC"/runlevels.toml

ssh -T -o StrictHostKeyChecking=no root@"$VPS_IP" << EOF
  tar -xzf /tmp/mev.tar.gz -C "$REMOTE_OPT"/mev/
  tar -xzf /tmp/jetweb-time-machine.tar.gz -C "$REMOTE_OPT"/
  tar -xzf /tmp/substrate-node-template.tar.gz -C "$REMOTE_OPT"/
EOF

echo "=== [PQRL5] Initializing Configuration & State Spine ==="
ssh -T -o StrictHostKeyChecking=no root@"$VPS_IP" << EOF
  cd "$REMOTE_OPT"/mev
  cp .env.example .env
  # Update environment variables
  sed -i 's|HETZNER_API_KEY=.*|HETZNER_API_KEY=NWp5z8GM8t5ufitsknndPyrsk62W7uIn27cIiv5cURKWAGxzTkmpw9LfpyR9S2ww|' .env
EOF

echo "=== [PQRL7] Starting Core Consensus & MEV Services (Docker Stack) ==="
ssh -T -o StrictHostKeyChecking=no root@"$VPS_IP" << 'EOF'
  cd "$REMOTE_OPT"/mev
  echo "Booting substrate-node & mev-engine..."
  docker compose -f docker-compose.prod.yml up -d --build substrate-node mev-engine
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

  wait_for_container "substrate-node"
  wait_for_container "mev-engine"

  echo "Booting bcpd, gemini-agentd, mev-node, time-machine-go, mesh-adapter, and markets sidecar..."
  docker compose -f docker-compose.prod.yml up -d --build bcpd gemini-agentd mev-node time-machine-go mesh-adapter markets
  wait_for_container "bcpd"
  wait_for_container "gemini-agentd"
  wait_for_container "mev-node"
  wait_for_container "time-machine-go"
  wait_for_container "mesh-adapter"
  wait_for_container "markets"
EOF

echo "=== [PQRL9] Booting Observability Infrastructure & Testing Heath ==="
ssh -T -o StrictHostKeyChecking=no root@"$VPS_IP" << EOF
  cd "$REMOTE_OPT"/mev
  docker compose -f docker-compose.prod.yml up -d prometheus grafana
  
  echo "Verifying health endpoints..."
  curl -s -X POST http://localhost:8080/healer/trigger-recovery

  echo "Verifying backchannel spine and agent status gates..."
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

  if [ -n "${SOS_TIMESLIP_ENDPOINT:-}" ]; then
    curl -s -X POST "$SOS_TIMESLIP_ENDPOINT" -H 'Content-Type: application/json' -d '{"service":"sos-deploy","event":"deployment_completed","ts":"'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"}' >/dev/null 2>&1 || true
  fi
  
  echo "SOVEREIGN_READY" > /var/sos/SOVEREIGN_READY
  echo "SOS stack is running and healthy!"
EOF
