#!/bin/bash
# Usage:
#   ./install.sh              install for the current user (~/.local/bin)
#   ./install.sh <root-dir>   install into a system image (e.g. ghostkey ISO build)
set -e
NAME="forge"

if [ -n "${1:-}" ]; then
  BINDIR="$1/usr/local/bin"
else
  BINDIR="$HOME/.local/bin"
fi

if ! command -v go >/dev/null; then
  if [ "$(id -u)" = 0 ] && command -v apt-get >/dev/null; then
    apt-get install -y golang-go
  else
    echo "Go is required: https://go.dev/dl/"; exit 1
  fi
fi

mkdir -p "$BINDIR"
CGO_ENABLED=0 go build -o "$BINDIR/$NAME" .
echo "Installed $NAME to $BINDIR"
