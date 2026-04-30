#!/bin/bash

ITERATIONS=5
PAYLOAD_SIZE=200
BATCH_SIZE=10000

cleanup() {
    echo -e "\n🛑 Signal received. Cleaning up $ITERATIONS benchmark workers..."
    trap - SIGINT SIGTERM
    kill 0
    exit 1
}

trap cleanup SIGINT SIGTERM

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -i, --iterations    Number of concurrent benchmark workers (default: $ITERATIONS)"
    echo "  -p, --payload       Payload size in bytes (default: $PAYLOAD_SIZE)"
    echo "  -b, --batch         Batch size for producer (default: $BATCH_SIZE)"
    echo "  -h, --help          Display this help message"
    exit 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -i|--iterations) ITERATIONS="$2"; shift 2 ;;
        -p|--payload)    PAYLOAD_SIZE="$2"; shift 2 ;;
        -b|--batch)      BATCH_SIZE="$2";   shift 2 ;;
        -h|--help)       usage ;;
        *) echo "Unknown parameter: $1"; usage ;;
    esac
done

echo "🧪 Starting Pressure Benchmark Suite"
echo "Strategy: $ITERATIONS parallel workers simulating distributed producers"
echo "Config:   Payload=${PAYLOAD_SIZE}B, Batch=$BATCH_SIZE"
echo "------------------------------------------------------------"

for ((i=1; i<=ITERATIONS; i++)); do
    echo "  [Worker $i] Launching bench process..."
    docker compose --profile tools run -T --rm producer \
        bench --payload-size "$PAYLOAD_SIZE" --batch-size "$BATCH_SIZE" &
done

echo "⏳ Benchmark active. Waiting for workers to complete (PID: $$)..."
wait
echo "------------------------------------------------------------"
echo "✅ All benchmark iterations finished successfully."
