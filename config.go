package aconvert

// Config contains configuration parameters for a Client.
type Config struct {

	// TestFile is a path to the file which will be used for discovery.
	TestFile string `json:"test_file"`

	// TestFormat is a format which test file will be converted to.
	TestFormat string `json:"test_format"`
}
