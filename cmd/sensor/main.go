// Package main implements the Telemetry Sensor agent.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iho/test_task_collector/internal/sensor"
)

func main() {
	cfg := sensor.LoadConfig()

	agent, err := sensor.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down sensor...")
		cancel()
	}()

	log.Printf(
		"Starting sensor '%s' sending to %s at %d msg/sec",
		cfg.Name,
		cfg.SinkAddr,
		cfg.Rate,
	)

	agent.Start(ctx)
}
