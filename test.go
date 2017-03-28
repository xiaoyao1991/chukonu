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

func main() {
	httpengine := core.NewHttpEngine(core.ChukonuConfig{RequestTimeout: 5 * time.Second})
	req, err := http.NewRequest("GET", "http://localhost:3000", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := httpengine.RunRequest(core.ChukonuHttpRequest{
		Request: req,
	})

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.Status())
}
