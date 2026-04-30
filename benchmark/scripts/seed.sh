#!/bin/bash

ITERATIONS=10
COUNT=5000000
PAYLOAD_SIZE=200
BATCH_SIZE=10000

cleanup() {
    echo -e "\n🛑 Signal received! Cleaning up $ITERATIONS background producers..."
    trap - SIGINT SIGTERM
    kill 0
    exit 1
}

trap cleanup SIGINT SIGTERM

usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -i, --iterations    Number of concurrent producer instances (default: $ITERATIONS)"
    echo "  -c, --count         Total events to seed per instance (default: $COUNT)"
    echo "  -p, --payload       Size of each payload in bytes (default: $PAYLOAD_SIZE)"
    echo "  -b, --batch         Events per database transaction/batch (default: $BATCH_SIZE)"
    echo "  -h, --help          Display this help message"
    exit 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -i|--iterations) ITERATIONS="$2"; shift 2 ;;
        -c|--count)      COUNT="$2";      shift 2 ;;
        -p|--payload)    PAYLOAD_SIZE="$2"; shift 2 ;;
        -b|--batch)      BATCH_SIZE="$2";   shift 2 ;;
        -h|--help)       usage ;;
        *) echo "Unknown parameter: $1"; usage ;;
    esac
done

echo "🚀 Starting $ITERATIONS producers..."
echo "Config: Count=$COUNT, Payload=${PAYLOAD_SIZE}B, Batch=$BATCH_SIZE"

for ((i=1; i<=ITERATIONS; i++)); do
    docker compose --profile tools run -T --rm producer \
        seed --count "$COUNT" \
        --payload-size "$PAYLOAD_SIZE" \
        --batch-size "$BATCH_SIZE" &
done

echo "⏳ Waiting for all producers to finish (PID: $$)..."
wait
echo "✅ Seeding complete."
