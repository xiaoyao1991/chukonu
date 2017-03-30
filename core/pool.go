package core

import (
	"fmt"
	"sync"
	"time"
)

type Pool struct {
}

func (p Pool) Start(engine Engine, provider RequestProvider, metricsManager MetricsManager, config ChukonuConfig) {
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	queue := provider.Provide()
	go metricsManager.MeasureThroughput()
	// go metricsManager.SampleThroughput()
	//START SENDING REQUEST
	throughputQueue := metricsManager.GetQueue()
	startTime := time.Now()
	for i := 0; i < config.Concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			for req := range queue {
				metricsManager.RecordRequest(req)
				fmt.Println(fmt.Sprintf("goroutine %d running request...", i))
				resp, err := engine.RunRequest(req)
				throughputQueue <- 1
				if err != nil {
					fmt.Println(err)
					metricsManager.RecordError(err)
					continue
				}
				metricsManager.RecordResponse(resp)
				fmt.Println("\t" + resp.Status())
			}
		}(i)
	}
	wg.Wait()
	elapseTime := time.Since(startTime)
	metricsManager.RecordThroughput(elapseTime.Seconds())
	close(throughputQueue)
	// metricsManager.RecordThroughput(float64(requestSent) / elapseTime.Seconds())

}
