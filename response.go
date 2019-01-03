package aconvert

import (
	"fmt"
	"strconv"
)

// Response represents a aconvert response JSON.
type Response struct {

	// Server is the number of a server with the conversion result.
	Server string `json:"server"`

	// Filename is the file name.
	Filename string `json:"filename"`

	// State is the request state (SUCCESS or ERROR).
	State string `json:"state"`

	server int
}

func (r *Response) init() error {
	if r.State != "SUCCESS" {
		return fmt.Errorf("state is %s, not SUCCESS", r.State)
	}

	var err error
	r.server, err = strconv.Atoi(r.Server)

	return err
}
