package main

import (
	"fmt"

	"github.com/pushwoosh/parallel"
)

func main() {
	applyChan()
	applySlice()
}

func applyChan() {
	fmt.Println("applyChan:")
	input := make(chan int, 5)
	go func() {
		for i := 1; i <= 20; i++ {
			input <- i
		}
		close(input)
	}()

	parallel.ApplyChan(input, func(in int) {
		fmt.Printf("%d * %d = %d\n", in, in, in*in)
	})
}

func applySlice() {
	fmt.Println("applySlice:")
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	parallel.ApplySlice(input, func(in int) {
		fmt.Printf("%d * %d = %d\n", in, in, in*in)
	})
}
