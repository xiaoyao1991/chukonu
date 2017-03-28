package core

import (
	"fmt"
	"sync"
)

type Pool struct {
}

func (p Pool) Start(engine Engine, provider RequestProvider, config ChukonuConfig) {
	var wg sync.WaitGroup
	wg.Add(config.Concurrency)

	queue := provider.Provide()
	for i := 0; i < config.Concurrency; i++ {
		go func(i int) {
			defer wg.Done()
			for req := range queue {
				fmt.Println(fmt.Sprintf("goroutine %d running request...", i))
				resp, err := engine.RunRequest(req)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("\t" + resp.Status())
			}
		}(i)
	}

	wg.Wait()
}
