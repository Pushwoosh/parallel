package parallel_test

import (
	"errors"
	"testing"

	"github.com/pushwoosh/parallel"
)

func Test_Execute_CollectsErrorsConcurrently(t *testing.T) {
	const totalCallbacks = 1000

	var cbs []func() error
	for i := 0; i < totalCallbacks; i++ {
		cbs = append(cbs, func() error {
			return errors.New("callback error")
		})
	}

	errs := parallel.Execute(cbs...)
	if len(errs) != totalCallbacks {
		t.Fatalf("expected %d errors, got %d", totalCallbacks, len(errs))
	}
}

func Test_ExecuteOpts_CollectsErrorsConcurrently(t *testing.T) {
	const totalCallbacks = 1000

	var cbs []func() error
	for i := 0; i < totalCallbacks; i++ {
		cbs = append(cbs, func() error {
			return errors.New("callback error")
		})
	}

	errs := parallel.ExecuteOpts(cbs, parallel.WithExecuteConcurrency(50))
	if len(errs) != totalCallbacks {
		t.Fatalf("expected %d errors, got %d", totalCallbacks, len(errs))
	}
}
