#!/bin/bash
set -euo pipefail

[ -z "${MAELSTROM_PATH:-}" ] && echo "MAELSTROM_PATH not set" && exit 1

rm -f bin
go build -o bin
BIN_PATH="${PWD}/bin"

cd "${MAELSTROM_PATH}"
./maelstrom test -w broadcast \
    --bin "${BIN_PATH}" \
    --node-count 5 \
    --time-limit 20 \
    --rate 10 \
    --nemesis partition