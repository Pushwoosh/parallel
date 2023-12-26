package parallel

// ConcurrencyLimiter is a helper that can limit amount of concurrently processed requests.
// See concurrency_limiter_test.go for usage example.
type ConcurrencyLimiter interface {
	// Acquire acquires a slot in the concurrency limiter.
	// Blocks until a slot is available.
	Acquire()

	// Release releases a slot in the concurrency limiter.
	Release()
}

func NewConcurrencyLimiter(concurrency int) ConcurrencyLimiter {
	ret := make(concurrencyLimiter, concurrency)
	for i := 0; i < concurrency; i++ {
		ret <- struct{}{}
	}
	return ret
}

type concurrencyLimiter chan struct{}

func (c concurrencyLimiter) Acquire() {
	<-c
}

func (c concurrencyLimiter) Release() {
	c <- struct{}{}
}
