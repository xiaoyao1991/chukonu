package core

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
}

func (p Pool) Start(engines []Engine, provider RequestProvider, metricsManager MetricsManager, config ChukonuConfig, fuse chan bool, ack chan bool) {
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	queue := make(chan *ChukonuWorkflow, config.Concurrency)
	go provider.Provide(queue)
	// go metricsManager.MeasureThroughput() // start a goroutine to listen for atomic changes
	// go metricsManager.SampleMetrics()
	// go metricsManager.StartRecording()
	// throughputQueue := metricsManager.GetThroughputQueue()
	// requestQueue := metricsManager.GetRequestQueue()
	// responseQueue := metricsManager.GetResponseQueue()
	// errorQueue := metricsManager.GetErrorQueue()
	throughputQueue, requestQueue, responseQueue, errorQueue := metricsManager.StartRecording()
	startTime := time.Now()
	var i int
	for i = 0; i < config.Concurrency; i++ {
		if fuse != nil {
			_, ok := <-fuse
			if !ok {
				break
			}
		}
		go func(i int) {
			defer wg.Done()
			for workflow := range queue {
				for workflow.HasNext() {
					req := workflow.Next()
					requestQueue <- req
					// fmt.Printf("goroutine %d running request...", i)
					resp, err := engines[i].RunRequest(req)
					throughputQueue <- 1
					if err != nil {
						// TODO: differentiate custom errors
						fmt.Print("Http response Error: ")
						fmt.Println(err)
						errorQueue <- err
						continue
					} else { //TODO: what to do on error
						workflow.PostProcess(req, resp)
					}
					responseQueue <- resp
				}

				engines[i].ResetState()
			}
		}(i)

		ack <- true
	}
	close(ack)

	// TODO: handle when i != concurrency
	if i < config.Concurrency {
		fmt.Println("Not able to spawn all users, spawned: ", i)
	}

	wg.Wait()
	elapseTime := time.Since(startTime)
	metricsManager.RecordThroughput(elapseTime.Seconds())
	close(throughputQueue)
	// metricsManager.RecordThroughput(float64(requestSent) / elapseTime.Seconds())
}
