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
}

func (m *DruidRequestProvider) Provide(queue chan core.ChukonuRequest) {
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

		// if i == 100 {
		// 	break
		// }
	}

	// close(queue)
}

func main() {
	config := core.ChukonuConfig{Concurrency: 10, RequestTimeout: 5 * time.Minute}
	httpengine := impl.NewHttpEngine(config)
	var pool core.Pool

	pool.Start(httpengine, &DruidRequestProvider{}, impl.NewHttpMetricsManager(), config)
}
