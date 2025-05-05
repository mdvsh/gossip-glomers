#!/bin/bash
set -euo pipefail

[ -z "${MAELSTROM_PATH:-}" ] && echo "MAELSTROM_PATH not set" && exit 1

rm -f bin
go build -o bin
BIN_PATH="${PWD}/bin"

cd "${MAELSTROM_PATH}"
# Run standard test with multiple nodes
./maelstrom test -w txn-rw-register \
    --bin "${BIN_PATH}" \
    --node-count 2 \
    --time-limit 20 \
    --rate 1000 \
    --concurrency 2n \
    --consistency-models read-uncommitted

# Run test with network partitions
./maelstrom test -w txn-rw-register \
    --bin "${BIN_PATH}" \
    --node-count 2 \
    --time-limit 20 \
    --rate 1000 \
    --concurrency 2n \
    --consistency-models read-uncommitted \
    --availability total \
    --nemesis partition