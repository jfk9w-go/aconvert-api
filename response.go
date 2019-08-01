package aconvert

import (
	"fmt"
)

// Response represents a aconvert response JSON.
type Response struct {

	// Server is the number of a server with the conversion result.
	Server string `json:"server"`

	// Filename is the file name.
	Filename string `json:"filename"`

	// State is the request state (SUCCESS or ERROR).
	State string `json:"state"`

	host string
}

func (r *Response) init() error {
	if r.State != "SUCCESS" {
		return fmt.Errorf("state is %s, not SUCCESS", r.State)
	}

	r.host = host(r.Server)
	return nil
}
