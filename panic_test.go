package parallel_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/pushwoosh/parallel"
)

var errBoom = errors.New("boom")

// recoverPanicError runs fn and returns the *parallel.PanicError recovered
// from it. Fails the test if fn does not panic with *parallel.PanicError.
func recoverPanicError(t *testing.T, fn func()) *parallel.PanicError {
	t.Helper()

	var panicErr *parallel.PanicError
	func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic, got none")
			}
			var ok bool
			panicErr, ok = r.(*parallel.PanicError)
			if !ok {
				t.Fatalf("expected *parallel.PanicError, got %T: %v", r, r)
			}
		}()
		fn()
	}()

	return panicErr
}

func Test_ApplySlice_PanicPropagates(t *testing.T) {
	const totalNumbers = 1000

	var input []int
	for i := 0; i < totalNumbers; i++ {
		input = append(input, i)
	}

	var processed int64
	panicErr := recoverPanicError(t, func() {
		parallel.ApplySlice(input, func(i int) {
			if i == 13 {
				panic(errBoom)
			}
			atomic.AddInt64(&processed, 1)
		})
	})

	if !errors.Is(panicErr, errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErr.Value)
	}
	if len(panicErr.Stack) == 0 {
		t.Fatal("expected non-empty worker stack trace")
	}
	// all workers must have finished before the panic is re-raised
	if got := atomic.LoadInt64(&processed); got != totalNumbers-1 {
		t.Fatalf("expected %d processed items, got %d", totalNumbers-1, got)
	}
}

func Test_ApplyChan_PanicPropagates(t *testing.T) {
	const totalNumbers = 1000

	input := make(chan int, totalNumbers)
	for i := 0; i < totalNumbers; i++ {
		input <- i
	}
	close(input)

	var processed int64
	panicErr := recoverPanicError(t, func() {
		parallel.ApplyChan(input, func(i int) {
			if i == 13 {
				panic("string panic")
			}
			atomic.AddInt64(&processed, 1)
		})
	})

	if panicErr.Value != "string panic" {
		t.Fatalf("expected panic value %q, got %v", "string panic", panicErr.Value)
	}
	if got := atomic.LoadInt64(&processed); got != totalNumbers-1 {
		t.Fatalf("expected %d processed items, got %d", totalNumbers-1, got)
	}
}

func Test_Execute_PanicPropagates(t *testing.T) {
	var processed int64
	panicErr := recoverPanicError(t, func() {
		parallel.Execute(
			func() error { atomic.AddInt64(&processed, 1); return nil },
			func() error { panic(errBoom) },
			func() error { atomic.AddInt64(&processed, 1); return nil },
		)
	})

	if !errors.Is(panicErr, errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErr.Value)
	}
	if got := atomic.LoadInt64(&processed); got != 2 {
		t.Fatalf("expected 2 processed callbacks, got %d", got)
	}
}

func Test_ExecuteOpts_PanicPropagates(t *testing.T) {
	var cbs []func() error
	var processed int64
	for i := 0; i < 100; i++ {
		i := i
		cbs = append(cbs, func() error {
			if i == 13 {
				panic(errBoom)
			}
			atomic.AddInt64(&processed, 1)
			return nil
		})
	}

	panicErr := recoverPanicError(t, func() {
		parallel.ExecuteOpts(cbs, parallel.WithExecuteConcurrency(5))
	})

	if !errors.Is(panicErr, errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErr.Value)
	}
	if got := atomic.LoadInt64(&processed); got != 99 {
		t.Fatalf("expected 99 processed callbacks, got %d", got)
	}
}

func Test_MapSlice_PanicPropagates(t *testing.T) {
	const totalNumbers = 1000

	var input []int
	for i := 0; i < totalNumbers; i++ {
		input = append(input, i)
	}

	panicErr := recoverPanicError(t, func() {
		parallel.MapSlice(input, func(i int) (int, error) {
			if i == 13 {
				panic(errBoom)
			}
			return i * 2, nil
		})
	})

	if !errors.Is(panicErr, errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErr.Value)
	}
}

func Test_MapSliceOrdered_PanicPropagates(t *testing.T) {
	const totalNumbers = 1000

	var input []int
	for i := 0; i < totalNumbers; i++ {
		input = append(input, i)
	}

	panicErr := recoverPanicError(t, func() {
		parallel.MapSliceOrdered(input, func(i int) (int, error) {
			if i == 13 {
				panic(errBoom)
			}
			return i * 2, nil
		})
	})

	if !errors.Is(panicErr, errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErr.Value)
	}
}

func Test_MapChan_PanicGoesToErrsChannel(t *testing.T) {
	const totalNumbers = 1000

	input := make(chan int, totalNumbers)
	for i := 0; i < totalNumbers; i++ {
		input <- i
	}
	close(input)

	output, errs := parallel.MapChan(input, func(i int) (int, error) {
		if i == 13 {
			panic(errBoom)
		}
		return i * 2, nil
	})

	var results int
	var panicErrs []*parallel.PanicError
	for output != nil || errs != nil {
		select {
		case _, more := <-output:
			if !more {
				output = nil
				continue
			}
			results++
		case err, more := <-errs:
			if !more {
				errs = nil
				continue
			}
			var panicErr *parallel.PanicError
			if !errors.As(err, &panicErr) {
				t.Fatalf("expected *parallel.PanicError in errs channel, got %T: %v", err, err)
			}
			panicErrs = append(panicErrs, panicErr)
		}
	}

	if results != totalNumbers-1 {
		t.Fatalf("expected %d results, got %d", totalNumbers-1, results)
	}
	if len(panicErrs) != 1 {
		t.Fatalf("expected 1 panic error, got %d", len(panicErrs))
	}
	if !errors.Is(panicErrs[0], errBoom) {
		t.Fatalf("expected panic value to unwrap to errBoom, got %v", panicErrs[0].Value)
	}
}
