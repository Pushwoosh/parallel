# Parallel
Parallel is a go library containing helper functions for parallel processing.

# Installation
```bash
go get github.com/pushwoosh/parallel
```

# Examples
See [examples](examples) directory for full examples.

# Panic handling
A panic inside a callback happens in a worker goroutine spawned by this library.
The caller can not recover it with its own `defer`/`recover` (recover only works
inside the panicking goroutine), and an unrecovered panic in any goroutine kills
the whole process. To prevent that, every worker recovers panics and the library
propagates them to the caller:

- Blocking functions (`ApplyChan`, `ApplySlice`, `Execute`, `ExecuteOpts`,
  `MapSlice`, `MapSliceOrdered`) re-raise the first recovered panic in the
  **calling** goroutine as `*parallel.PanicError` after all workers finish.
  For the caller it looks exactly like a panic in synchronous code, so an
  existing recovery layer (e.g. a gRPC recovery interceptor) handles it.
- `MapChan` is non-blocking, so recovered panics are delivered to the returned
  errors channel as `*parallel.PanicError` values.

`PanicError` keeps the original panic value (`Value`) and the stack trace of the
worker goroutine (`Stack`). If the panic value is an `error`, `PanicError`
unwraps to it, so `errors.Is`/`errors.As` work through it.

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
ConcurrencyLimiter is a helper that can limit amount of concurrently processed requests.

## Example
```go
limiter := NewConcurrencyLimiter(ops.concurrency)
for i := 0; i < 10000; i++ {
    limiter.Acquire()
    go func(i int) {
        defer limiter.Release()
        longJob(i)
    }(i)
}
```
