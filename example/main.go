package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/flu/metrics"
	"github.com/sirupsen/logrus"
)

//noinspection GoUnhandledErrorResult
func main() {
	logrus.SetLevel(logrus.TraceLevel)
	webm := flu.File("testdata/test1.webm")
	mp4 := flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	err := os.RemoveAll(mp4.Path())
	if err != nil {
		panic(err)
	}

	defer os.RemoveAll(mp4.Path())
	config := &Config{
		Probe: &Probe{
			File:   webm,
			Format: "mp4",
		},
	}

	ctx := context.Background()
	client := NewClient(ctx, metrics.DummyRegistry{}, config)
	resp, err := client.Convert(context.Background(), webm, make(Opts).TargetFormat("mp4"))
	if err != nil {
		panic(err)
	}
	logrus.Infof("State: %s\n", resp.State)
	handler := new(sizeResponseHandler)
	err = client.HEAD(resp.URL()).Execute().HandleResponse(handler).Error
	if err != nil {
		panic(err)
	}

	logrus.Infof("Content-Length: %d b", handler.size)
	err = client.GET(resp.URL()).Execute().DecodeBodyTo(mp4).Error
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
