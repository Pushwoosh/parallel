package parallel

import (
	"sync"
)

// ApplyChan executes `fn` on each element of `input` channel in multiple threads.
// Options:
//
//	WithApplyConcurrency(int) - limits the number of parallel threads. Default: ApplyDefaultConcurrency
//
// To stop processing, close `input` channel.
//
// A panic in `fn` does not kill the process: it is recovered in the worker,
// remaining items are still processed, and the first panic is re-raised in
// the calling goroutine as *PanicError after all workers finish.
func ApplyChan[T any](input <-chan T, fn func(in T), opts ...ApplyOption) {
	ops := parseApplyOptions(opts)

	wg := sync.WaitGroup{}
	catcher := panicCatcher{}

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	// run callback for each input item
	for item := range input {
		limiter.Acquire()
		wg.Add(1)
		go func(item T) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()

			catcher.call(func() {
				fn(item)
			})
		}(item)
	}

	wg.Wait()
	catcher.repanic()
}

// ApplySlice does the same as ApplyChan, but works with slice instead of a channel.
func ApplySlice[T any](input []T, fn func(in T), opts ...ApplyOption) {
	ops := parseApplyOptions(opts)

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	wg := sync.WaitGroup{}
	catcher := panicCatcher{}
	for _, item := range input {
		limiter.Acquire()
		wg.Add(1)
		go func(item T) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()

			catcher.call(func() {
				fn(item)
			})
		}(item)
	}
	wg.Wait()
	catcher.repanic()
}
