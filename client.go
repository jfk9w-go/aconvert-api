package aconvert

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/flu/backoff"
	httpf "github.com/jfk9w-go/flu/httpf"
	"github.com/jfk9w-go/flu/me3x"
	"github.com/pkg/errors"
)

var (
	DefaultServerIDs = []int{3 /*, 5*/, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27}
	MaxRetries       = 3
)

// Probe denotes a file to be used for discovering servers.
type Probe struct {
	File   flu.File
	Format string
}

type Config struct {
	ServerIDs []int
	Probe     *Probe
}

// Client is an entity allowing access to aconvert.
type Client struct {
	*httpf.Client
	servers chan server
	metrics me3x.Registry
}

func NewClient(ctx context.Context, metrics me3x.Registry, config *Config) *Client {
	if config.ServerIDs == nil {
		config.ServerIDs = DefaultServerIDs
	}

	client := &Client{
		Client: httpf.NewTransport().
			ResponseHeaderTimeout(3*time.Minute).
			NewClient().
			Timeout(5*time.Minute).
			AcceptStatus(http.StatusOK).
			SetHeader("Referer", "https://www.aconvert.com/"),
		servers: make(chan server, len(config.ServerIDs)),
		metrics: metrics,
	}

	if config.Probe == nil {
		log.Printf("using %d configured aconvert servers", len(config.ServerIDs))
		for _, id := range config.ServerIDs {
			client.servers <- makeServer(id)
		}
	} else {
		_ = new(flu.WaitGroup).Go(ctx, func(ctx context.Context) {
			client.discover(ctx, config.Probe, config.ServerIDs)
		})
	}

	return client
}

// Convert converts the provided media and returns a response.
func (c *Client) Convert(ctx context.Context, in flu.Input, opts Opts) (*Response, error) {
	var resp *Response
	retry := backoff.Retry{
		Retries: MaxRetries,
		Backoff: backoff.Const(0),
		Body: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case server := <-c.servers:
				defer func() {
					if ctx.Err() == nil {
						select {
						case <-ctx.Done():
						case c.servers <- server:
						}
					}
				}()

				var err error
				resp, err = c.convert(ctx, server, in, opts)
				return err
			}
		},
	}

	return resp, retry.Do(ctx)
}

func (c *Client) convert(ctx context.Context, server server, in flu.Input, opts Opts) (*Response, error) {
	req, err := opts.Code(81000).makeRequest(c.Client, in)
	if err != nil {
		return nil, errors.Wrap(err, "make request")
	}

	metrics := c.metrics.WithPrefix("convert")
	labels := server.Labels().AddAll(opts.Labels())
	metrics.Counter("attempts", labels).Inc()

	resp := new(Response)
	if err := req.Context(ctx).
		URL(server.convertURL).
		Execute().
		DecodeBody(resp).
		Error; err != nil {
		metrics.Counter("failed", labels).Inc()
		return nil, err
	} else {
		metrics.Counter("ok", labels).Inc()
		return resp, nil
	}
}

func (c *Client) discover(ctx context.Context, probe *Probe, serverIDs []int) {
	discovered := new(int32)
	work := new(flu.WaitGroup)
	for i := range serverIDs {
		serverID := serverIDs[i]
		_ = work.Go(ctx, func(ctx context.Context) {
			server := makeServer(serverID)
			retry := backoff.Retry{
				Retries: MaxRetries,
				Backoff: backoff.Exp{
					Base:  2,
					Power: 1,
				},
				Body: func(ctx context.Context) error {
					_, err := c.convert(ctx, server, probe.File, make(Opts).TargetFormat(probe.Format))
					return err
				},
			}

			if err := retry.Do(ctx); err != nil {
				log.Printf("aconvert %s init failed: %s", server.Labels(), err)
			} else {
				log.Printf("aconvert %s init ok", server.Labels())
				atomic.AddInt32(discovered, 1)
				c.servers <- server
			}
		})
	}

	work.Wait()
	if *discovered == 0 {
		log.Panicf("no aconvert hosts discovered")
	} else {
		log.Printf("discovered %d aconvert servers", *discovered)
	}
}

func makeServer(id interface{}) server {
	url, err := url.Parse(host(id) + "/convert/convert3.php")
	if err != nil {
		log.Panicf("invalid convert-batch URL: %s", err)
	}

	return server{url, id}
}

func host(serverID interface{}) string {
	return fmt.Sprintf("https://s%v.aconvert.com", serverID)
}

type server struct {
	convertURL *url.URL
	id         interface{}
}

func (s server) Labels() me3x.Labels {
	return me3x.Labels{}.
		Add("server", s.id)
}
