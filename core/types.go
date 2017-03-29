package core

import "time"

const (
	MicrosecondsInOneSecond = 1e6
)

type ChukonuRequest interface {
	RawRequest() interface{} // Unwrap and get the actual request
	Timeout() time.Duration
	//TODO
	//getValidator() func(ChukonuRequest, ChukonuResponse) bool
}

type ChukonuResponse interface {
	Duration() time.Duration
	Status() string
	Size() int64
	RawResponse() interface{}
}

type RequestProvider interface {
	Provide() chan ChukonuRequest
}

type ChukonuConfig struct {
	Concurrency int
	WarmUp      time.Duration
	Iterations  int
	// Cookie????
	TotalTimeout   time.Duration
	RequestTimeout time.Duration
}

type SLA struct {
	LatencySLA    map[float32]float64 // percentile to SLA
	ThroughputSLA map[float32]float64 // percentile to SLA
}

type ResponseValidator interface {
	Validate(response ChukonuResponse) bool
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
}

type LogReplayer interface {
	ParseLog(filename string) RequestProvider
}
