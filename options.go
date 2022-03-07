package aconvert

import (
	"net/url"
	"os"
	"strconv"

	"github.com/jfk9w-go/flu"
	httpf "github.com/jfk9w-go/flu/httpf"
	"github.com/jfk9w-go/flu/me3x"
	"github.com/pkg/errors"
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

func (o Opts) Labels() me3x.Labels {
	return me3x.Labels{}.
		Add("targetformat", o.values().Get("targetformat"))
}

const Legal = "We DO NOT allow directly running our PHP programs from any third-party websites, software or apps. Illegal piracy may return files with warning messages!"

func (o Opts) makeRequest(url string, in flu.Input) (req *httpf.RequestBuilder, err error) {
	var body flu.EncoderTo
	counter := new(flu.IOCounter)
	o.values().Set("legal", Legal)
	o.values().Set("chunks", "1")
	o.values().Set("chunk", "0")
	form := new(httpf.Form).SetAll(o.values())
	if u, ok := in.(flu.URL); ok {
		form.
			Set("filelocation", "online").
			Set("file", u.Unmask())

		if err = flu.EncodeTo(form, counter); err != nil {
			err = errors.Wrap(err, "on multipart count")
			return
		}

		body = form
	} else {
		multipart := form.
			Set("filelocation", "local").
			Multipart()

		if file, ok := in.(flu.File); ok {
			if err = flu.EncodeTo(multipart, counter); err != nil {
				err = errors.Wrap(err, "on multipart write count")
				return
			}
			var stat os.FileInfo
			if stat, err = os.Stat(file.Path()); err != nil {
				return
			}

			counter.Add(stat.Size() + 170)
			multipart = multipart.File("file", "", in)
		} else {
			multipart = multipart.File("file", "", in)
			if err = flu.EncodeTo(multipart, counter); err != nil {
				err = errors.Wrap(err, "on file write count")
				return
			}
		}

		body = multipart
	}

	req = httpf.POST(url, body)
	req.Request.ContentLength = counter.Value()
	return
}
