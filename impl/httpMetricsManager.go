package impl

import (
	"fmt"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
)

type HttpMetricsManager struct {
	requestSent int
	countQueue  chan int
	// throughputRecord []int
}

func (m *HttpMetricsManager) RecordRequest(request core.ChukonuRequest) {
}
func (m *HttpMetricsManager) RecordResponse(response core.ChukonuResponse) {
}
func (m *HttpMetricsManager) RecordError(err error) {
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

func NewHttpMetricsManager() *HttpMetricsManager {
	manager := new(HttpMetricsManager)
	manager.requestSent = 0
	manager.countQueue = make(chan int, 10)
	// manager.throughputRecord = make([]int, 0, 3600)
	return manager
}

func (m *HttpMetricsManager) MeasureThroughput() {
	// startTime := time.Now()
	for count := range m.countQueue {
		m.requestSent = m.requestSent + count
	}
}

func (m *HttpMetricsManager) SampleThroughput() {
	// preRequestSent := m.requestSent
	// windowRequestSent := 0
	startTime := time.Now()
	tick := time.Tick(1 * time.Second)
	for {
		<-tick
		// windowRequestSent = m.requestSent - preRequestSent
		// preRequestSent = m.requestSent
		elapseTime := time.Since(startTime)
		// fmt.Println("Throughput/sec:	", windowRequestSent)
		fmt.Println("cumulative	", float64(m.requestSent)/elapseTime.Seconds(), "total ", m.requestSent)
	}
}
