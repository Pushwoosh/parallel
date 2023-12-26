package parallel

import (
	"errors"
	"sync"
	"testing"

	"gitlab.corp.pushwoosh.com/uds/pkg/parallel"
)

func Test_MapChan(t *testing.T) {
	const inputBufferSize = 10
	const totalNumbers = 100000

	input := make(chan int, inputBufferSize)
	go func() {
		for i := 0; i < totalNumbers; i++ {
			input <- i
		}
		close(input)
	}()

	evenIntegers, errs := parallel.MapChan(input, filterEvenIntegers)
	wg := sync.WaitGroup{}
	wg.Add(2)

	var resultsCount int
	var errorsCount int

	go func() {
		defer wg.Done()
		for range errs {
			errorsCount++
		}
	}()

	go func() {
		defer wg.Done()
		for range evenIntegers {
			resultsCount++
		}
	}()

	wg.Wait()

	if resultsCount != totalNumbers/2 {
		t.Fatalf("expected %d results, got %d", totalNumbers/2, resultsCount)
	}

	if errorsCount != 1 {
		t.Fatalf("expected %d errors, got %d", 1, errorsCount)
	}
}

func Test_MapSlice(t *testing.T) {
	const totalNumbers = 100000

	var input []int
	for i := 0; i < totalNumbers; i++ {
		input = append(input, i)
	}

	results, errs := parallel.MapSlice(input, filterEvenIntegers)

	if len(results) != totalNumbers/2 {
		t.Fatalf("expected %d results, got %d", totalNumbers/2, len(results))
	}

	if len(errs) != 1 {
		t.Fatalf("expected %d errors, got %d", 1, len(errs))
	}
}

func mul2(i int) (int, error) {
	if i == 13 {
		return 0, errors.New("processing error")
	}
	return i * 2, nil
}

func Test_MapSliceOrdered(t *testing.T) {
	const totalNumbers = 100000

	var input []int
	for i := 0; i < totalNumbers; i++ {
		input = append(input, i)
	}

	results, errs := parallel.MapSliceOrdered(input, mul2)
	for i := range results {
		if i == 13 {
			if errs[i] == nil {
				t.Fatalf("expected error, got nil")
			}
			continue
		}

		if errs[i] != nil {
			t.Fatalf("expected nil error, got %v", errs[i])
		}

		if results[i] != i*2 {
			t.Fatalf("expected %d, got %d", i*2, results[i])
		}
	}
}

func filterEvenIntegers(i int) (int, error) {
	if i%2 == 0 {
		return i, nil
	}

	if i == 13 {
		return 0, errors.New("processing error")
	}

	return i, parallel.ErrMapSkip
}
