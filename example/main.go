package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
)

//noinspection GoUnhandledErrorResult
func main() {
	file := flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	err := os.RemoveAll(file.Path())
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(file.Path())
	c := NewClient(nil, Config{})
	resp, err := c.Convert(flu.File("example/testdata/test.webm"), make(Opts).TargetFormat("mp4"))
	if err != nil {
		panic(err)
	}
	log.Printf("State: %s\n", resp.State)
	err = c.Download(resp.URL(), file)
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(file.Path())
	if err != nil {
		panic(err)
	}
	size := stat.Size()
	log.Printf("Converted file size: %d Kb", size>>10)
	_ = os.RemoveAll(file.Path())
	err = c.NewRequest().
		Resource(resp.URL()).
		HEAD().
		Send().
		HandleResponse(responseHandler{}).
		Error
	if err != nil {
		panic(err)
	}
}

type responseHandler struct{}

func (responseHandler) Handle(resp *http.Response) error {
	header := resp.Header.Get("Content-Length")
	length, err := strconv.Atoi(header)
	if err != nil {
		return err
	}
	log.Printf("Content-Length: %d Kb", length>>10)
	return nil
}
