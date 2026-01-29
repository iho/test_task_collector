# Variables
BIN_DIR := bin
CERTS_DIR := certs
PROTO_DIR := proto

SINK_BINARY := $(BIN_DIR)/sink
SENSOR_BINARY := $(BIN_DIR)/sensor

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(SINK_BINARY) ./cmd/sink
	go build -o $(SENSOR_BINARY) ./cmd/sensor
	@echo "Build complete."

proto:
	@mkdir -p $(PROTO_DIR)/telemetry
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       $(PROTO_DIR)/telemetry.proto
	@mkdir -p $(PROTO_DIR)/telemetry
	@if [ -f "proto/telemetry.pb.go" ]; then mv proto/*.pb.go $(PROTO_DIR)/telemetry/; fi

gen-certs:
	@mkdir -p $(CERTS_DIR)
	@echo "Generating CA..."
	openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout $(CERTS_DIR)/ca.key -out $(CERTS_DIR)/ca.crt -subj "/C=US/ST=State/L=City/O=Org/OU=Unit/CN=ca"
	
	@echo "subjectAltName=DNS:localhost,IP:127.0.0.1" > $(CERTS_DIR)/server-ext.cnf
	openssl x509 -req -in $(CERTS_DIR)/server.csr -CA $(CERTS_DIR)/ca.crt -CAkey $(CERTS_DIR)/ca.key -CAcreateserial -out $(CERTS_DIR)/server.crt -days 365 -extfile $(CERTS_DIR)/server-ext.cnf
	@rm $(CERTS_DIR)/server-ext.cnf

	@echo "Generating Client Cert..."
	openssl req -newkey rsa:4096 -nodes -keyout $(CERTS_DIR)/client.key -out $(CERTS_DIR)/client.csr -subj "/C=US/ST=State/L=City/O=Org/OU=Unit/CN=client"
	openssl x509 -req -in $(CERTS_DIR)/client.csr -CA $(CERTS_DIR)/ca.crt -CAkey $(CERTS_DIR)/ca.key -CAcreateserial -out $(CERTS_DIR)/client.crt -days 365

	@echo "Certificates generated in $(CERTS_DIR)/"

run-sink-plain: build
	@echo "Starting Sink (Plaintext)..."
	$(SINK_BINARY) \
		--bind :50051 \
		--log telemetry.log \
		--rate-limit 51200

run-sink: build
	@echo "Starting Sink..."
	$(SINK_BINARY) \
		--bind :50051 \
		--log telemetry.log \
		--cert $(CERTS_DIR)/server.crt \
		--key $(CERTS_DIR)/server.key \
		--ca $(CERTS_DIR)/ca.crt \
		--rate-limit 51200

run-sensor: build
	@echo "Starting Sensor..."
	$(SENSOR_BINARY) \
		--sink localhost:50051 \
		--rate 5 \
		--name sensor-mtls \
		--cert $(CERTS_DIR)/client.crt \
		--key $(CERTS_DIR)/client.key \
		--ca $(CERTS_DIR)/ca.crt

test:
	@if ! command -v gotestsum >/dev/null 2>&1; then \
		echo "gotestsum not found, installing..."; \
		go install gotest.tools/gotestsum@latest; \
	fi
	gotestsum --format testname -- -v ./...

lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.5.0; \
	fi
	GOFLAGS="-buildvcs=false" $(shell go env GOPATH)/bin/golangci-lint run

stress-test:
	@echo "Running k6 stress test..."
	k6 run k6/stress.js

stress-test-mtls:
	@echo "Running k6 stress test with mTLS..."
	k6 run k6/stress_mtls.js

docker-stress-test:
	@echo "Running k6 stress test in Docker..."
	docker-compose run --rm stress-test

docker-stress-test-mtls:
	@echo "Running k6 stress test with mTLS in Docker..."
	docker-compose run --rm stress-test-mtls

clean:
	rm -rf $(BIN_DIR) $(CERTS_DIR) *.log
	@echo "Cleaned up."

.PHONY: all build proto gen-certs run-sink run-sensor clean
