package aconvert

import (
	"errors"
	"os"
)

// Config contains configuration parameters for a Client.
type Config struct {
	// QueueSize is the maximum request queue size.
	QueueSize int

	// MaxRetries is the maximum number of times a request will be retried.
	MaxRetries int

	// TestFile is a path to the file which will be used for discovery.
	TestFile string

	// TestFormat is a format which test file will be converted to.
	TestFormat string
}

func (c *Config) validate() error {
	if c.QueueSize < 0 {
		return errors.New("queue size must be non-negative")
	}

	if c.TestFile == "" {
		return errors.New("test file path is not set")
	}

	if c.TestFormat == "" {
		return errors.New("test format is not set")
	}

	stat, err := os.Stat(c.TestFile)
	if err != nil || stat.IsDir() {
		return errors.New("test file does not exist")
	}

	return nil
}
