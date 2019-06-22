package aconvert_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
)

// ExampleClient provides a Client usage example.
func ExampleClient() {
	// Create a resource which will contain the converted file.
	resource := flu.NewFileSystemResource(filepath.Join(os.TempDir(), "test.mp4"))

	// Cleanup.
	err := os.RemoveAll(resource.Path())
	if err != nil {
		fmt.Println(err)
		return
	}

	// First we create a new Client.
	// Pass test file path and format used for discovery in Config.
	c := NewClient(nil, &Config{
		TestFile:   "testdata/test.webm",
		TestFormat: "mp4",
	})

	// Convert the test file.
	resp, err := c.ConvertResource(
		flu.NewFileSystemResource("testdata/test.webm"),
		NewOpts().TargetFormat("mp4"))
	//Param("code", "81000"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("State: %s\n", resp.State)

	// Download the converted file.
	err = c.Download(resp, resource)
	if err != nil {
		fmt.Println(err)
		return
	}

	// No way to introspect the file so this will have to do.
	_, err = os.Stat(resource.Path())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("File exists")

	c.Shutdown()
	_ = os.RemoveAll(resource.Path())

	// Output:
	// State: SUCCESS
	// File exists
}
