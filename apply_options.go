package parallel

const (
	ApplyDefaultConcurrency = 10
)

type ApplyOption interface {
	isApplyOption()
}

// applyOptionConcurrency sets concurrency for parallel execution.
type applyOptionConcurrency struct{ concurrency int }

func (o applyOptionConcurrency) isApplyOption() {}

func WithApplyConcurrency(concurrency int) ApplyOption {
	return applyOptionConcurrency{concurrency: concurrency}
}

type applyOptions struct {
	concurrency int
}

func parseApplyOptions(opts []ApplyOption) applyOptions {
	ret := applyOptions{
		concurrency: ApplyDefaultConcurrency,
	}

	for i := range opts {
		// nolint: gocritic
		switch opt := opts[i].(type) {
		case applyOptionConcurrency:
			ret.concurrency = opt.concurrency
		}
	}

	return ret
}
