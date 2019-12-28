package aconvert

import (
	"fmt"
	"log"
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
	servers    chan server
	maxRetries int
}

// NewClient creates a new aconvert HTTP client and runs server discovery in the background.
func NewClient(http *flu.Client, config Config) *Client {
	if config.Servers == nil {
		config.Servers = DefaultServers
	}
	if http == nil {
		http = flu.NewTransport().
			ResponseHeaderTimeout(3 * time.Minute).
			NewClient().
			Timeout(5 * time.Minute).
			AcceptResponseCodes(200)
	}
	client := &Client{
		Client:     http,
		servers:    make(chan server, len(config.Servers)),
		maxRetries: config.MaxRetries,
	}
	if config.Probe == nil {
		log.Printf("Using %d configured servers", len(config.Servers))
		for _, id := range config.Servers {
			client.servers <- server{http, baseURI(id)}
		}
	} else {
		go client.discover(config.Probe, config.Servers)
	}
	return client
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(in flu.Readable, opts Opts) (resp *Response, err error) {
	body := opts.body(in)
	for i := 0; i <= c.maxRetries; i++ {
		server := <-c.servers
		resp, err = server.convert(body)
		c.servers <- server
		if err != nil {
			continue
		} else {
			return
		}
	}
	return
}

func (c *Client) discover(probe *Probe, servers []int) {
	discovered := new(int32)
	workers := new(sync.WaitGroup)
	workers.Add(len(servers))
	body := make(Opts).TargetFormat(probe.Format).body(probe.File)
	for _, id := range servers {
		go func(id int) {
			server := server{c.Client, baseURI(id)}
			if server.test(body, c.maxRetries) {
				atomic.AddInt32(discovered, 1)
				c.servers <- server
			}
			workers.Done()
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
