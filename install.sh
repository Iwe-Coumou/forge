#!/bin/bash
# install.sh in a tool repo. Usage: ./install.sh <root-dir>
set -e
ROOT="$1"
CGO_ENABLED=0 go build -o "$ROOT/usr/local/bin/forge" .
