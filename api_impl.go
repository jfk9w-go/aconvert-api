package aconvert

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jfk9w-go/flu"
	"github.com/jfk9w-go/lego"
)

type ApiImpl struct {
	client *flu.Client
	pool   lego.Pool
}

func (api *ApiImpl) convert(host string, entity interface{}, opts Opts) (r *Response, err error) {
	var endpoint = host + "/convert/convert-batch.php"
	r = new(Response)
	err = api.client.NewRequest().
		Endpoint(endpoint).
		Post().Body(opts.body(entity)).Sync().Retrieve().
		StatusCodes(http.StatusOK).
		ReadJson(r).
		Done()

	if err != nil {
		return
	}

	err = r.check()
	if err != nil {
		return
	}

	return
}

func (api *ApiImpl) Convert(entity interface{}, opts Opts) (*Response, error) {
	var r *Response
	return r, api.pool.Process(func(host string) error {
		var _r, err = api.convert(host, entity, opts)
		if err == nil {
			r = _r
		}

		return err
	})
}

func (api *ApiImpl) Download(r *Response, resource flu.WriteResource) error {
	return api.client.NewRequest().
		Endpoint(GetHost(r.Server) + "/convert/p3r68-cdx67/" + r.Filename).
		Get().Retrieve().
		StatusCodes(http.StatusOK).
		ReadResource(resource).
		Done()
}

const HostTemplate = "https://s%s.aconvert.com"

func GetHost(number string) string {
	return fmt.Sprintf(HostTemplate, number)
}

const (
	hosts   = 30
	retries = 5
)

func (api *ApiImpl) discover(file string, format string) {
	var discovery = make(chan string, hosts)
	go func() {
		var hostsDiscovered = 0
		for host := range discovery {
			hostsDiscovered++

			var _host = host
			api.pool.With(func(task lego.Task) {
				var err = task.Run.(func(string) error)(_host)
				if err != nil && task.Retries < retries {
					task.Retry()
				} else {
					task.Complete(err)
				}

				time.Sleep(time.Second)
			})
		}

		if hostsDiscovered == 0 {
			panic("no hosts discovered")
		} else {
			log.Printf("Discovered %d hosts", hostsDiscovered)
		}
	}()

	var (
		resource  = flu.NewFileSystemResource(file)
		waitGroup sync.WaitGroup
	)

	waitGroup.Add(hosts)
	for i := 0; i < hosts; i++ {
		go func(i int) {
			defer waitGroup.Done()

			var (
				host = GetHost(strconv.Itoa(i))
				err  error
			)

			for j := 0; j < retries; j++ {
				_, err = api.convert(host, resource, NewOpts().TargetFormat(format))
				if err == nil {
					discovery <- host
					return
				}

				if j < retries-1 {
					time.Sleep(time.Duration(2^j) * time.Second)
				}
			}
		}(i)
	}

	waitGroup.Wait()
	close(discovery)
}
