package aconvert

import (
	"net/url"
	"strconv"

	"github.com/jfk9w-go/flu"
)

type Opts url.Values

func (o Opts) values() url.Values {
	return url.Values(o)
}

func (o Opts) Param(key, value string) Opts {
	o.values().Set(key, value)
	return o
}

func (o Opts) TargetFormat(targetFormat string) Opts {
	return o.Param("targetformat", targetFormat)
}

func (o Opts) VideoOptionSize(videoOptionSize int) Opts {
	return o.Param("videooptionsize", strconv.Itoa(videoOptionSize))
}

func (o Opts) Code(code int) Opts {
	return o.Param("code", strconv.Itoa(code))
}

func (o Opts) body(in flu.Readable) flu.BodyEncoderTo {
	if url, ok := in.(flu.URL); ok {
		return flu.FormValues(o.values()).
			Add("filelocation", "online").
			Add("file", string(url))
	} else {
		return flu.MultipartFormValues(o.values()).
			Add("filelocation", "local").
			File("file", in)
	}
}
