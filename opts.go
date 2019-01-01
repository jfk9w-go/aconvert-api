package aconvert

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/jfk9w-go/flu"
)

type Opts url.Values

func NewOpts() Opts {
	return Opts{}
}

func (opts Opts) values() url.Values {
	return url.Values(opts)
}

func (opts Opts) TargetFormat(targetFormat string) Opts {
	opts.values().Set("targetformat", targetFormat)
	return opts
}

func (opts Opts) VideoOptionSize(videoOptionSize int) Opts {
	opts.values().Set("videooptionsize", strconv.Itoa(videoOptionSize))
	return opts
}

func (opts Opts) Code(code int) Opts {
	opts.values().Set("code", strconv.Itoa(code))
	return opts
}

func (opts Opts) body(entity interface{}) flu.RequestBodyBuilder {
	switch entity := entity.(type) {
	case string:
		return flu.FormWith(opts.values()).
			Add("filelocation", "online").
			Add("file", entity)

	case flu.ReadResource:
		opts.values().Set("filelocation", "local")
		return flu.MultipartFormWith(opts.values()).
			Add("filelocation", "local").
			Resource("file", entity)

	default:
		panic(fmt.Errorf("unknown entity type: %T", entity))
	}
}
