package fsmonitor

import "time"

// Config defines the configuration options for the Monitor.
type Config struct {
	FlushDelay time.Duration // Duration to delay the flush operation
}

// DefaultConfig returns the default configuration for the Monitor.
func DefaultConfig() *Config {
	return &Config{
		FlushDelay: 3 * time.Second, // Set a default flush delay
	}
}
