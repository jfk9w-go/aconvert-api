package aconvert

import (
	"net/url"
	"strconv"
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
