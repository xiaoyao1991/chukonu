package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/xiaoyao1991/chukonu/core"
	"github.com/xiaoyao1991/chukonu/impl"
)

type DruidRequestProvider struct {
}

func (m *DruidRequestProvider) MetricsManager() core.MetricsManager {
	return impl.NewHttpMetricsManager()
}

func (m *DruidRequestProvider) UseEngine() func(core.ChukonuConfig) core.Engine {
	return impl.NewHttpEngine
}
func (m *DruidRequestProvider) Provide(queue chan *core.ChukonuWorkflow) {
	// throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		// <-throttle
		// fmt.Printf("Generating %dth request\n", i)
		workflow := core.NewWorkflow("druid_workflow")

		var fn1 = func(ctx context.Context) core.ChukonuRequest {
			req, err := http.NewRequest("GET", "http://sp17-cs525-g13-01.cs.illinois.edu:3000/druid/v2/datasources", nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			return impl.NewChukonuHttpRequest("datasources_req", 0, true, true, func(ctx context.Context, resp core.ChukonuResponse) context.Context {
				defer resp.RawResponse().(*http.Response).Body.Close()
				// dump, err := resp.Dump()
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// fmt.Println(string(dump))
				bodyBytes, _ := ioutil.ReadAll(resp.RawResponse().(*http.Response).Body)
				bodyString := string(bodyBytes)
				datasource := bodyString[2 : len(bodyString)-2]
				return context.WithValue(ctx, "datasource", datasource)
			}, nil, req)
		}

		workflow.AddRequest(fn1)

		var fn2 = func(ctx context.Context) core.ChukonuRequest {
			var jsonStr = []byte(fmt.Sprintf(`
        {
          "queryType" : "topN",
          "dataSource" : "%s",
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
        }`, ctx.Value("datasource")))
			req, err := http.NewRequest("POST", "http://sp17-cs525-g13-01.cs.illinois.edu:3000/druid/v2/", bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				fmt.Println(err)
				return nil
			}
			return impl.NewChukonuHttpRequest("topN", 0, true, true, func(ctx context.Context, resp core.ChukonuResponse) context.Context {
				defer resp.RawResponse().(*http.Response).Body.Close()
				// dump, err := resp.Dump()
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// fmt.Println(string(dump))
				return ctx
			}, nil, req)
		}

		workflow.AddRequest(fn2)

		// fmt.Print("adding workflow to queue")
		queue <- workflow
		i++

		if i == 50 {
			break
		}
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
	pool.Start(engines, &DruidRequestProvider{}, impl.NewHttpMetricsManager(), config, nil, nil)
}
