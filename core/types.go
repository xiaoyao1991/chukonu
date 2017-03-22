package core

import (
	"net/http"
	"time"
)

const (
	MicrosecondsInOneSecond = 1e6
)

type ChukonuRequest interface {
	// Unwrap and get the actual request
	Request() interface{}
	Timeout() time.Duration
}

type ChukonuResponse interface {
	Duration() time.Duration
	Err() error
	Status() int
	Size() int
	RawResponse() interface{}
}

type RequestProvider <-chan ChukonuRequest

type ChukonuConfig struct {
	Name        string
	Concurrency int
	WarmUp      time.Duration
	Iterations  int
	Timeout     time.Duration
}

type ChukonuHttpRequest http.Request
type ChukonuHttpResponse http.Response

func (c *ChukonuHttpRequest) Request() interface{} {
	return c
}

func (c *ChukonuHttpResponse) RawResponse() interface{} {
	return c
}

type RequestThrottler interface {
	Throttler() <-chan time.Time
}

type Engine interface {
	Register(discoveryAddress string) error
	Run(requestProvider RequestProvider) error
	RunRequest(request ChukonuRequest) error
	Err() error
}
