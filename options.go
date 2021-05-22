package aconvert

import (
	"net/url"
	"os"
	"strconv"

	"github.com/pkg/errors"

	fluhttp "github.com/jfk9w-go/flu/http"

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

func (o Opts) makeRequest(client *fluhttp.Client, in flu.Input) (req *fluhttp.Request, err error) {
	var body flu.EncoderTo
	counter := new(flu.IOCounter)
	if u, ok := in.(flu.URL); ok {
		form := new(fluhttp.Form).AddValues(o.values()).
			Add("filelocation", "online").
			Add("file", u.Unmask())

		if err = flu.EncodeTo(form, counter); err != nil {
			err = errors.Wrap(err, "on multipart count")
			return
		}

		body = form
	} else {
		multipart := fluhttp.NewMultipartForm().
			AddValues(o.values()).
			Add("filelocation", "local")

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

	req = client.POST("").BodyEncoder(body)
	req.Request.ContentLength = counter.Value()
	return
}
