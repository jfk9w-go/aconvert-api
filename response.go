package aconvert

import (
	"io"

	"github.com/jfk9w-go/flu"
	"github.com/pkg/errors"
)

// Response represents a aconvert response JSON.
type Response struct {

	// Server is the number of a server with the conversion result.
	Server string `json:"server"`

	// Filename is the file baseURI.
	Filename string `json:"filename"`

	// State is the request state (SUCCESS or ERROR).
	State string `json:"state"`

	host string
}

func (r *Response) ReadFrom(reader io.Reader) error {
	err := flu.JSON(r).ReadFrom(reader)
	if err != nil {
		return err
	}
	if r.State != "SUCCESS" {
		return errors.Errorf("state is %s, not SUCCESS", r.State)
	}
	r.host = baseURI(r.Server)
	return nil
}

func (r *Response) URL() string {
	return r.host + "/convert/p3r68-cdx67/" + r.Filename
}
