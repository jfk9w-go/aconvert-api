package aconvert

import (
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfk9w-go/flu"
)

// HostTemplate is a string template used to generate hosts.
// %d is substituted with a corresponding server number.
var HostTemplate = "https://s%v.aconvert.com"

// Client is an entity allowing access to aconvert.
type Client struct {
	http       *flu.Client
	queue      chan *request
	maxRetries int
	workers    sync.WaitGroup
}

type request struct {
	body  flu.BodyEncoderTo
	resp  *Response
	err   error
	retry int
	work  sync.WaitGroup
}

func newRequest(body flu.BodyEncoderTo) *request {
	req := &request{body: body}
	req.work.Add(1)
	return req
}

// NewClient creates a new aconvert HTTP client and runs server discovery in the background.
func NewClient(http *flu.Client, config *Config) *Client {
	if err := config.validate(); err != nil {
		panic(err)
	}

	if http == nil {
		http = flu.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute).
			AcceptResponseCodes(200)
	}

	client := &Client{
		http:       http,
		queue:      make(chan *request, config.QueueSize),
		maxRetries: config.MaxRetries,
	}

	go client.discover(config.TestFile, config.TestFormat)
	return client
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(media Media, opts Options) (*Response, error) {
	req := newRequest(media.body(opts.values()))
	c.queue <- req
	req.work.Wait()
	return req.resp, req.err
}

// ConvertURL accepts URL as argument.
func (c *Client) ConvertURL(url string, opts Options) (*Response, error) {
	return c.Convert(URL(url), opts)
}

// ConvertResource accepts flu.ReadResource as argument.
func (c *Client) ConvertResource(resource flu.ResourceReader, opts Options) (*Response, error) {
	return c.Convert(Resource{resource}, opts)
}

// Download saves the converted file to a resource.
func (c *Client) Download(r *Response, resource flu.ResourceWriter) error {
	return c.http.NewRequest().
		GET().
		Resource(r.host + "/convert/p3r68-cdx67/" + r.Filename).
		Send().
		ReadResource(resource).
		Error
}

func (c *Client) Shutdown() {
	close(c.queue)
	c.workers.Wait()
}

func (c *Client) discover(filepath string, format string) {
	file := flu.File(filepath)
	discovered := new(int32)
	workers := new(sync.WaitGroup)
	workers.Add(30)
	for i := 0; i < 30; i++ {
		go func(hostID int) {
			if c.trySpawnWorker(hostID, file, NewOpts().TargetFormat(format)) {
				atomic.AddInt32(discovered, 1)
			}

			workers.Done()
		}(i)
	}

	workers.Wait()
	if *discovered == 0 {
		panic("no hosts discovered")
	}

	log.Printf("Discovered %d aconvert workers", *discovered)
}

func (c *Client) trySpawnWorker(hostID int, file flu.File, options Options) bool {
	host := host(hostID)
	body := Resource{file}.body(options.values())
	for j := 0; j <= c.maxRetries; j++ {
		_, err := c.convert(host, body)
		if err == nil {
			go c.runWorker(host)
			return true
		}

		if j < 2 {
			time.Sleep(time.Duration(math.Pow(2, float64(j))) * time.Second)
		}
	}

	return false
}

func (c *Client) runWorker(host string) {
	c.workers.Add(1)
	defer c.workers.Done()
	for req := range c.queue {
		resp, err := c.convert(host, req.body)
		if err != nil && req.retry <= c.maxRetries {
			req.retry++
			c.queue <- req
		} else {
			req.resp = resp
			req.err = err
			req.work.Done()
		}
	}
}

func (c *Client) convert(host string, body flu.BodyEncoderTo) (*Response, error) {
	resp := new(Response)
	err := c.http.NewRequest().
		POST().
		Resource(host + "/convert/convert-batch.php").
		Body(body).
		Buffer().
		Send().
		Decode(resp). // DecodeBody checks the Content-Type header which doesn't match in this case
		Error

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func host(server interface{}) string {
	return fmt.Sprintf(HostTemplate, server)
}
