package other

<<<<<<< Updated upstream
import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/xiaoyao1991/chukonu/core"
	"github.com/xiaoyao1991/chukonu/impl"
)

// func consume(id int, queue chan int) {
=======
// import (
// 	"fmt"
// 	"net/http"
// 	"time"
//
// 	"github.com/xiaoyao1991/chukonu/core"
// 	"github.com/xiaoyao1991/chukonu/impl"
// )
//
// // func consume(id int, queue chan int) {
// // 	for {
// // 		select {
// // 		case j := <-queue:
// // 			fmt.Println(fmt.Sprintf("consumer#%d: %d", id, j))
// // 		}
// // 	}
// // }
// //
// // func main() {
// // 	fmt.Println("Hello, playground")
// //
// // 	q := make(chan int)
// // 	throttle := time.Tick(1 * time.Second)
// //
// // 	for i := 1; i <= 5; i++ {
// // 		go consume(i, q)
// // 	}
// //
// // 	counter := 0
// // 	for {
// // 		<-throttle
// // 		q <- counter
// // 		counter++
// // 	}
// // }
//
// type MyRequestProvider struct {
// 	queue chan core.ChukonuRequest
// }
//
// func (m *MyRequestProvider) Provide() chan core.ChukonuRequest {
// 	if m.queue == nil {
// 		m.queue = make(chan core.ChukonuRequest, 5)
// 	}
//
// 	return m.queue
// }
//
// func (m *MyRequestProvider) Gen() {
// 	queue := m.Provide()
//
// 	throttle := time.Tick(200 * time.Millisecond)
// 	i := 0
>>>>>>> Stashed changes
// 	for {
// 		<-throttle
// 		// fmt.Printf("Generating %dth request\n", i)
// 		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:3000/%d", i), nil)
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		queue <- impl.ChukonuHttpRequest{Request: req}
// 		i++
//
// 		if i == 100 {
// 			break
// 		}
// 	}
//
// 	close(queue)
// }
//
// func main() {
// 	config := core.ChukonuConfig{Concurrency: 5, RequestTimeout: 5 * time.Second}
// 	httpengine := impl.NewHttpEngine(config)
// 	var pool core.Pool
//
// 	// rawResp := resp.RawResponse().(*http.Response)
// 	// defer rawResp.Body.Close()
// 	// bodyBytes, err := ioutil.ReadAll(rawResp.Body)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// fmt.Println(string(bodyBytes))
//
// 	provider := &MyRequestProvider{}
// 	go provider.Gen()
//
// 	//add response
// 	pool.Start(httpengine, provider, impl.NewHttpMetricsManager(), config)
// }
<<<<<<< Updated upstream

type MyRequestProvider struct {
	queue chan core.ChukonuRequest
}

func (m *MyRequestProvider) Provide(queue chan *core.ChukonuWorkflow) {
	// throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		// <-throttle
		// fmt.Printf("Generating %dth request\n", i)
		workflow := core.NewWorkflow("test_workflow")

		var fn1 = func(ctx context.Context) core.ChukonuRequest {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:3000/%d", i), nil)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			return impl.NewChukonuHttpRequest("datasources_req", 0, true, true, func(ctx context.Context, resp core.ChukonuResponse) context.Context {
				defer resp.RawResponse().(*http.Response).Body.Close()
				return ctx
			}, nil, req)
		}

		workflow.AddRequest(fn1)
		queue <- workflow
		i++

		// if i == 100 {
		// 	break
		// }
	}

	// close(queue)
}

func (m *MyRequestProvider) MetricsManager() core.MetricsManager {
	return impl.NewHttpMetricsManager()
}

func (m *MyRequestProvider) Gen() {
	queue := m.Provide()

	throttle := time.Tick(200 * time.Millisecond)
	i := 0
	for {
		<-throttle
		// fmt.Printf("Generating %dth request\n", i)
		req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:3000/%d", i), nil)
		if err != nil {
			fmt.Println(err)
		}
		queue <- impl.ChukonuHttpRequest{Request: req}
		i++

		if i == 100 {
			break
		}
	}

	close(queue)
}

func main() {

	config := core.ChukonuConfig{Concurrency: 5, RequestTimeout: 5 * time.Minute}
	var engines []core.Engine = make([]core.Engine, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		engines[i] = impl.NewHttpEngine(config)
	}
	var pool core.Pool
	provider := &MyRequestProvider{}
	go provider.Gen()
	fuse := make(chan bool, 1)
	ack := make(chan bool, 1)

	pool.Start(engines, provider, requestProvider.MetricsManager(), config, fuse, ack)
	// pool.Start(engines, requestProvider, requestProvider.MetricsManager(), config, fuse, ack)

}
=======
>>>>>>> Stashed changes
