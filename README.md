# Parallel
Parallel is a go library containing helper functions for parallel processing.

# Installation
```bash
go get github.com/Pushwoosh/parallel
```

# Examples
See [examples](examples) directory for full examples.

# Apply
Apply executes given function on each element of the input slice or channel.

There are two versions of Apply: `ApplySlice` and `ApplyChan`:
```go
func ApplyChan[T any](input <-chan T, fn func(in T), opts ...ApplyOption) {}
func ApplySlice[T any](input []T, fn func(in T), opts ...ApplyOption) {}
```

To stop processing you can close the input channel.

## Options

- `WithApplyConcurrency(int)` - limits the number of parallel threads. Default: `ApplyDefaultConcurrency`.

## Example
```go
ch := make(chan int)
...
ApplyChan(ch, func(in int) {
    fmt.Println(in)
})
```

# Map

Map executes given function on each element of the input slice or channel and returns the result.

There are two versions of Map: `MapSlice` and `MapChan`:
```go
func MapChan[Input any, Output any](input <-chan Input, fn func(in Input) (Output, error), opts ...MapOption) (<-chan Output, <-chan error) {}
func MapSlice[Input any, Output any](input []Input, fn func(in Input) (Output, error), opts ...MapOption) ([]Output, []error) {}
```

To stop processing you can close the input channel.

## Options

- `WithMapConcurrency(int)` - limits the number of parallel threads. Default: `MapDefaultConcurrency`.
- `WithMapStopOnFirstError` - forces executor to stop processing new items after the first error occurred.

## Example
```go
input := make(chan string, 100)
input <- "hello"
input <- "world"
close(input)

parallel.MapChan(input, func(s string) (string, error) { return strings.ToTitle(s), nil })
```

## Flow control
Output channel is filled with results of the callback function by the following rules:
- If error is nil, result is sent to output channel.
- If error is ErrMapSkip, result is not sent to output channel.
- If error is not ErrMapSkip, result is not sent to output channel and error is sent to errors output channel.

Same rules apply to `MapSlice` function.

# MapSliceOrdered
MapSliceOrdered is a special version of MapSlice that guarantees that output
slice will contain results in the same order as input slice.

Output slice will always contain the same number of elements as input slice. If callback function
returns an error, output slice will contain `nil` at the corresponding position.

# Execute

Execute executes given functions in parallel. There is no limit on the number of parallel threads. All given
functions will be started at the same time.

To limit concurrency use `ExecuteOpts`.

## Options
- `WithExecuteConcurrency(int)` - limits the number of parallel threads. Default: `ExecuteDefaultConcurrency`.

## Example
```go
parallel.Execute(
    func() { fmt.Println("Hello") },
    func() { fmt.Println("World") },
)
```

# ConcurrencyLimiter
