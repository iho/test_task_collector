// Package sink implements the Telemetry Sink service.
package sink

import (
	"flag"
	"time"
)

// Config holds the configuration for the Telemetry Sink.
type Config struct {
	BindAddr      string
	LogFile       string
	BufferSize    int
	FlushInterval time.Duration
	RateLimit     int
	CertFile      string
	KeyFile       string
	CAFile        string
	EncryptKey    string
}

// LoadConfig loads configuration from command-line flags.
func LoadConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.BindAddr, "bind", ":50051", "Address to bind the gRPC server")
	flag.StringVar(&cfg.LogFile, "log", "telemetry.log", "Path to the output log file")
	flag.IntVar(&cfg.BufferSize, "buffer", 4096, "Buffer size in bytes")
	flag.DurationVar(&cfg.FlushInterval, "flush", 100*time.Millisecond, "Buffer flush interval")
	flag.IntVar(&cfg.RateLimit, "rate-limit", 0, "Max input flow rate in bytes/sec (0 for unlimited)")
	flag.StringVar(&cfg.CertFile, "cert", "", "Path to TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "key", "", "Path to TLS key file")
	flag.StringVar(&cfg.CAFile, "ca", "", "Path to CA certificate file for mTLS")
	flag.StringVar(&cfg.EncryptKey, "encrypt-key", "", "Hex-encoded 32-byte key for log encryption")

	flag.Parse()
	return cfg
}
