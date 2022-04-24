// Package main demonstrates how to convert a webm video file to mp4 format
// with github.com/jfk9w-go/aconvert-api.
package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jfk9w-go/aconvert-api"
	. "github.com/jfk9w-go/aconvert-api"
	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/flu/apfel"
	"github.com/jfk9w-go/flu/httpf"
	"github.com/jfk9w-go/flu/logf"
)

//noinspection GoUnhandledErrorResult
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		webm = flu.File("testdata/test1.webm")
		mp4  = flu.File(filepath.Join(os.TempDir(), "test.mp4"))
	)

	err := os.RemoveAll(mp4.String())
	logf.Resultf(ctx, logf.Trace, logf.Panic, "remove %s: %v", mp4, err)

	defer os.RemoveAll(mp4.String())

	var client aconvert.Client[aconvert.Context]
	if err := client.Standalone(ctx, apfel.Default[aconvert.Config]()); err != nil {
		logf.Panicf(ctx, "init: %v", err)
	}

	resp, err := client.Convert(ctx, webm, make(Options).TargetFormat("mp4"))
	logf.Resultf(ctx, logf.Info, logf.Panic, "convert %s => %s (%v)", webm, resp, err)

	var size int64
	err = httpf.HEAD(resp.URL()).
		Exchange(ctx, client).
		HandleFunc(func(resp *http.Response) (err error) {
			header := resp.Header.Get("Content-Length")
			size, err = strconv.ParseInt(header, 10, 64)
			return
		}).
		Error()
	logf.Resultf(ctx, logf.Info, logf.Panic, "get size => %d (%v)", size, err)

	err = httpf.GET(resp.URL()).
		Exchange(ctx, client).
		CopyBody(mp4).
		Error()
	logf.Resultf(ctx, logf.Info, logf.Panic, "download mp4 file: %v", err)

	stat, err := os.Stat(mp4.String())
	logf.Resultf(ctx, logf.Debug, logf.Panic, "stat mp4 file: %v", err)

	size = stat.Size()
	logf.Infof(ctx, "downloaded file size: %d b", size)
}
