package main

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
)

// Client usage example.
func main() {
	// Create a in which will contain the converted file.
	file := flu.File(filepath.Join(os.TempDir(), "test.mp4"))

	// Cleanup.
	err := os.RemoveAll(file.Path())
	if err != nil {
		fmt.Println(err)
		return
	}

	config := &Config{
		Servers:    []int{7},
		TestFile:   "example/testdata/test.webm",
		TestFormat: "mp4",
	}

	// First we create a new Client.
	// Pass test file path and format used for discovery in Config.
	c := NewClient(nil, config)

	// Convert the test file.
	resp, err := c.Convert(
		config.TestFile,
		NewOpts().TargetFormat("mp4"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("State: %s\n", resp.State)

	// Download the converted file.
	err = c.Download(resp.URL(), file)
	if err != nil {
		fmt.Println(err)
		return
	}

	stat, err := os.Stat(file.Path())
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

	_ = os.RemoveAll(file.Path())

	// Output:
	// State: SUCCESS
}
