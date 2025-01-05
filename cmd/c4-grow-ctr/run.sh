#!/bin/bash
set -euo pipefail

[ -z "${MAELSTROM_PATH:-}" ] && echo "MAELSTROM_PATH not set" && exit 1

rm -f bin
go build -o bin
BIN_PATH="${PWD}/bin"

cd "${MAELSTROM_PATH}"
./maelstrom test -w g-counter \
    --bin "${BIN_PATH}" \
    --time-limit 20 \
    --rate 100 \
    --node-count 3 \
    --nemesis partition \
