package aconvert

import (
	"net/url"
	"strconv"

	"github.com/jfk9w-go/flu"
)

type Options url.Values

func NewOpts() Options {
	return Options{}
}

func (opts Options) values() url.Values {
	return url.Values(opts)
}

func (opts Options) Param(key, value string) Options {
	opts.values().Set(key, value)
	return opts
}

func (opts Options) TargetFormat(targetFormat string) Options {
	return opts.Param("targetformat", targetFormat)
}

func (opts Options) VideoOptionSize(videoOptionSize int) Options {
	return opts.Param("videooptionsize", strconv.Itoa(videoOptionSize))
}

func (opts Options) Code(code int) Options {
	return opts.Param("code", strconv.Itoa(code))
}

func (opts Options) body(in flu.Readable) flu.BodyWriter {
	if url, ok := in.(flu.URL); ok {
		return flu.FormValues(opts.values()).
			Add("filelocation", "online").
			Add("file", string(url))
	} else {
		return flu.MultipartFormValues(opts.values()).
			Add("filelocation", "local").
			File("file", in)
	}
}
