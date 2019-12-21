package aconvert

import (
	"log"
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
	err := s.http.NewRequest().
		POST().
		Resource(s.baseURI + "/convert/convert-batch.php").
		Body(body).
		Buffer().
		Send().
		Read(resp). // ReadBody checks the Content-Type header which doesn't match in this case
		Error
	if err != nil {
		log.Printf("Failed to convert on %s: resp = %v, err = %v", s.baseURI, resp, err)
		return nil, err
	}
	return resp, nil
}
