package core

import "time"

const (
	MicrosecondsInOneSecond = 1e6
)

type ChukonuRequest interface {
	// RequestType() string
	// ID() int64
	RawRequest() interface{}
	Timeout() time.Duration
	Validator() func(ChukonuRequest, ChukonuResponse) bool
}

type ChukonuResponse interface {
	// ID() int64
	Duration() time.Duration
	Status() string
	Size() int64
	RawResponse() interface{}
}

// A flow of requests that will be run sequentially in order by one goroutine
type ChukonuWorkflow struct {
	Requests []ChukonuRequest
}

type RequestProvider interface {
	Provide(chan ChukonuWorkflow)
}

type ChukonuConfig struct {
	Concurrency int
	WarmUp      time.Duration
	Iterations  int
	// Cookie????
	TotalTimeout   time.Duration
	RequestTimeout time.Duration
}

type RequestThrottler interface {
	Throttle() chan time.Time
}

type Engine interface {
	// LoadMetricsManager(metricsManager MetricsManager) error
	// Run(requestProvider RequestProvider) error
	RunRequest(request ChukonuRequest) (ChukonuResponse, error)
}

type MetricsManager interface {
	RecordRequest(request ChukonuRequest)
	RecordResponse(response ChukonuResponse)
	RecordError(err error)
	// RecordThroughput(throughput float64)
	GetQueue() chan int
	MeasureThroughput()
	RecordThroughput(sec float64)
	GetThroughput() int
	SampleThroughput()
}

type LogReplayer interface {
	ParseLog(filename string) RequestProvider
}
