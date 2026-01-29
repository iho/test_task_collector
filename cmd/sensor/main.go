// Package main implements the Telemetry Sensor agent.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iho/test_task_collector/internal/sensor"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := sensor.LoadConfig()

	agent, err := sensor.NewAgent(cfg)
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	defer agent.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
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
	return nil
}
