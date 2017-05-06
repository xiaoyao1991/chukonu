package impl

import (
	"fmt"
	"sort"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
)

type HttpMetricsManager struct {
	requestSent int
	durationSum time.Duration
	numErrors   int
	numRecords  int
	histogram   map[time.Duration]int

	// channels
	countQueue    chan int
	requestQueue  chan core.ChukonuRequest
	responseQueue chan core.ChukonuResponse
	errorQueue    chan error
}

func (m *HttpMetricsManager) RecordRequest(request core.ChukonuRequest) {
}
func (m *HttpMetricsManager) RecordResponse(response core.ChukonuResponse) {
	if response, ok := response.(ChukonuHttpResponse); ok {
		m.histogram[response.Duration()/time.Millisecond] += 1
		m.numRecords += 1
		m.durationSum += response.Duration() / time.Millisecond
	} else {
		panic("Response not a HTTPResponse")
	}

}
func (m *HttpMetricsManager) RecordError(err error) {
	m.numErrors += 1
	m.numRecords += 1
}
func (m *HttpMetricsManager) RecordThroughput(t float64) {
	fmt.Printf("Sent: %v, time: %v, throughput: %v\n", m.requestSent, t, float64(m.requestSent)/t)
}
func (m *HttpMetricsManager) GetThroughputQueue() chan int {
	return m.countQueue
}
func (m *HttpMetricsManager) GetRequestQueue() chan core.ChukonuRequest {
	return m.requestQueue
}
func (m *HttpMetricsManager) GetResponseQueue() chan core.ChukonuResponse {
	return m.responseQueue
}
func (m *HttpMetricsManager) GetErrorQueue() chan error {
	return m.errorQueue
}
func (m *HttpMetricsManager) GetThroughput() int {
	return m.requestSent
}

// return the sampletime(duration) at input percentile
func (m *HttpMetricsManager) GetDurationAt(thres []int) []time.Duration {
	percentiles := []time.Duration{}
	keySet := []time.Duration{}
	for key := range m.histogram {
		keySet = append(keySet, key)
	}
	sort.Slice(keySet, func(i, j int) bool { return keySet[i] < keySet[j] })
	for _, thre := range thres {
		targetCount := thre * m.numRecords / 100
		count := 0
		percentile := time.Duration(-1)
		for _, key := range keySet {
			count += m.histogram[key]
			if count >= targetCount {
				percentile = key
				break
			}
		}
		percentiles = append(percentiles, percentile)
	}
	return percentiles
}

func (m *HttpMetricsManager) GetMeanDuration() time.Duration {
	return time.Duration(int(m.durationSum) / m.numRecords)
}

// Data for each response shoud be periodically dump to disk
func (m *HttpMetricsManager) DumpToDisk() {
	// TODO:
}

func NewHttpMetricsManager() *HttpMetricsManager {
	manager := HttpMetricsManager{}
	manager.requestSent = 0
	manager.durationSum = 0
	manager.histogram = make(map[time.Duration]int)
	manager.numErrors = 0
	manager.numRecords = 0

	manager.countQueue = make(chan int, 10)
	manager.requestQueue = make(chan core.ChukonuRequest, 10)
	manager.responseQueue = make(chan core.ChukonuResponse, 10)
	manager.errorQueue = make(chan error, 10)
	return &manager
}

func (m *HttpMetricsManager) MeasureThroughput() {
	for count := range m.countQueue {
		m.requestSent = m.requestSent + count
	}
}

func (m *HttpMetricsManager) StartRecording() (chan int, chan core.ChukonuRequest, chan core.ChukonuResponse, chan error) {
	go m.MeasureThroughput()
	go m.SampleMetrics()
	go func() {
		for request := range m.requestQueue {
			m.RecordRequest(request)
		}
	}()
	go func() {
		for response := range m.responseQueue {
			m.RecordResponse(response)
		}
	}()
	go func() {
		for error := range m.errorQueue {
			m.RecordError(error)
		}
	}()
	return m.GetThroughputQueue(), m.GetRequestQueue(), m.GetResponseQueue(), m.GetErrorQueue()
}

// TODO: debounce instead of fixed ticking, because some metrics collection may take longer than 1 sec
func (m *HttpMetricsManager) SampleMetrics() {
	batchStartTime := time.Now()
	deltaStartTime := batchStartTime
	deltaTick := time.Tick(1 * time.Second)
	batchTick := time.Tick(10 * time.Second)
	prevRequestCount := 0
	for {
		select {
		case <-deltaTick:
			elapsed := time.Since(deltaStartTime)
			deltaRequestCount := m.requestSent - prevRequestCount
			percentiles := m.GetDurationAt([]int{50, 90, 95, 99})
			fmt.Printf("delta: %f, total: %d in %f sec, err rate %f, mean %d, median %d, 90percentile %d, 95percentile %d, 99percentile %d\n", float64(deltaRequestCount)/elapsed.Seconds(), deltaRequestCount, elapsed.Seconds(), float64(m.numErrors)/float64(m.numRecords), m.GetMeanDuration(), percentiles[0], percentiles[1], percentiles[2], percentiles[3])
			deltaStartTime = time.Now()
			prevRequestCount = m.requestSent
		case <-batchTick:
			elapsed := time.Since(batchStartTime)
			fmt.Printf("overall: %f, total: %d in %f sec\n", float64(m.requestSent)/elapsed.Seconds(), m.requestSent, elapsed.Seconds())

			// test printing all response
			// if m.requestSent == 100 {
			// 	sum1 := 0
			// 	sum2 := 0
			// 	num := 0
			// 	for key := range m.histogram {
			// 		fmt.Printf("key: %d, val: %v\n", key, m.histogram[key])
			// 		sum2 += int(key) * m.histogram[key]
			// 		num += m.histogram[key]
			//
			// 	}
			// 	fmt.Printf("sum1: %v, sum2: %v, #: %v \n", sum1, sum2, num)
			// 	return
			// }
		default:
		}

	}
}
