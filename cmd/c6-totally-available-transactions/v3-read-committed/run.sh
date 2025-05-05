#!/bin/bash
set -euo pipefail

[ -z "${MAELSTROM_PATH:-}" ] && echo "MAELSTROM_PATH not set" && exit 1

rm -f bin
go build -o bin
BIN_PATH="${PWD}/bin"

cd "${MAELSTROM_PATH}"
# Run test with total availability and network partitions
./maelstrom test -w txn-rw-register \
    --bin "${BIN_PATH}" \
    --node-count 2 \
    --time-limit 20 \
    --rate 1000 \
    --concurrency 2n \
    --consistency-models read-committed \
    --availability total \
    --nemesis partition