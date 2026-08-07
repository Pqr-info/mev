#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
TARGET_ROOT="${1:-/home/sos/sources/pqr.info}"
MEV_TARGET="$TARGET_ROOT/mev"

mkdir -p "$TARGET_ROOT"

echo "Preparing sanitized SOS source tree under $TARGET_ROOT"

copy_tree() {
  local src="$1"
  local dest="$2"
  if [ -e "$src" ]; then
    mkdir -p "$dest"
    rsync -a --delete \
      --exclude '.git' \
      --exclude '.github' \
      --exclude 'node_modules' \
      --exclude '.venv' \
      --exclude '__pycache__' \
      --exclude '*.pyc' \
      --exclude '.DS_Store' \
      --exclude 'target' \
      --exclude 'bin' \
      "$src/" "$dest/"
  fi
}

copy_tree "$REPO_ROOT" "$MEV_TARGET"
cp "$REPO_ROOT/runlevels.toml" "$TARGET_ROOT/runlevels.toml" 2>/dev/null || true

if [ -d "$REPO_ROOT/../jetweb-time-machine" ]; then
  copy_tree "$REPO_ROOT/../jetweb-time-machine" "$TARGET_ROOT/jetweb-time-machine"
fi

if [ -d "$REPO_ROOT/../substrate-node-template" ]; then
  copy_tree "$REPO_ROOT/../substrate-node-template" "$TARGET_ROOT/substrate-node-template"
fi

echo "Sanitized SOS source is ready at $TARGET_ROOT"
