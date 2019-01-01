package aconvert

import (
	"fmt"
	"time"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/lego"
)

type Response struct {
	Server   string `json:"server"`
	Filename string `json:"filename"`
	State    string `json:"state"`
}

func (r *Response) check() error {
	if r.State != "SUCCESS" {
		return fmt.Errorf("state is %s, not SUCCESS", r.State)
	}

	return nil
}

type Api interface {
	Convert(interface{}, Opts) (*Response, error)
	Download(*Response, flu.WriteResource) error
}

func NewApi(client *flu.Client, config Config) Api {
	if client == nil {
		client = flu.NewClient(nil).
			ResponseHeaderTimeout(120 * time.Second)
	}

	var api = &ApiImpl{
		client: client,
		pool:   lego.NewPool(),
	}

	go api.discover(config.TestFile, config.TestFormat)
	return api
}
