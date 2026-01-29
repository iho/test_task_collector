package tests

import (
	"bufio"
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iho/test_task_collector/internal/sensor"
	"github.com/iho/test_task_collector/internal/sink"
	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc"
)

func TestIntegration(t *testing.T) {
	// 1. Setup Sink
	logFile := "integration.log"
	defer func() { _ = os.Remove(logFile) }()

	writer, err := sink.NewBufferedWriter(logFile, "", 1024, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer func() { _ = writer.Close() }()

	limiter := sink.NewRateLimiter(0) // Unlimited
	srv := sink.NewServer(writer, limiter)

	lis, err := net.Listen("tcp", ":0") // Random port
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer func() { _ = lis.Close() }()

	grpcServer := grpc.NewServer() // No TLS for basic integration test
	pb.RegisterTelemetryServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	addr := lis.Addr().String()
	t.Logf("Sink listening on %s", addr)

	// 2. Setup Sensor
	sensorCfg := &sensor.Config{
		SinkAddr: addr,
		Rate:     20, // 20 msg/sec
		Name:     "integration-sensor",
	}

	// We can't use sensor.NewAgent directly easily because it uses grpc.Dial inside which blocks or we want to inject insecure creds without modifying config to be complex.
	// Actually sensor.NewAgent handles insecure if certs are empty.
	agent, err := sensor.NewAgent(sensorCfg)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1100*time.Millisecond)
	defer cancel()

	// 3. Run
	startTime := time.Now()
	agent.Start(ctx)
	duration := time.Since(startTime)

	// 4. Verification
	// Flush writer
	_ = writer.Close() // Forces flush

	// Read log file
	file, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "integration-sensor") {
			count++
		}
	}

	t.Logf("Sent for %v, received %d messages", duration, count)

	// 20 msg/sec * 1.1s ~= 22 messages. Expect at least 15.
	if count < 15 {
		t.Errorf("Expected at least 15 messages, got %d", count)
	}
}
