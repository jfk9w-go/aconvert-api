package aconvert

import (
	"errors"
	"os"

	"github.com/jfk9w-go/flu"
)

// Config contains configuration parameters for a Client.
type Config struct {
	// Servers allows user to specify which aconvert.com servers he would like to use.
	Servers []int

	// MaxRetries is the maximum number of times a request will be retried.
	MaxRetries int

	// TestFile is a path to the file which will be used for discovery.
	TestFile flu.File

	// TestFormat is a format which test file will be converted to.
	TestFormat string
}

func (c *Config) validate() error {
	if c.TestFile == "" {
		return errors.New("test file path is not set")
	}
	if c.TestFormat == "" {
		return errors.New("test format is not set")
	}
	stat, err := os.Stat(c.TestFile.Path())
	if err != nil || stat.IsDir() {
		return errors.New("test file does not exist")
	}
	if c.Servers == nil {
		c.Servers = DefaultServers
	}
	return nil
}
