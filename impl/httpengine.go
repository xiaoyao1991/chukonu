package impl

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"time"

	"context"

	"github.com/satori/go.uuid"
	"github.com/xiaoyao1991/chukonu/core"
)

type HttpEngine struct {
	config core.ChukonuConfig
	http.Client
}

type ChukonuHttpRequest struct {
	id             uuid.UUID
	name           string
	timeout        time.Duration
	followRedirect bool
	keepAlive      bool
	postProcess    func(context.Context, core.ChukonuResponse) context.Context
	validate       func(core.ChukonuRequest, core.ChukonuResponse) bool
	*http.Request
}

type ChukonuHttpResponse struct {
	id       uuid.UUID
	duration time.Duration
	*http.Response
}

func NewChukonuHttpRequest(name string,
	timeout time.Duration,
	followRedirect bool,
	keepAlive bool,
	postProcess func(context.Context, core.ChukonuResponse) context.Context,
	validate func(core.ChukonuRequest, core.ChukonuResponse) bool,
	req *http.Request) ChukonuHttpRequest {
	return ChukonuHttpRequest{
		id:             uuid.NewV4(),
		name:           name,
		timeout:        timeout,
		followRedirect: followRedirect,
		keepAlive:      keepAlive,
		postProcess:    postProcess,
		validate:       validate,
		Request:        req,
	}
}

func (c ChukonuHttpRequest) Name() string {
	return c.name
}

func (c ChukonuHttpRequest) ID() uuid.UUID {
	return c.id
}

func (c ChukonuHttpRequest) Timeout() time.Duration {
	return c.timeout
}

func (c ChukonuHttpRequest) RawRequest() interface{} {
	return c.Request
}

func (c ChukonuHttpRequest) Validator() func(core.ChukonuRequest, core.ChukonuResponse) bool {
	return c.validate
}

func (c ChukonuHttpRequest) PostProcessor() func(context.Context, core.ChukonuResponse) context.Context {
	return c.postProcess
}

// TODO: reconsider the body=true param
func (c ChukonuHttpRequest) Dump() ([]byte, error) {
	return httputil.DumpRequestOut(c.Request, true)
}

func (c ChukonuHttpResponse) ID() uuid.UUID {
	return c.id
}

func (c ChukonuHttpResponse) RawResponse() interface{} {
	return c.Response
}

func (c ChukonuHttpResponse) Duration() time.Duration {
	return c.duration
}

func (c ChukonuHttpResponse) Status() string {
	return c.Response.Status
}

func (c ChukonuHttpResponse) Size() int64 {
	return c.Response.ContentLength
}

// TODO: reconsider the body=false param, set to false now because body would be closed early somewhere before dump is called
func (c ChukonuHttpResponse) Dump() ([]byte, error) {
	return httputil.DumpResponse(c.Response, false)
}

func NewHttpEngine(config core.ChukonuConfig) core.Engine {
	jar, _ := cookiejar.New(nil)

	return &HttpEngine{
		config: config,
		Client: http.Client{
			Timeout: config.RequestTimeout,
			Jar:     jar,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConnsPerHost:   config.Concurrency,
				MaxIdleConns:          config.Concurrency,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (e *HttpEngine) RunRequest(request core.ChukonuRequest) (core.ChukonuResponse, error) {
	start := time.Now()
	resp, err := e.Do(request.RawRequest().(*http.Request))
	duration := time.Since(start)
	if err != nil {
		return ChukonuHttpResponse{}, err
	}

	// defer resp.Body.Close() //TODO: delegate close responsibility to users?
	chukonuResp := ChukonuHttpResponse{
		id:       request.ID(),
		duration: duration,
		Response: resp,
	}

	return chukonuResp, nil
}

func (e *HttpEngine) ResetState() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	e.Client.Jar = jar
	return nil
}
