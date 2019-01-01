package aconvert

import (
	"log"
	"testing"

	"github.com/jfk9w-go/flu"
)

func TestDiscovery(t *testing.T) {
	var (
		api = NewApi(nil, Config{
			TestFile:   "testdata/test.webm",
			TestFormat: "mp4",
		})

		r, err = api.Convert(
			flu.NewFileSystemResource("testdata/test.webm"),
			NewOpts().TargetFormat("mp4"))
	)

	if err != nil {
		t.Fatal(err)
	}

	log.Printf("Received %+v", r)

	var output = flu.NewFileSystemResource("output.mp4")
	err = api.Download(r, output)
	if err != nil {
		t.Fatal(err)
	}

	log.Printf("Downloaded file")
}
