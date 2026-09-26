#!/bin/bash
set -e
ROOT="${1:?Usage: install.sh <root-dir>}"

command -v go >/dev/null || apt-get install -y golang-go
mkdir -p "$ROOT/usr/local/bin"
CGO_ENABLED=0 go build -o "$ROOT/usr/local/bin/forge" .
