package main

import (
	"errors"
	"fmt"

	"github.com/Pushwoosh/parallel"
)

func main() {
	mapSlice()
	mapChan()
}

func mapSlice() {
	fmt.Println("mapSlice:")
	var input []int
	for i := 0; i < 100; i++ {
		input = append(input, i)
	}

	results, errs := parallel.MapSlice(input, filterEvenIntegers)

	for _, out := range results {
		fmt.Println("output: ", out)
	}

	for _, err := range errs {
		fmt.Println("error: ", err)
	}
}

func mapChan() {
	fmt.Println("mapChan:")
	input := make(chan int, 100)
	for i := 0; i < 100; i++ {
		input <- i
	}
	close(input)

	results, errs := parallel.MapChan(input, filterEvenIntegers)

	for out := range results {
		fmt.Println("output: ", out)
	}

	for err := range errs {
		fmt.Println("error: ", err)
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
