package impl

import (
	"fmt"
	"sort"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
)

type HttpMetricsManager struct {
	requestSent int
	countQueue  chan int
	// durationSum DurationSummation
	numErrors  int
	numRecords int
	// records     HttpRecords
	histogram map[time.Duration]int
}

// type HttpRecords []ChukonuHttpResponse

// type DurationSummation struct {
// 	sum   time.Duration
// 	index int
// }

// ------------sorting interface for struct HttpRecord-----------
// func (slice HttpRecords) Len() int {
// 	return len(slice)
// }
// func (slice HttpRecords) Less(i, j int) bool {
// 	return slice[i].duration < slice[j].duration
// }
// func (slice HttpRecords) Swap(i, j int) {
// 	slice[i], slice[j] = slice[j], slice[i]
// }

//----------------------------------------------------------------

func (m *HttpMetricsManager) RecordRequest(request core.ChukonuRequest) {
}
func (m *HttpMetricsManager) RecordResponse(response core.ChukonuResponse) {
	if response, ok := response.(ChukonuHttpResponse); ok {
		// m.records = append(m.records, response)
		m.histogram[response.Duration()/time.Millisecond] += 1
		m.numRecords += 1
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
func (m *HttpMetricsManager) GetQueue() chan int {
	return m.countQueue
}
func (m *HttpMetricsManager) GetThroughput() int {
	return m.requestSent
}

// return the sampletime(duration) at input percentile
func (m *HttpMetricsManager) GetDurationAt(thre int) time.Duration {
	targetCount := thre * m.numRecords / 100
	count := 0
	keySet := []time.Duration{}
	for key := range m.histogram {
		keySet = append(keySet, key)
	}
	sort.Slice(keySet, func(i, j int) bool { return keySet[i] < keySet[j] })
	for _, key := range keySet {
		count += m.histogram[key]
		if count >= targetCount {
			return key
		}
	}
	return -1
}

func (m *HttpMetricsManager) GetMeanDuration() time.Duration {
	// Approach using histogram
	sum := 0
	numRecords := 0
	for key := range m.histogram {
		sum += int(key) * m.histogram[key]
		numRecords += m.histogram[key]
	}
	return time.Duration(sum / numRecords)
}

// Data for each response shoud be periodically dump to disk
func (m *HttpMetricsManager) DumpToDisk() {
	// TODO:
}

func NewHttpMetricsManager() *HttpMetricsManager {
	manager := HttpMetricsManager{}
	manager.requestSent = 0
	manager.countQueue = make(chan int, 10)
	// manager.records = []ChukonuHttpResponse{}
	manager.histogram = make(map[time.Duration]int)
	// manager.durationSum = DurationSummation{}
	manager.numErrors = 0
	manager.numRecords = 0
	// manager.throughputRecord = make([]int, 0, 3600)
	return &manager
}

func (m *HttpMetricsManager) MeasureThroughput() {
	for count := range m.countQueue {
		m.requestSent = m.requestSent + count
	}
}

// TODO: debounce instead of fixed ticking, because some metrics collection may take longer than 1 sec
func (m *HttpMetricsManager) SampleMetrics() {
	batchStartTime := time.Now()
	deltaStartTime := batchStartTime
	deltaTick := time.Tick(1 * time.Second)
	batchTick := time.Tick(2 * time.Second)
	prevRequestCount := 0
	for {
		select {
		case <-deltaTick:
			elapsed := time.Since(deltaStartTime)
			deltaRequestCount := m.requestSent - prevRequestCount
			fmt.Printf("delta: %f, total: %d in %f sec, err rate %f, mean %d, median %d, 90percentile %d, 95percentile %d, 99percentile %d\n", float64(deltaRequestCount)/elapsed.Seconds(), deltaRequestCount, elapsed.Seconds(), float64(m.numErrors)/float64(m.numRecords), m.GetMeanDuration(), m.GetDurationAt(50), m.GetDurationAt(90), m.GetDurationAt(95), m.GetDurationAt(99))
			deltaStartTime = time.Now()
			prevRequestCount = m.requestSent
		case <-batchTick:
			elapsed := time.Since(batchStartTime)
			fmt.Printf("overall: %f, total: %d in %f sec\n", float64(m.requestSent)/elapsed.Seconds(), m.requestSent, elapsed.Seconds())

			//test printing all response
			// if m.requestSent == 100 {
			// 	sum1 := 0
			// 	sum2 := 0
			// 	num := 0
			// 	// for i := 0; i < 100; i++ {
			// 	// 	fmt.Printf("sample time: %d, status: %v\n", m.records[i].Duration()/time.Millisecond, m.records[i].Status())
			// 	// 	sum1 += int(m.records[i].Duration() / time.Millisecond)
			// 	// }
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
