package promise

import "go.uber.org/zap"

type (
	Option struct {
		maxConcurrency int        // max goroutine run concurrent
		logger         zap.Logger // show log debug event
	}
)

func WithMaxConcurrency(max int) func(option *Option) *Option {
	return func(option *Option) *Option {
		option.maxConcurrency = max
		return option
	}
}

func WithZapLog(log zap.Logger) func(option *Option) *Option {
	return func(option *Option) *Option {
		option.logger = log
		return option
	}
}
