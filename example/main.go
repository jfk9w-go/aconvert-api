package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/flu/httpf"
	"github.com/jfk9w-go/flu/me3x"
)

//noinspection GoUnhandledErrorResult
func main() {
	var (
		webm = flu.File("testdata/test1.webm")
		mp4  = flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	)

	if err := os.RemoveAll(mp4.Path()); err != nil {
		log.Panicf("remove %s: %v", mp4, err)
	}

	defer os.RemoveAll(mp4.Path())
	config := &Config{
		Probe: &Probe{
			File:   webm,
			Format: "mp4",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := NewClient(ctx, me3x.DummyRegistry{}, config)
	resp, err := client.Convert(ctx, webm, make(Opts).TargetFormat("mp4"))
	if err != nil {
		log.Panicf("convert %s: %v", webm, err)
	} else {
		log.Printf("response state: %s\n", resp.State)
	}

	var size int64
	if err := httpf.HEAD(resp.URL()).
		Exchange(ctx, client).
		HandleFunc(func(resp *http.Response) (err error) {
			header := resp.Header.Get("Content-Length")
			size, err = strconv.ParseInt(header, 10, 64)
			return
		}).
		Error(); err != nil {
		log.Panicf("get size from head request: %v", err)
	} else {
		log.Printf("response content length: %d b", size)
	}

	if err := httpf.GET(resp.URL()).
		Exchange(ctx, client).
		DecodeBodyTo(mp4).
		Error(); err != nil {
		log.Panicf("get mp4 file: %v", err)
	}

	stat, err := os.Stat(mp4.Path())
	if err != nil {
		log.Panicf("stat mp4 file: %v", err)
	}

	size = stat.Size()
	log.Printf("downloaded file size: %d b", size)
}
