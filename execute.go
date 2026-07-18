package parallel

import (
	"sync"
)

// Execute executes multiple callback functions `cbs` in parallel.
//
// A panic in a callback does not kill the process: it is recovered in the
// worker, remaining callbacks are still executed, and the first panic is
// re-raised in the calling goroutine as *PanicError after all workers finish.
func Execute(cbs ...func() error) []error {
	var errs []error
	wg := sync.WaitGroup{}
	catcher := panicCatcher{}

	for i := range cbs {
		wg.Add(1)
		go func(idx int) {
			defer func() {
				wg.Done()
			}()
			catcher.call(func() {
				if err := cbs[idx](); err != nil {
					errs = append(errs, err)
				}
			})
		}(i)
	}

	wg.Wait()
	catcher.repanic()

	return errs
}

// ExecuteOpts executes slice of callback functions `cbs` with custom options.
func ExecuteOpts(cbs []func() error, opts ...ExecuteOption) []error {
	var errs []error
	ops := parseExecuteOptions(opts)

	wg := sync.WaitGroup{}
	catcher := panicCatcher{}

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	for i := range cbs {
		limiter.Acquire()
		wg.Add(1)
		go func(idx int) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()
			catcher.call(func() {
				if err := cbs[idx](); err != nil {
					errs = append(errs, err)
				}
			})
		}(i)
	}

	wg.Wait()
	catcher.repanic()

	return errs
}
