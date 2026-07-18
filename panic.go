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
//     calling goroutine after all workers finish, so the caller's own
//     defer/recover (e.g. a gRPC recovery interceptor) can handle it just
//     like a panic in synchronous code;
//   - MapChan is non-blocking, so recovered panics are delivered to the
//     returned errors channel as *PanicError values.
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
	once sync.Once
	err  *PanicError
}

// call runs fn, recovering a panic and recording the first one.
func (c *panicCatcher) call(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			err := &PanicError{Value: r, Stack: debug.Stack()}
			c.once.Do(func() {
				c.err = err
			})
		}
	}()

	fn()
}

// repanic re-raises the first recorded panic in the current goroutine.
// Must be called after all workers have finished (i.e. after wg.Wait()),
// which also guarantees visibility of c.err without extra synchronization.
func (c *panicCatcher) repanic() {
	if c.err != nil {
		panic(c.err)
	}
}
