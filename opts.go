package aconvert

import (
	"net/url"
	"strconv"
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
