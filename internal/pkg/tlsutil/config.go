// Package tlsutil provides helper functions for loading TLS configurations.
package tlsutil

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds the paths required to load TLS certificates.
type Config struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// LoadTLSConfig loads the TLS configuration from the provided file paths.
func LoadTLSConfig(cfg Config, clientAuthType tls.ClientAuthType) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   clientAuthType,
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
		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
