package aconvert

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/jfk9w-go/flu"
)

var DefaultServers = []int{3 /*5, */, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27}

type server struct {
	http    *flu.Client
	baseURI string
}

func (s server) test(ctx context.Context, body flu.BodyEncoderTo, maxRetries int) error {
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			timeout := time.Duration(math.Pow(2, float64(i))) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(timeout):
			}
		}

		_, err := s.convert(ctx, body)
		if err == nil {
			return nil
		}
	}

	return errors.New("exceeded max retries")
}

func (s server) convert(ctx context.Context, body flu.BodyEncoderTo) (*Response, error) {
	resp := new(Response)
	return resp, s.http.
		POST(s.baseURI + "/convert/convert-batch.php").
		Body(body).Buffer().
		Context(ctx).
		Execute().
		Decode(resp).
		Error
}
