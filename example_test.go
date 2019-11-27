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
	// Create a res which will contain the converted file.
	resource := flu.File(filepath.Join(os.TempDir(), "test.mp4"))

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
		flu.File("testdata/test.webm"),
		NewOpts().TargetFormat("mp4"))

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

	stat, err := os.Stat(resource.Path())
	if err != nil {
		fmt.Println(err)
		return
	}

	// source file size is 500 KB
	// this should not be lower 350 KB
	size := stat.Size()
	if size < 350000 {
		fmt.Println("Invalid file size: ", size)
		return
	}

	c.Shutdown()
	_ = os.RemoveAll(resource.Path())

	// Output:
	// State: SUCCESS
}
