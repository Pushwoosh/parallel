package parallel

import (
	"errors"
	"runtime/debug"
	"sync"
)

const mapOutputBufferSize = 1000

// ErrMapSkip is a special error that can be returned from Map function to skip the item.
var ErrMapSkip = errors.New("skip")

// MapChan executes `fn` on each element of `input` channel in several threads.
//
// A panic in `fn` does not kill the process: it is recovered in the worker
// and delivered to the returned errors channel as *PanicError. MapChan is
// non-blocking, so the panic can not be re-raised in the calling goroutine.
func MapChan[Input any, Output any](input <-chan Input, fn func(in Input) (Output, error), opts ...MapOption) (<-chan Output, <-chan error) {
	return mapChan(input, fn, nil, opts)
}

// mapChan implements MapChan and MapSlice. If `catcher` is nil, recovered
// worker panics are delivered to the returned errors channel; otherwise they
// are recorded in `catcher` (for the caller to re-raise) and no new items
// are dispatched after the first one.
func mapChan[Input any, Output any](input <-chan Input, fn func(in Input) (Output, error), catcher *panicCatcher, opts []MapOption) (<-chan Output, <-chan error) {
	ops := parseMapOptions(opts)

	wg := sync.WaitGroup{}

	// init goroutines limiter
	limiter := NewConcurrencyLimiter(ops.concurrency)

	output := make(chan Output, mapOutputBufferSize)
	errs := make(chan error, mapOutputBufferSize)
	go func() {
		defer func() {
			close(output)
			close(errs)
		}()

		// run callback for each input item
		for item := range input {
			if catcher != nil && catcher.caught() {
				break
			}
			limiter.Acquire()
			wg.Add(1)
			go func(item Input) {
				defer func() {
					limiter.Release()
					wg.Done()
				}()

				res, err, panicErr := protectedCall(fn, item)
				if panicErr != nil {
					if catcher != nil {
						catcher.record(panicErr)
					} else {
						errs <- panicErr
					}
					return
				}

				if err != nil {
					// We do not care if some inner function returns ErrMapSkip.
					// Only direct error from `fn` callback is important. So no errors.Is here
					//goland:noinspection GoDirectComparisonOfErrors
					if err != ErrMapSkip {
						errs <- err
					}
					return
				}

				output <- res
			}(item)
		}

		wg.Wait()
	}()

	return output, errs
}

// protectedCall invokes fn(item), converting a panic into *PanicError.
// Completion is tracked with a flag instead of checking recover()'s result,
// so panic(nil) is detected too (see panicCatcher.call).
func protectedCall[Input any, Output any](fn func(in Input) (Output, error), item Input) (res Output, err error, panicErr *PanicError) {
	completed := false
	defer func() {
		if !completed {
			panicErr = &PanicError{Value: recover(), Stack: debug.Stack()}
		}
	}()

	res, err = fn(item)
	completed = true
	return
}

// MapSlice does the same as MapChan, but works with slices instead of channels in input and output.
//
// Unlike MapChan, MapSlice blocks until the workers finish, so the first
// panic recovered in a worker is re-raised in the calling goroutine
// as *PanicError. After the first panic no new items are started.
func MapSlice[Input any, Output any](input []Input, fn func(in Input) (Output, error), opts ...MapOption) ([]Output, []error) {
	catcher := newPanicCatcher()

	// convert slice to channel
	inputChan := make(chan Input, len(input))
	outputChan, errsChan := mapChan(inputChan, fn, catcher, opts)

	go func() {
		for _, item := range input {
			inputChan <- item
		}
		close(inputChan)
	}()

	// collect results from the channels to output slices
	var output []Output
	var errs []error
	var outputClosed, errsClosed bool
	for {
		if outputClosed && errsClosed {
			break
		}
		select {
		case item, more := <-outputChan:
			if !more {
				outputClosed = true
				continue
			}
			output = append(output, item)
		case err, more := <-errsChan:
			if !more {
				errsClosed = true
				continue
			}
			errs = append(errs, err)
		}
	}

	// both channels are closed, so all workers have finished
	catcher.repanic()

	return output, errs
}

// MapSliceOrdered does the same as MapSlice, but returns results in the same order as input.
func MapSliceOrdered[Input any, Output any](input []Input, fn func(in Input) (Output, error), opts ...MapOption) ([]Output, []error) {
	ops := parseMapOptions(opts)

	wg := sync.WaitGroup{}
	catcher := newPanicCatcher()

	limiter := NewConcurrencyLimiter(ops.concurrency)

	output := make([]Output, len(input))
	errs := make([]error, len(input))

	for i := range input {
		if catcher.caught() {
			break
		}
		limiter.Acquire()
		wg.Add(1)
		go func(i int) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()

			catcher.call(func() {
				res, err := fn(input[i])

				// We do not care if some inner function returns ErrMapSkip.
				// Only direct error from `fn` callback is important. So no errors.Is here
				//goland:noinspection GoDirectComparisonOfErrors
				if err != ErrMapSkip {
					errs[i] = err
				}

				if err == nil {
					output[i] = res
				}
			})
		}(i)
	}

	wg.Wait()
	catcher.repanic()

	return output, errs
}
