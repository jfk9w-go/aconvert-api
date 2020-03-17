package aconvert

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/jfk9w-go/flu"
	fluhttp "github.com/jfk9w-go/flu/http"
)

// BaseURITemplate is a string template used to generate base URIs.
// %v is substituted with a corresponding server number.
var BaseURITemplate = "https://s%v.aconvert.com"

// Probe denotes a file to be used for discovering servers.
type Probe struct {
	File   flu.File
	Format string
}

// Client is an entity allowing access to aconvert.
type Client struct {
	fluhttp.Client
	Servers    []int
	MaxRetries int
	Probe      *Probe
	servers    chan concreteClient
	work       sync.WaitGroup
}

func (c *Client) Init() *Client {
	if c.Servers == nil {
		c.Servers = DefaultServers
	}

	if c.Client.Client == nil {
		c.Client = fluhttp.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute).
			AcceptStatus(http.StatusOK)
	}

	c.servers = make(chan concreteClient, len(c.Servers))
	if c.Probe == nil {
		log.Printf("Using %d configured servers", len(c.Servers))
		for _, id := range c.Servers {
			c.servers <- newConcreteClient(baseURI(id))
		}
	} else {
		go c.discover(context.TODO(), c.Probe, c.Servers)
	}

	return c
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(ctx context.Context, in flu.Input, opts Opts) (resp *Response, err error) {
	req, err := opts.makeRequest(c.Client, in)
	if err != nil {
		err = errors.Wrap(err, "make request")
		return
	}
	for i := 0; i <= c.MaxRetries; i++ {
		var server concreteClient
		select {
		case <-ctx.Done():
			return
		case server = <-c.servers:
			resp, err = server.convert(ctx, req)
			c.servers <- server
			if err != nil {
				continue
			} else {
				err = errors.Wrap(err, "convert")
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
	req, err := make(Opts).
		TargetFormat(probe.Format).
		makeRequest(c.Client, probe.File)
	if err != nil {
		log.Fatalf("Failed to make aconvert probe request: %s", err)
	}
	for _, id := range servers {
		go func(id int) {
			server := newConcreteClient(baseURI(id))
			if err := server.test(ctx, req, c.MaxRetries); err == nil {
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
