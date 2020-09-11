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
	fluhttp "github.com/jfk9w-go/flu/http"
	"github.com/pkg/errors"
)

// BaseURITemplate is a string template used to generate base URIs.
// %v is substituted with a corresponding server number.
var BaseURITemplate = "https://s%v.aconvert.com"

var MaxRetries = 3

// Probe denotes a file to be used for discovering servers.
type Probe struct {
	File   flu.File
	Format string
}

// Client is an entity allowing access to aconvert.
type Client struct {
	*fluhttp.Client
	servers chan concreteClient
}

func NewClient(client *fluhttp.Client, servers []int, probe *Probe) *Client {
	if servers == nil {
		servers = DefaultServers
	}

	if client == nil {
		client = fluhttp.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute).
			AcceptStatus(http.StatusOK)
	}

	c := &Client{
		Client:  client,
		servers: make(chan concreteClient, len(servers)),
	}

	if probe == nil {
		log.Printf("aconvert: using %d configured servers", len(servers))
		for _, id := range servers {
			c.servers <- newConcreteClient(baseURI(id))
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		go func(c *Client) {
			c.discover(ctx, probe, servers)
			cancel()
		}(c)
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
	for i := 0; i <= MaxRetries; i++ {
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
	req, err := make(Opts).TargetFormat(probe.Format).makeRequest(c.Client, probe.File)
	if err != nil {
		log.Fatalf("aconvert: probe request failed: %s", err)
	}

	work := new(sync.WaitGroup)
	work.Add(len(servers))
	for _, id := range servers {
		go func(id int) {
			defer work.Done()
			server := newConcreteClient(baseURI(id))
			if err := server.test(ctx, req, MaxRetries); err == nil {
				atomic.AddInt32(discovered, 1)
				c.servers <- server
			}
		}(id)
	}

	work.Wait()
	if *discovered == 0 {
		panic("no hosts discovered")
	} else {
		log.Printf("aconvert: discovered %d servers", *discovered)
	}
}

func baseURI(id interface{}) string {
	return fmt.Sprintf(BaseURITemplate, id)
}
