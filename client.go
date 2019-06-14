package aconvert

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfk9w-go/flu"
)

// HostTemplate is a string template used to generate hosts.
// %d is substituted with a corresponding server number.
var HostTemplate = "https://s%d.aconvert.com"

// Client is an entity allowing access to aconvert.
type Client struct {
	httpClient *flu.Client
	queue      chan *request
	maxRetries int
	wg         *sync.WaitGroup
}

type request struct {
	body  flu.BodyWriter
	resp  *Response
	err   error
	retry int
	done  chan struct{}
}

// NewClient creates a new aconvert HTTP client and runs server discovery in the background.
func NewClient(httpClient *flu.Client, config *Config) *Client {
	if httpClient == nil {
		httpClient = flu.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute)
	}

	client := &Client{
		httpClient: httpClient,
		queue:      make(chan *request, config.QueueSize),
		maxRetries: config.MaxRetries,
		wg:         new(sync.WaitGroup),
	}

	if client.maxRetries < 1 {
		client.maxRetries = math.MaxInt32
	}

	go client.discover(config.TestFile, config.TestFormat)
	return client
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(media Media, opts Opts) (*Response, error) {
	req := &request{
		body: media.body(opts.values()),
		done: make(chan struct{}),
	}

	c.queue <- req
	for range req.done {
		// wait until the request is done
	}

	return req.resp, req.err
}

// ConvertURL accepts URL as argument.
func (c *Client) ConvertURL(url string, opts Opts) (*Response, error) {
	return c.Convert(URL{url}, opts)
}

// ConvertResource accepts flu.ReadResource as argument.
func (c *Client) ConvertResource(resource flu.ReadResource, opts Opts) (*Response, error) {
	return c.Convert(Resource{resource}, opts)
}

// Download saves the converted file to a resource.
func (c *Client) Download(r *Response, resource flu.WriteResource) error {
	return c.httpClient.NewRequest().
		GET().
		Resource(host(r.server) + "/convert/p3r68-cdx67/" + r.Filename).
		Send().
		CheckStatusCode(http.StatusOK).
		ReadResource(resource).
		Error
}

func (c *Client) Shutdown() {
	close(c.queue)
	c.wg.Wait()
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
	}
}

func (c *Client) trySpawnWorker(hostID int, resource flu.FileSystemResource, opts Opts, onComplete func(bool)) {
	host := host(hostID)
	body := Resource{resource}.body(opts.values())
	for j := 0; j < c.maxRetries; j++ {
		_, err := c.convert(host, body)
		if err == nil {
			go c.runWorker(host)
			onComplete(true)
			return
		}

		if j < 2 {
			time.Sleep(time.Duration(2^j) * time.Second)
		}
	}

	onComplete(false)
}

func (c *Client) runWorker(host string) {
	c.wg.Add(1)
	defer c.wg.Done()
	for req := range c.queue {
		resp, err := c.convert(host, req.body)
		if err == nil {
			err = resp.init()
		}

		if err != nil && req.retry < c.maxRetries {
			req.retry++
			c.queue <- req
		} else {
			req.resp = resp
			req.err = err
			close(req.done)
		}
	}
}

func (c *Client) convert(host string, body flu.BodyWriter) (*Response, error) {
	resp := new(Response)
	err := c.httpClient.NewRequest().
		POST().
		Resource(host + "/convert/convert-batch.php").
		Body(body).
		Buffer().
		Send().
		CheckStatusCode(http.StatusOK).
		ReadBodyFunc(flu.JSON(resp).Read).
		Error

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func host(number int) string {
	return fmt.Sprintf(HostTemplate, number)
}
