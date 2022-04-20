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

	data string
	host string
}

func (r *Response) DecodeFrom(reader io.Reader) error {
	var buf flu.ByteBuffer
	if _, err := flu.Copy(flu.IO{R: reader}, &buf); err != nil {
		return err
	}

	r.data = buf.String()

	if err := flu.DecodeFrom(&buf, flu.JSON(r)); err != nil {
		return errors.Wrapf(err, "failed to decode [%s]: %v", r, err)
	}

	if r.State != "SUCCESS" {
		return errors.Errorf("state is %s, not SUCCESS (%s)", r.State, r)
	}

	r.host = host(r.Server)
	return nil
}

func (r *Response) String() string {
	if r == nil {
		return "<nil>"
	}

	return r.data
}

func (r *Response) URL() string {
	return r.host + "/convert/p3r68-cdx67/" + r.Filename
}
