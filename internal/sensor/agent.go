// Package sensor implements the Telemetry Sensor agent.
package sensor

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/iho/test_task_collector/internal/pkg/tlsutil"
	pb "github.com/iho/test_task_collector/proto/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Agent is a sensor that generates and sends metrics to a sink.
type Agent struct {
	cfg    *Config
	client pb.TelemetryServiceClient
	conn   *grpc.ClientConn
}

// NewAgent creates a new Sensor Agent.
func NewAgent(cfg *Config) (*Agent, error) {
	opts := []grpc.DialOption{}

	if cfg.CertFile != "" && cfg.KeyFile != "" {
		tlsConfig, err := tlsutil.LoadTLSConfig(tlsutil.Config{
			CertFile: cfg.CertFile,
			KeyFile:  cfg.KeyFile,
			CAFile:   cfg.CAFile,
		}, tls.NoClientCert)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.SinkAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("did not connect: %v", err)
	}

	client := pb.NewTelemetryServiceClient(conn)
	return &Agent{
		cfg:    cfg,
		client: client,
		conn:   conn,
	}, nil
}

// Start begins the metric generation and transmission loop.
func (a *Agent) Start(ctx context.Context) {
	interval := time.Second / time.Duration(a.cfg.Rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.sendMetric(ctx)
		}
	}
}

func (a *Agent) sendMetric(ctx context.Context) {
	val := rand.Int63n(100)

	metric := &pb.Metric{
		Name:      a.cfg.Name,
		Value:     val,
		Timestamp: timestamppb.Now(),
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	_, err := a.client.Publish(reqCtx, metric)
	if err != nil {
		log.Printf("Error sending metric: %v", err)
	}
}

// Close closes the gRPC connection.
func (a *Agent) Close() {
	if err := a.conn.Close(); err != nil {
		log.Printf("Failed to close connection: %v", err)
	}
}
