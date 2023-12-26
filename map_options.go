package parallel

const (
	MapDefaultConcurrency = 10
)

type MapOption interface {
	isMapOption()
}

// mapOptionConcurrency sets concurrency for parallel execution.
type mapOptionConcurrency struct{ concurrency int }

// mapOptionStopOnFirstError force executor to stop processing right after the first error occurred.
type mapOptionStopOnFirstError struct{}

func (o mapOptionConcurrency) isMapOption()      {}
func (o mapOptionStopOnFirstError) isMapOption() {}

func WithMapConcurrency(concurrency int) MapOption {
	return mapOptionConcurrency{concurrency: concurrency}
}

func WithMapStopOnFirstError() MapOption {
	return mapOptionStopOnFirstError{}
}

type mapOptions struct {
	concurrency      int
	stopOnFirstError bool
}

func parseMapOptions(opts []MapOption) mapOptions {
	ret := mapOptions{
		concurrency:      MapDefaultConcurrency,
		stopOnFirstError: false,
	}

	for i := range opts {
		switch opt := opts[i].(type) {
		case mapOptionConcurrency:
			ret.concurrency = opt.concurrency
		case mapOptionStopOnFirstError:
			ret.stopOnFirstError = true
		}
	}

	return ret
}
