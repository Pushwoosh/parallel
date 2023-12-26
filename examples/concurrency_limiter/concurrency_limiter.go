package main

import (
	"fmt"
	"time"

	"github.com/pushwoosh/parallel"
)

func main() {
	concurrencyLimiter()
}

func concurrencyLimiter() {
	limiter := parallel.NewConcurrencyLimiter(3)

	longJob := func(i int) {
		time.Sleep(time.Second * 1)
		fmt.Println(i)
	}

	for i := 0; i < 10000; i++ {
		limiter.Acquire()
		go func(i int) {
			defer limiter.Release()
			longJob(i)
		}(i)
	}
}
