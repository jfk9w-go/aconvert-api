package aconvert

import (
	"net/url"

	"github.com/jfk9w-go/flu"
)

// Media represents a format conversion input.
type Media interface {
	body(url.Values) flu.BodyEncoderTo
}

// URL is an online res from which the media will be pulled.
type URL struct {
	url string
}

func (u URL) body(values url.Values) flu.BodyEncoderTo {
	return flu.FormValues(values).
		Add("filelocation", "online").
		Add("file", u.url)
}

// Resource is a flu.ReadResource which will be uploaded for conversion.
// Since the request may be retried, it is mandatory that resource may be read multiple times.
type Resource struct {
	res flu.ResourceReader
}

func (r Resource) body(values url.Values) flu.BodyEncoderTo {
	return flu.MultipartFormValues(values).
		Add("filelocation", "local").
		Resource("file", r.res)
}
