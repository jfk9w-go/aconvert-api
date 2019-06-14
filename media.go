package aconvert

import (
	"net/url"

	"github.com/jfk9w-go/flu"
)

// Media represents a format conversion input.
type Media interface {
	body(url.Values) flu.BodyWriter
}

// URL is an online resource from which the media will be pulled.
type URL struct {
	url string
}

func (u URL) body(values url.Values) flu.BodyWriter {
	return flu.FormValues(values).
		Add("filelocation", "online").
		Add("file", u.url)
}

// Resource is a flu.ReadResource which will be uploaded for conversion.
// Since the request may be retried, it is mandatory that resource may be read multiple times.
type Resource struct {
	resource flu.ReadResource
}

func (r Resource) body(values url.Values) flu.BodyWriter {
	return flu.MultipartFormValues(values).
		Add("filelocation", "local").
		Resource("file", r.resource)
}
