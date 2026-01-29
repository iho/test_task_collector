#!/bin/bash

# Default to 3 sensors if not specified
COUNT=${1:-3}
RATE=${2:-5}

BIN_DIR=./bin
CERTS_DIR=./certs

echo "Starting $COUNT sensors with rate $RATE msg/sec..."

for i in $(seq 1 $COUNT); do
    NAME="sensor-$i"
    echo "Launching $NAME..."
    $BIN_DIR/sensor \
        --sink localhost:50051 \
        --rate $RATE \
        --name $NAME \
        --cert $CERTS_DIR/client.crt \
        --key $CERTS_DIR/client.key \
        --ca $CERTS_DIR/ca.crt &
done

echo "$COUNT sensors running in background."
echo "Press Ctrl+C to stop all sensors."

# Wait for signal to kill all children
trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT
wait
