package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
	"github.com/xiaoyao1991/chukonu/impl"
)

type DruidRequestProvider struct {
	queue chan core.ChukonuRequest
}

func (m *DruidRequestProvider) Provide() chan core.ChukonuRequest {
	if m.queue == nil {
		m.queue = make(chan core.ChukonuRequest, 10)
	}

	return m.queue
}

func (m *DruidRequestProvider) Gen() {
	queue := m.Provide()

	// throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		// <-throttle
		// fmt.Printf("Generating %dth request\n", i)
		var jsonStr = []byte(`
      {
        "queryType" : "topN",
        "dataSource" : "wikipedia",
        "intervals" : ["2013-08-01/2013-08-03"],
        "granularity" : "all",
        "dimension" : "page",
        "metric" : "edits",
        "threshold" : 25,
        "aggregations" : [
          {
            "type" : "longSum",
            "name" : "edits",
            "fieldName" : "count"
          }
        ]
      }`)
		req, err := http.NewRequest("POST", "http://sp17-cs525-g13-01.cs.illinois.edu:3000/druid/v2/", bytes.NewBuffer(jsonStr))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			fmt.Println(err)
		}
		queue <- impl.ChukonuHttpRequest{Request: req}
		i++

		if i == 100 {
			break
		}
	}

	close(queue)
}

func main() {
	config := core.ChukonuConfig{Concurrency: 1, RequestTimeout: 5 * time.Minute}
	httpengine := impl.NewHttpEngine(config)
	var pool core.Pool

	provider := &DruidRequestProvider{}
	go provider.Gen()
	pool.Start(httpengine, provider, impl.NewHttpMetricsManager(), config)
}
