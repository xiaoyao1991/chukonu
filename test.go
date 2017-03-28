package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
)

// func consume(id int, queue chan int) {
// 	for {
// 		select {
// 		case j := <-queue:
// 			fmt.Println(fmt.Sprintf("consumer#%d: %d", id, j))
// 		}
// 	}
// }
//
// func main() {
// 	fmt.Println("Hello, playground")
//
// 	q := make(chan int)
// 	throttle := time.Tick(1 * time.Second)
//
// 	for i := 1; i <= 5; i++ {
// 		go consume(i, q)
// 	}
//
// 	counter := 0
// 	for {
// 		<-throttle
// 		q <- counter
// 		counter++
// 	}
// }

type MyRequestProvider struct {
	queue chan core.ChukonuRequest
}

func (m *MyRequestProvider) Provide() chan core.ChukonuRequest {
	if m.queue == nil {
		m.queue = make(chan core.ChukonuRequest, 5)
	}

	return m.queue
}

func (m *MyRequestProvider) Gen() {
	queue := m.Provide()

	throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		<-throttle
		fmt.Printf("Generating %dth request\n", i)
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:3000/%d", i), nil)
		if err != nil {
			fmt.Println(err)
		}
		queue <- core.ChukonuHttpRequest{Request: req}
		i++
	}
	// for i := 0; i < 100; i++ {
	// 	fmt.Printf("Generating %dth request\n", i)
	// 	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:3000/%d", i), nil)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	queue <- core.ChukonuHttpRequest{Request: req}
	// }
	// close(queue)
}

func main() {
	config := core.ChukonuConfig{Concurrency: 5, RequestTimeout: 5 * time.Second}
	httpengine := core.NewHttpEngine(config)
	var pool core.Pool

	// rawResp := resp.RawResponse().(*http.Response)
	// defer rawResp.Body.Close()
	// bodyBytes, err := ioutil.ReadAll(rawResp.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(string(bodyBytes))

	provider := &MyRequestProvider{}
	go provider.Gen()
	pool.Start(httpengine, provider, config)
}
