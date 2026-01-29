// Package main implements the Telemetry Sink service.
package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/iho/test_task_collector/internal/pkg/tlsutil"
	"github.com/iho/test_task_collector/internal/sink"
	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := sink.LoadConfig()

	writer, err := sink.NewBufferedWriter(cfg.LogFile, cfg.EncryptKey, cfg.BufferSize, cfg.FlushInterval)
	if err != nil {
		return fmt.Errorf("create writer: %w", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			log.Printf("Failed to close writer: %v", err)
		}
	}()

	limiter := sink.NewRateLimiter(cfg.RateLimit)
	srv := sink.NewServer(writer, limiter)

	var opts []grpc.ServerOption
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsConfig, err := tlsutil.LoadTLSConfig(tlsutil.Config{
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
			CAFile:   cfg.CAFile,
		}, tls.RequireAndVerifyClientCert)
		if err != nil {
			return fmt.Errorf("load tls config: %w", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.Creds(creds))
		log.Println("TLS enabled")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterTelemetryServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	// Graceful shutdown channel
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
		if err := writer.Flush(); err != nil {
			log.Printf("Failed to flush writer: %v", err)
		}
	}()

	log.Printf("Sink server listening on %s", cfg.BindAddr)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
