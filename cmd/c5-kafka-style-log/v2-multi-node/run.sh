#!/bin/bash
set -euo pipefail

[ -z "${MAELSTROM_PATH:-}" ] && echo "MAELSTROM_PATH not set" && exit 1

rm -f bin
go build -o bin
BIN_PATH="${PWD}/bin"

cd "${MAELSTROM_PATH}"
./maelstrom test -w kafka \
    --bin "${BIN_PATH}" \
    --time-limit 20 \
    --rate 1000 \
    --node-count 2 \
    --concurrency 2n \
