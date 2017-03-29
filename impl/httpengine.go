package impl

import (
	"net/http"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
)

type HttpEngine struct {
	config core.ChukonuConfig
	http.Client
}

type ChukonuHttpRequest struct {
	timeout        time.Duration
	followRedirect bool
	keepAlive      bool
	*http.Request
}

type ChukonuHttpResponse struct {
	duration time.Duration
	*http.Response
}

func (c ChukonuHttpRequest) Timeout() time.Duration {
	return c.timeout
}

func (c ChukonuHttpRequest) RawRequest() interface{} {
	return c.Request
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

func NewHttpEngine(config core.ChukonuConfig) *HttpEngine {
	return &HttpEngine{
		config: config,
		Client: http.Client{
			Timeout: config.RequestTimeout,
			// Jar: ,
			// Transport: &http.Transport{},
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

	chukonuResp := ChukonuHttpResponse{
		duration: duration,
		Response: resp,
	}
	return chukonuResp, nil
}
