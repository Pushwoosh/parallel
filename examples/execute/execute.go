package main

import (
	"errors"
	"fmt"

	"github.com/pushwoosh/parallel"
)

func main() {
	execute()
}

func execute() {
	test := func(value string) error {
		if value == "b" {
			return errors.New("test error")
		}

		fmt.Printf("execute test func '%s'\n", value)
		return nil
	}

	_ = parallel.Execute(
		func() error { return test("a") },
		func() error { return test("b") },
		func() error { return test("c") },
	)
}
