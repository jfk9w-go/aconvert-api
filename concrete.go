package aconvert

import (
	"context"
	"errors"
	"math"
	_url "net/url"
	"time"

	fluhttp "github.com/jfk9w-go/flu/http"
)

var DefaultServers = []int{3 /*, 5*/, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27}

type concreteClient struct {
	convertURL *_url.URL
}

func newConcreteClient(baseURI string) concreteClient {
	convertURL, err := _url.Parse(baseURI + "/convert/convert-batch.php")
	if err != nil {
		panic(err)
	}
	return concreteClient{convertURL}
}

func (c concreteClient) test(ctx context.Context, req fluhttp.Request, maxRetries int) error {
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			timeout := time.Duration(math.Pow(2, float64(i))) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(timeout):
			}
		}

		_, err := c.convert(ctx, req)
		if err == nil {
			return nil
		}
	}

	return errors.New("exceeded max retries")
}

func (c concreteClient) convert(ctx context.Context, req fluhttp.Request) (*Response, error) {
	resp := new(Response)
	return resp, req.URL(c.convertURL).
		Context(ctx).
		Execute().
		DecodeBody(resp).
		Error
}
