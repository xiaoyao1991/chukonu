package impl

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
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
	validate       func(core.ChukonuRequest, core.ChukonuResponse) bool
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

func (c ChukonuHttpRequest) Validator() func(core.ChukonuRequest, core.ChukonuResponse) bool {
	return c.validate
}

// TODO: reconsider the body=true param
func (c ChukonuHttpRequest) Dump() ([]byte, error) {
	return httputil.DumpRequestOut(c.Request, true)
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

func NewHttpEngine(config core.ChukonuConfig) *HttpEngine {
	jar, _ := cookiejar.New(nil)

	return &HttpEngine{
		config: config,
		Client: http.Client{
			Timeout: config.RequestTimeout,
			Jar:     jar,
			// Transport: &http.Transport{},
		},
	}
}

// TODO: need to add a param to consume the resp body, if no, then close the body right away
// likely it's gonna be some io.WriterCloser, or custom parser if users want to use the data in response?
// TODO: dump response https://golang.org/src/net/http/httputil/dump.go?s=8166:8231#L271
func (e *HttpEngine) RunRequest(request core.ChukonuRequest) (core.ChukonuResponse, error) {
	start := time.Now()
	resp, err := e.Do(request.RawRequest().(*http.Request))
	duration := time.Since(start)
	if err != nil {
		return ChukonuHttpResponse{}, err
	}

	defer resp.Body.Close() //TODO: where to close body
	chukonuResp := ChukonuHttpResponse{
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
