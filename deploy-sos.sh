#!/usr/bin/env bash
# Spark-OS (SOS) Automated Runlevels Deployment Script
set -euo pipefail

# Configuration
VPS_IP="46.224.219.174"
REMOTE_OPT="/opt/sos"
REMOTE_ETC="/etc/sos"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOS_WSL_SOURCE_ROOT="${SOS_WSL_SOURCE_ROOT:-/home/sos/sources/pqr.info}"

echo "=== Preparing Source tree ==="
"$SCRIPT_DIR/scripts/prepare_sos_wsl_source.sh" "$SOS_WSL_SOURCE_ROOT"

echo "=== [PQRL0] Preparing Server Directory Topology ==="
ssh -n -T -o StrictHostKeyChecking=no root@"$VPS_IP" "mkdir -p /opt/sos/mev /opt/sos/jetweb-time-machine /opt/sos/substrate-node-template /etc/sos /var/sos"

echo "=== [PQRL1] Uploading Codebase Tarballs ==="
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

echo "=== [PQRL5] Transferring & running remote deployment ==="
scp -o StrictHostKeyChecking=no "$SOS_WSL_SOURCE_ROOT/mev/deploy_remote.sh" root@"$VPS_IP":/tmp/deploy_remote.sh
ssh -n -T -o StrictHostKeyChecking=no root@"$VPS_IP" "bash /tmp/deploy_remote.sh"
