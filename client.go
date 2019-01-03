package aconvert

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/lego/pool"
)

// HostTemplate is a string template used to generate hosts.
// %d is substituted with a corresponding server number.
var HostTemplate = "https://s%d.aconvert.com"

// Client is an entity allowing access to aconvert.
type Client struct {
	http *flu.Client
	pool pool.Pool
}

// NewClient creates a new aconvert HTTP client and runs server discovery in the background.
func NewClient(http *flu.Client, config *Config) *Client {
	if http == nil {
		http = flu.NewClient(nil).
			ResponseHeaderTimeout(120 * time.Second)
	}

	var client = &Client{
		http: http,
		pool: pool.New(),
	}

	go client.discover(config.TestFile, config.TestFormat)
	return client
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(entity interface{}, opts Opts) (*Response, error) {
	ptr := &taskPtr{entity: entity, opts: opts}
	err := c.pool.Execute(ptr)
	return ptr.resp, err
}

// Download saves the converted file to a resource.
func (c *Client) Download(r *Response, resource flu.WriteResource) error {
	return c.http.NewRequest().
		Endpoint(host(r.server) + "/convert/p3r68-cdx67/" + r.Filename).
		Get().
		StatusCodes(http.StatusOK).
		ReadResource(resource).
		Error
}

// Close shuts down the worker pool.
func (c *Client) Close() {
	c.pool.Close()
}

func (c *Client) discover(file string, format string) {
	resource := flu.NewFileSystemResource(file)
	hostsDiscovered := new(int32)
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(30)
	for i := 0; i < 30; i++ {
		go c.trySpawnWorker(i, resource, NewOpts().TargetFormat(format), func(discovered bool) {
			if discovered {
				atomic.AddInt32(hostsDiscovered, 1)
			}

			waitGroup.Done()
		})
	}

	waitGroup.Wait()
	if *hostsDiscovered == 0 {
		panic("no hosts discovered")
	} else {
		log.Printf("Discovered %d hosts\n", *hostsDiscovered)
	}
}

func (c *Client) trySpawnWorker(hostID int, resource *flu.FileSystemResource, opts Opts, onComplete func(bool)) {
	worker := &worker{c.http, host(hostID)}
	for j := 0; j < 3; j++ {
		_, err := worker.execute(resource, opts)
		if err == nil {
			c.pool.Spawn(worker)
			onComplete(true)
			return
		}

		if j < 2 {
			time.Sleep(time.Duration(2^j) * time.Second)
		}
	}

	onComplete(false)
}

type worker struct {
	http *flu.Client
	host string
}

func (w *worker) Execute(task *pool.Task) {
	ptr := task.Ptr.(*taskPtr)
	resp, err := w.execute(ptr.entity, ptr.opts)
	if err != nil && ptr.retry < 3 {
		ptr.retry += 1
		task.Retry()
	} else {
		ptr.resp = resp
		task.Complete(err)
	}
}

func (w *worker) execute(entity interface{}, opts Opts) (*Response, error) {
	resp := new(Response)
	err := w.http.NewRequest().
		Endpoint(w.host + "/convert/convert-batch.php").
		Body(opts.body(entity)).Sync().
		Post().
		StatusCodes(http.StatusOK).
		ReadBodyFunc(flu.JSON(resp).Read).
		Error

	if err != nil {
		return nil, err
	}

	return resp, resp.init()
}

type taskPtr struct {
	entity interface{}
	opts   Opts
	resp   *Response
	retry  int
}

func host(number int) string {
	return fmt.Sprintf(HostTemplate, number)
}
