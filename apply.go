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
// ApplyChan stops reading new items (so the panic surfaces even if `input`
// is never closed), waits for already-started workers, and re-raises the
// first panic in the calling goroutine as *PanicError.
func ApplyChan[T any](input <-chan T, fn func(in T), opts ...ApplyOption) {
	ops := parseApplyOptions(opts)

	wg := sync.WaitGroup{}
	catcher := newPanicCatcher()

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	// run callback for each input item
loop:
	for {
		select {
		case item, ok := <-input:
			if !ok {
				break loop
			}
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
		case <-catcher.done:
			break loop
		}
	}

	wg.Wait()
	catcher.repanic()
}

// ApplySlice does the same as ApplyChan, but works with slice instead of a channel.
//
// A panic in `fn` does not kill the process: after the first recovered panic
// no new items are started, already-started workers finish, and the first
// panic is re-raised in the calling goroutine as *PanicError.
func ApplySlice[T any](input []T, fn func(in T), opts ...ApplyOption) {
	ops := parseApplyOptions(opts)

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	wg := sync.WaitGroup{}
	catcher := newPanicCatcher()
	for _, item := range input {
		if catcher.caught() {
			break
		}
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
