package retrier

import "time"

type (
	Options struct {
		onErrorFunc      func(error)
		useErrorFunc     bool
		errChan          chan error
		sendErrorsInChan bool
	}

	Option func(option *Options)
)

func WithErrChan(output chan error) Option {
	return func(option *Options) {
		option.errChan = output
		option.sendErrorsInChan = true
	}
}

func WithOnErrorFunc(errFunc func(error)) Option {
	return func(option *Options) {
		option.onErrorFunc = errFunc
		option.useErrorFunc = true
	}
}

func DoWithInterval(interval time.Duration, do func() error, opts ...Option) {
	option := new(Options)
	option.sendErrorsInChan = false
	option.useErrorFunc = false

	for _, opt := range opts {
		opt(option)
	}

	go func() {
		time.Sleep(interval)

		if err := do(); err != nil {
			if option.sendErrorsInChan {
				option.errChan <- err
			}
			if option.useErrorFunc {
				option.onErrorFunc(err)
			}
		}
	}()
}
