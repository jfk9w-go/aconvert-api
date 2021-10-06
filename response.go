package aconvert

import (
	"io"

	"github.com/jfk9w-go/flu"
	"github.com/pkg/errors"
)

type Response struct {
	Server   string `json:"server"`
	Filename string `json:"filename"`
	State    string `json:"state"`
	Result   string `json:"result"`
	host     string `json:"-"`
}

func (r *Response) DecodeFrom(reader io.Reader) error {
	err := flu.JSON(r).DecodeFrom(reader)
	if err != nil {
		return err
	}

	if r.State != "SUCCESS" {
		return errors.Errorf("state is %s, not SUCCESS (%s)", r.State, r.Result)
	}

	r.host = host(r.Server)
	return nil
}

func (r *Response) URL() string {
	return r.host + "/convert/p3r68-cdx67/" + r.Filename
}
