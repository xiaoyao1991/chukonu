package core

import (
	"time"

	"context"

	"github.com/satori/go.uuid"
)

const (
	MicrosecondsInOneSecond = 1e6
)

type ChukonuRequest interface {
	ID() uuid.UUID
	Name() string // name of a type of request
	RawRequest() interface{}
	Timeout() time.Duration
	Validator() func(ChukonuRequest, ChukonuResponse) bool
	PostProcessor() func(context.Context, ChukonuResponse) context.Context
	Dump() ([]byte, error)
}

type ChukonuResponse interface {
	ID() uuid.UUID
	Duration() time.Duration
	Status() string
	Size() int64
	RawResponse() interface{}
	Dump() ([]byte, error)
}

// A flow of requests that will be run sequentially in order by one goroutine
type ChukonuWorkflow struct {
	Name     string // name of a type of workflow
	Ctx      context.Context
	Requests []func(context.Context) ChukonuRequest
	iter     int
}

func NewWorkflow(name string) *ChukonuWorkflow {
	return &ChukonuWorkflow{
		Name: name,
		Ctx:  context.Background(),
		iter: 0,
	}
}

func (c *ChukonuWorkflow) AddRequest(fn func(context.Context) ChukonuRequest) {
	c.Requests = append(c.Requests, fn)
}

func (c *ChukonuWorkflow) HasNext() bool {
	return c.iter < len(c.Requests)
}

func (c *ChukonuWorkflow) Next() ChukonuRequest {
	defer func() { c.iter = c.iter + 1 }()
	fn := c.Requests[c.iter]
	return fn(c.Ctx)
}

func (c *ChukonuWorkflow) PostProcess(req ChukonuRequest, res ChukonuResponse) {
	fn := req.PostProcessor()
	if fn != nil {
		newctx := fn(c.Ctx, res)
		c.Ctx = newctx
	}
}

type RequestProvider interface {
	Provide(chan *ChukonuWorkflow)
	UseEngine() func(ChukonuConfig) Engine
	MetricsManager() MetricsManager
}

// should be able to represent as json
type ChukonuConfig struct {
	TenantId       string                 `json:"tenantId"`
	Concurrency    int                    `json:"concurrency"`
	Iterations     int                    `json:"iterations"`
	TotalTimeout   time.Duration          `json:"totalTimeout"`
	RequestTimeout time.Duration          `json:"requestTimeout"`
	EngineSpec     map[string]interface{} `json:"engineSpec"`
}

type Engine interface {
	RunRequest(request ChukonuRequest) (ChukonuResponse, error)
	ResetState() error
}

type MetricsManager interface {
	MeasureThroughput()
	RecordThroughput(sec float64)
	GetThroughput() int
	SampleMetrics()
	StartRecording() (chan int, chan ChukonuRequest, chan ChukonuResponse, chan error)
	SetConsulConfig(tenantId string, cid string, consulAddress string)
}

type LogReplayer interface {
	ParseLog(filename string) RequestProvider
}
