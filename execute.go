package parallel

import (
	"sync"
)

// Execute executes multiple callback functions `cbs` in parallel.
func Execute(cbs ...func() error) []error {
	var errs []error
	wg := sync.WaitGroup{}

	for i := range cbs {
		wg.Add(1)
		go func(idx int) {
			defer func() {
				wg.Done()
			}()
			if err := cbs[idx](); err != nil {
				errs = append(errs, err)
			}
		}(i)
	}

	wg.Wait()

	return errs
}

// ExecuteOpts executes slice of callback functions `cbs` with custom options.
func ExecuteOpts(cbs []func() error, opts ...ExecuteOption) []error {
	var errs []error
	ops := parseExecuteOptions(opts)

	wg := sync.WaitGroup{}

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
			if err := cbs[idx](); err != nil {
				errs = append(errs, err)
			}
		}(i)
	}

	wg.Wait()

	return errs
}
