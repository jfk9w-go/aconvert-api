package aconvert

import (
	"math"
	"time"

	"github.com/jfk9w-go/flu"
)

var DefaultServers = []int{3 /*5, */, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27}

type server struct {
	http    *flu.Client
	baseURI string
}

func (s server) test(body flu.BodyWriter, maxRetries int) bool {
	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			timeout := time.Duration(math.Pow(2, float64(i))) * time.Second
			time.Sleep(timeout)
		}
		_, err := s.convert(body)
		if err == nil {
			return true
		}
	}
	return false
}

func (s server) convert(body flu.BodyWriter) (*Response, error) {
	resp := new(Response)
	err := s.http.
		POST(s.baseURI + "/convert/convert-batch.php").
		Body(body).Buffer().
		Execute().
		Read(resp).
		Error
	if err != nil {
		return nil, err
	}
	return resp, nil
}
