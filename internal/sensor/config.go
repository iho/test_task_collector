package sensor

import (
	"flag"
)

// Config holds the configuration for the Sensor Agent.
type Config struct {
	SinkAddr string
	Rate     int
	Name     string

	CertFile string
	KeyFile  string
	CAFile   string
}

// LoadConfig loads configuration from command-line flags.
func LoadConfig() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.SinkAddr, "sink", "localhost:50051", "Address of the telemetry sink")
	flag.IntVar(&cfg.Rate, "rate", 1, "Number of messages per second to send")
	flag.StringVar(&cfg.Name, "name", "sensor-1", "Name of the sensor")
	flag.StringVar(&cfg.CertFile, "cert", "", "Path to TLS certificate file")
	flag.StringVar(&cfg.KeyFile, "key", "", "Path to TLS key file")
	flag.StringVar(&cfg.CAFile, "ca", "", "Path to CA certificate file for mTLS")

	flag.Parse()
	return cfg
}
