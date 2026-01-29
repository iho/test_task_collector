# Telemetry System

A simple telemetry system consisting of a Sensor Node (producer) and a Telemetry Sink (consumer) written in Go.

## Features
- **gRPC** communication.
- **Rate limiting** in Sink (bytes/sec).
- **Buffered IO** with periodic flushing to log file.
- **Log Encryption** (AES-256-GCM, optional).
- **mTLS** support (optional).
- **Graceful Shutdown**.

## Quick Start (Makefile)

Generate certificates and build everything:
```bash
make gen-certs build
```

Run Sink (with mTLS):
```bash
make run-sink
```

Run Sensor (with mTLS):
```bash
make run-sensor
```

Run Multiple Sensors (simulates load):
```bash
make run-multiple-sensors COUNT=5 RATE=10
```

Run Stress Test (k6):
```bash
make stress-test
```

Run Stress Test with mTLS (k6):
```bash
make stress-test-mtls
```
(Requires [k6](https://k6.io/) installed)

## Docker Compose
Run the entire system (Sink + 2 Sensors + Postgres) with Docker:
```bash
docker-compose up --build
```
*Note: This runs the core services. Stress tests are excluded by default using Docker profiles.*

Run Stress Test in Docker:
```bash
make docker-stress-test
# OR
docker-compose run --rm stress-test
```

Run Stress Test with mTLS in Docker:
```bash
make docker-stress-test-mtls
# OR
docker-compose run --rm stress-test-mtls
```

## Testing (gotestsum)

Run tests locally (auto-installs gotestsum):
```bash
make test
```

Run tests in Docker:
```bash
docker-compose run tests
```

## Linting

```bash
make lint
```

## Manual Building

```bash
go build -o bin/sink ./cmd/sink
go build -o bin/sensor ./cmd/sensor
```

## Running

### Telemetry Sink
```bash
./bin/sink --bind :50051 --log telemetry.log --rate-limit 10240
```
Flags:
- `--bind`: Address to bind (default `:50051`).
- `--log`: Output log path (default `telemetry.log`).
- `--buffer`: Buffer size in bytes (default `4096`).
- `--flush`: Flush interval (default `100ms`).
- `--rate-limit`: Max bytes/sec (default `0` = unlimited).
- `--encrypt-key`: 32-byte hex key for log encryption.
- `--cert`, `--key`, `--ca`: TLS paths.

### Sensor Node
```bash
./bin/sensor --sink localhost:50051 --rate 10 --name sensor-alpha
```
Flags:
- `--sink`: Sink address (default `localhost:50051`).
- `--rate`: Messages per second.
- `--name`: Sensor name.
- `--cert`, `--key`, `--ca`: TLS paths.

## Testing

Run all unit and integration tests:
```bash
go test -v ./...
```

## Security Examples

### Encrypted Logs
Generate a key:
```bash
openssl rand -hex 32
```
Run sink:
```bash
./bin/sink --encrypt-key <KEY_FROM_ABOVE>
```

### mTLS
To use mTLS, generate CA, server, and client certificates, then run:

Sink:
```bash
./bin/sink --cert server.crt --key server.key --ca ca.crt
```

Sensor:
```bash
./bin/sensor --cert client.crt --key client.key --ca ca.crt
```

## Task 2: Database & SQL Analysis

This project includes a secondary task to demonstrate SQL skills (PostgreSQL).

### Components
- **Schema**: `sql/schema.sql` (Rooms, Sensors, Measurements).
- **Queries**: `sql/query.sql` (Aggregation and Ohm's Law calculation).
- **Tooling**: `sqlc` used for Go code generation.

### Running Verification
To verify Task 2 (apply schema, seed data, run analysis):

1. Ensure Docker is running.
2. Run the verification tool:
```bash
# Starts Postgres and runs the Go verification tool
docker-compose up -d postgres
go run ./cmd/task2/main.go
```

Expected output will show aggregated sensor data and calculated current (I).

### Development
To generate Go code from SQL (requires `sqlc`):
```bash
sqlc generate
```
