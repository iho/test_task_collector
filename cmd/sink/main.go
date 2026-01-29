// Package main implements the Telemetry Sink service.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/iho/test_task_collector/internal/sink"
	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	cfg := sink.LoadConfig()

	writer, err := sink.NewBufferedWriter(cfg.LogFile, cfg.EncryptKey, cfg.BufferSize, cfg.FlushInterval)
	if err != nil {
		log.Fatalf("Failed to create writer: %v", err)
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
		tlsConfig, err := loadTLSConfig(cfg)
		if err != nil {
			log.Fatalf("Failed to load TLS config: %v", err)
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.Creds(creds))
		log.Println("TLS enabled")
	}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterTelemetryServiceServer(grpcServer, srv)

	lis, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		grpcServer.GracefulStop()
		if err := writer.Flush(); err != nil {
			log.Printf("Failed to flush writer: %v", err)
		}
	}()

	log.Printf("Sink server listening on %s", cfg.BindAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func loadTLSConfig(cfg *sink.Config) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.NoClientCert,
	}

	if cfg.CAFile != "" {
		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}
		tlsConfig.ClientCAs = caCertPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}
