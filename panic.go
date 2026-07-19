package parallel

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// PanicError wraps a panic recovered in a worker goroutine.
//
// A panic in a goroutine spawned by this package can not be recovered by the
// caller: recover() only works in the goroutine where the panic happened, and
// an unrecovered panic in any goroutine kills the whole process. To prevent
// that, every worker goroutine recovers panics itself and the package
// propagates them to the caller:
//
//   - blocking functions (ApplyChan, ApplySlice, Execute, ExecuteOpts,
//     MapSlice, MapSliceOrdered) re-raise the first recovered panic in the
//     calling goroutine after the workers finish, so the caller's own
//     defer/recover (e.g. a gRPC recovery interceptor) can handle it just
//     like a panic in synchronous code. After the first panic no new items
//     or callbacks are started; workers that are already running finish
//     first. ApplyChan also stops reading from its input channel, so the
//     panic is re-raised even if the channel is never closed;
//   - MapChan is non-blocking, so recovered panics are delivered to the
//     returned errors channel as *PanicError values and processing
//     continues.
//
// Value holds the original value passed to panic(), Stack holds the stack
// trace of the worker goroutine captured at the moment of recovery.
type PanicError struct {
	Value any
	Stack []byte
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic in parallel worker: %v\n%s", e.Value, e.Stack)
}

// Unwrap returns the panic value if it is an error, so errors.Is and
// errors.As see through PanicError.
func (e *PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

// panicCatcher records the first panic recovered in worker goroutines.
type panicCatcher struct {
	mu sync.Mutex
	// err is the first recorded panic
	err *PanicError
	// done is closed when the first panic is recorded, so dispatch loops
	// blocked on a channel receive can stop waiting for further input
	done chan struct{}
}

func newPanicCatcher() *panicCatcher {
	return &panicCatcher{done: make(chan struct{})}
}

// record keeps the first recorded panic and signals done.
func (c *panicCatcher) record(err *PanicError) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err == nil {
		c.err = err
		close(c.done)
	}
}

// call runs fn, recovering a panic and recording the first one.
//
// Completion is tracked with a flag instead of checking recover()'s result:
// under the module's `go 1.18` semantics recover() returns nil after
// panic(nil), which would otherwise make such a panic vanish. A worker that
// exits via runtime.Goexit is treated the same as panic(nil).
func (c *panicCatcher) call(fn func()) {
	completed := false
	defer func() {
		if !completed {
			c.record(&PanicError{Value: recover(), Stack: debug.Stack()})
		}
	}()

	fn()
	completed = true
}

// caught reports whether a panic has been recorded.
func (c *panicCatcher) caught() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err != nil
}

// repanic re-raises the first recorded panic in the current goroutine.
// Call it after all workers have finished (i.e. after wg.Wait()).
func (c *panicCatcher) repanic() {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()

	if err != nil {
		panic(err)
	}
}
