package parallel

const (
	ExecuteDefaultConcurrency = 10
)

type ExecuteOption interface {
	isExecuteOption()
}

// executeOptionConcurrency sets concurrency for parallel execution.
type executeOptionConcurrency struct{ concurrency int }

func (o executeOptionConcurrency) isExecuteOption() {}

func WithExecuteConcurrency(concurrency int) ExecuteOption {
	return executeOptionConcurrency{concurrency: concurrency}
}

type executeOptions struct {
	concurrency int
}

func parseExecuteOptions(opts []ExecuteOption) executeOptions {
	ret := executeOptions{
		concurrency: ExecuteDefaultConcurrency,
	}

	for i := range opts {
		// nolint: gocritic
		switch opt := opts[i].(type) {
		case executeOptionConcurrency:
			ret.concurrency = opt.concurrency
		}
	}

	return ret
}
