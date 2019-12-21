package aconvert

import (
	"github.com/jfk9w-go/flu"
)

type Probe struct {
	File   flu.File
	Format string
}

// Config contains configuration parameters for a Client.
type Config struct {
	// Servers allows user to specify which aconvert.com servers he would like to use.
	// If nil, DefaultServers will be used.
	Servers []int

	// MaxRetries is the maximum number of times a request will be retried.
	MaxRetries int

	// Probe contains file and target format configuration for discovery purposes.
	// If nil, all servers from Servers will be used.
	Probe *Probe
}
