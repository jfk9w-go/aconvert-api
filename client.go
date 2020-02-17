package aconvert

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfk9w-go/flu"
)

// BaseURITemplate is a string template used to generate base URIs.
// %v is substituted with a corresponding server number.
var BaseURITemplate = "https://s%v.aconvert.com"

// Client is an entity allowing access to aconvert.
type Client struct {
	*flu.Client
	Servers    []int
	MaxRetries int
	Probe      *Probe
	servers    chan server
	work       sync.WaitGroup
}

func (c *Client) Start() *Client {
	if c.Servers == nil {
		c.Servers = DefaultServers
	}

	if c.Client == nil {
		c.Client = flu.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute).
			AcceptResponseCodes(http.StatusOK)
	}

	c.servers = make(chan server, len(c.Servers))
	if c.Probe == nil {
		log.Printf("Using %d configured servers", len(c.Servers))
		for _, id := range c.Servers {
			c.servers <- server{c.Client, baseURI(id)}
		}
	} else {
		go c.discover(context.TODO(), c.Probe, c.Servers)
	}

	return c
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(ctx context.Context, in flu.Readable, opts Opts) (resp *Response, err error) {
	body := opts.body(in)
	for i := 0; i <= c.MaxRetries; i++ {
		var server server
		select {
		case <-ctx.Done():
			return
		case server = <-c.servers:
			resp, err = server.convert(ctx, body)
			c.servers <- server
			if err != nil {
				continue
			} else {
				return
			}
		}
	}
	return
}

func (c *Client) discover(ctx context.Context, probe *Probe, servers []int) {
	discovered := new(int32)
	workers := new(sync.WaitGroup)
	workers.Add(len(servers))
	body := make(Opts).TargetFormat(probe.Format).body(probe.File)
	for _, id := range servers {
		go func(id int) {
			server := server{c.Client, baseURI(id)}
			if err := server.test(ctx, body, c.MaxRetries); err == nil {
				atomic.AddInt32(discovered, 1)
				c.servers <- server
			} else {
				workers.Done()
			}
		}(id)
	}
	workers.Wait()
	if *discovered == 0 {
		panic("no hosts discovered")
	}
	log.Printf("Discovered %d aconvert servers", *discovered)
}

func baseURI(id interface{}) string {
	return fmt.Sprintf(BaseURITemplate, id)
}
