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
	webm := flu.File("example/testdata/test.webm")
	mp4 := flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	err := os.RemoveAll(mp4.Path())
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(mp4.Path())
	c := NewClient(nil, Config{Servers: []int{13}})
	resp, err := c.Convert(webm, make(Opts).TargetFormat("mp4"))
	if err != nil {
		panic(err)
	}
	log.Printf("State: %s\n", resp.State)
	handler := new(sizeResponseHandler)
	err = c.HEAD(resp.URL()).Execute().HandleResponse(handler).Error
	if err != nil {
		panic(err)
	}
	log.Printf("Content-Length: %d b", handler.size)
	err = c.GET(resp.URL()).Execute().ReadBodyTo(mp4).Error
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(mp4.Path())
	if err != nil {
		panic(err)
	}
	size := stat.Size()
	log.Printf("Downloaded file size: %d b", size)
}

type sizeResponseHandler struct {
	size int64
}

func (h *sizeResponseHandler) Handle(resp *http.Response) (err error) {
	header := resp.Header.Get("Content-Length")
	h.size, err = strconv.ParseInt(header, 10, 64)
	return
}
