package parallel

import (
	"errors"
	"sync"
)

const mapOutputBufferSize = 1000

// ErrMapSkip is a special error that can be returned from Map function to skip the item.
var ErrMapSkip = errors.New("skip")

// MapChan executes `fn` on each element of `input` channel in several threads.
func MapChan[Input any, Output any](input <-chan Input, fn func(in Input) (Output, error), opts ...MapOption) (<-chan Output, <-chan error) {
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
			limiter.Acquire()
			wg.Add(1)
			go func(item Input) {
				defer func() {
					limiter.Release()
					wg.Done()
				}()

				res, err := fn(item)
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

// MapSlice does the same as MapChan, but works with slices instead of channels in input and output.
func MapSlice[Input any, Output any](input []Input, fn func(in Input) (Output, error), opts ...MapOption) ([]Output, []error) {
	// convert slice to channel
	inputChan := make(chan Input, len(input))
	outputChan, errsChan := MapChan(inputChan, fn, opts...)

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

	return output, errs
}

// MapSliceOrdered does the same as MapSlice, but returns results in the same order as input.
func MapSliceOrdered[Input any, Output any](input []Input, fn func(in Input) (Output, error), opts ...MapOption) ([]Output, []error) {
	ops := parseMapOptions(opts)

	wg := sync.WaitGroup{}

	limiter := NewConcurrencyLimiter(ops.concurrency)

	output := make([]Output, len(input))
	errs := make([]error, len(input))

	for i := range input {
		limiter.Acquire()
		wg.Add(1)
		go func(i int) {
			defer func() {
				limiter.Release()
				wg.Done()
			}()

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
		}(i)
	}

	wg.Wait()

	return output, errs
}
