package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/xiaoyao1991/chukonu/core"
	"github.com/xiaoyao1991/chukonu/impl"
)

type DruidRequestProvider struct {
}

func (m *DruidRequestProvider) Provide(queue chan core.ChukonuWorkflow) {
	// throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		// <-throttle
		// fmt.Printf("Generating %dth request\n", i)
		var workflow core.ChukonuWorkflow

		var fn1 = func(ctx context.Context) core.ChukonuRequest {
			req, err := http.NewRequest("GET", "http://sp17-cs525-g13-01.cs.illinois.edu:3000/druid/v2/datasources", nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			return impl.ChukonuHttpRequest{Request: req}
		}

		workflow.Requests = append(workflow.Requests, fn1)

		var fn2 = func(ctx context.Context) core.ChukonuRequest {
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
				return nil
			}
			return impl.ChukonuHttpRequest{Request: req}
		}

		workflow.Requests = append(workflow.Requests, fn2)

		queue <- workflow
		i++

		// if i == 100 {
		// 	break
		// }
	}

	// close(queue)
}

func main() {
	config := core.ChukonuConfig{Concurrency: 10, RequestTimeout: 5 * time.Minute}
	var engines []core.Engine = make([]core.Engine, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		engines[i] = impl.NewHttpEngine(config)
	}
	var pool core.Pool

	pool.Start(engines, &DruidRequestProvider{}, impl.NewHttpMetricsManager(), config)
}
