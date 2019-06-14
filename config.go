package aconvert

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
