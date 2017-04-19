package core

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
}

func (p Pool) Start(engines []Engine, provider RequestProvider, metricsManager MetricsManager, config ChukonuConfig) {
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	queue := make(chan *ChukonuWorkflow, config.Concurrency)
	go provider.Provide(queue)
	go metricsManager.MeasureThroughput() // start a goroutine to listen for atomic changes
	go metricsManager.SampleThroughput()

	throughputQueue := metricsManager.GetQueue()
	startTime := time.Now()
	for i := 0; i < config.Concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			for workflow := range queue {
				for workflow.HasNext() {
					req := workflow.Next()
					// metricsManager.RecordRequest(req)

					// fmt.Printf("goroutine %d running request...", i)
					resp, err := engines[i].RunRequest(req)
					throughputQueue <- 1
					if err != nil {
						// TODO: differentiate custom errors
						fmt.Println(err)
						metricsManager.RecordError(err)
						continue
					} else { //TODO: what to do on error
						workflow.PostProcess(req, resp)
					}
					metricsManager.RecordResponse(resp)
				}

				engines[i].ResetState()
			}
		}(i)
	}
	wg.Wait()
	elapseTime := time.Since(startTime)
	metricsManager.RecordThroughput(elapseTime.Seconds())
	close(throughputQueue)
	// metricsManager.RecordThroughput(float64(requestSent) / elapseTime.Seconds())
}
