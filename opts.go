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

func (opts Opts) Param(key, value string) Opts {
	opts.values().Set(key, value)
	return opts
}

func (opts Opts) TargetFormat(targetFormat string) Opts {
	return opts.Param("targetformat", targetFormat)
}

func (opts Opts) VideoOptionSize(videoOptionSize int) Opts {
	return opts.Param("videooptionsize", strconv.Itoa(videoOptionSize))
}

func (opts Opts) Code(code int) Opts {
	return opts.Param("code", strconv.Itoa(code))
}

func (opts Opts) body(entity interface{}) flu.BodyWriter {
	switch entity := entity.(type) {
	case string:
		return flu.FormValues(opts.values()).
			Add("filelocation", "online").
			Add("file", entity)

	case flu.ReadResource:
		opts.values().Set("filelocation", "local")
		return flu.MultipartFormValues(opts.values()).
			Add("filelocation", "local").
			Resource("file", entity)

	default:
		panic(fmt.Errorf("unknown entity type: %T", entity))
	}
}
