package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sirupsen/logrus"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
)

//noinspection GoUnhandledErrorResult
func main() {
	webm := flu.File("example/testdata/test1.webm")
	mp4 := flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	err := os.RemoveAll(mp4.Path())
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(mp4.Path())
	c := NewClient(nil, nil, nil)
	resp, err := c.Convert(context.Background(), webm, make(Opts).TargetFormat("mp4"))
	if err != nil {
		panic(err)
	}
	logrus.Infof("State: %s\n", resp.State)
	handler := new(sizeResponseHandler)
	err = c.HEAD(resp.URL()).Execute().HandleResponse(handler).Error
	if err != nil {
		panic(err)
	}
	logrus.Infof("Content-Length: %d b", handler.size)
	err = c.GET(resp.URL()).Execute().DecodeBodyTo(mp4).Error
	if err != nil {
		panic(err)
	}
	stat, err := os.Stat(mp4.Path())
	if err != nil {
		panic(err)
	}
	size := stat.Size()
	logrus.Infof("Downloaded file size: %d b", size)
}

type sizeResponseHandler struct {
	size int64
}

func (h *sizeResponseHandler) Handle(resp *http.Response) (err error) {
	header := resp.Header.Get("Content-Length")
	h.size, err = strconv.ParseInt(header, 10, 64)
	return
}
