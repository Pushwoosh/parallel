package parallel

import (
	"sync"
)

// Execute executes multiple callback functions `cbs` in parallel.
//
// A panic in a callback does not kill the process: it is recovered in the
// worker (all callbacks are started immediately, so the others still run),
// and the first panic is re-raised in the calling goroutine as *PanicError
// after all workers finish.
func Execute(cbs ...func() error) []error {
	var errs []error
	var errsMu sync.Mutex
	wg := sync.WaitGroup{}
	catcher := newPanicCatcher()

	for i := range cbs {
		wg.Add(1)
		go func(idx int) {
			defer func() {
				wg.Done()
			}()
			catcher.call(func() {
				if err := cbs[idx](); err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
				}
			})
		}(i)
	}

	wg.Wait()
	catcher.repanic()

	return errs
}

// ExecuteOpts executes slice of callback functions `cbs` with custom options.
//
// A panic in a callback does not kill the process: after the first recovered
// panic no new callbacks are started, already-started ones finish, and the
// first panic is re-raised in the calling goroutine as *PanicError.
func ExecuteOpts(cbs []func() error, opts ...ExecuteOption) []error {
	var errs []error
	var errsMu sync.Mutex
	ops := parseExecuteOptions(opts)

	wg := sync.WaitGroup{}
	catcher := newPanicCatcher()

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	for i := range cbs {
		if catcher.caught() {
			break
		}
		limiter.Acquire()
		wg.Add(1)
		go func(idx int) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()
			catcher.call(func() {
				if err := cbs[idx](); err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
				}
			})
		}(i)
	}

	wg.Wait()
	catcher.repanic()

	return errs
}
