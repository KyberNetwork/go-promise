package promise

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

type (
	Promise[T any] struct {
		execute func(ctx context.Context, resolve func(data T), reject func(err error))
	}

	Job[T any] struct {
		promise Promise[T]
		index   int
	}

	Result[T any] struct {
		data  T
		index int
	}

	DataOrError[T any] struct {
		data    T
		isData  bool
		err     error
		isError bool
	}
)

func NewPromise[T any](execute func(ctx context.Context, resolve func(data T), reject func(err error))) Promise[T] {
	return Promise[T]{execute: execute}
}

// All
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all
func All[T any](ctx context.Context, promises []Promise[T], optionFuncs ...func(option *Option) *Option) ([]T, error) {
	option := &Option{
		maxConcurrency: len(promises),
	}
	for _, optionFunc := range optionFuncs {
		option = optionFunc(option)
	}

	wg := &sync.WaitGroup{}

	jobs := make(chan Job[T], len(promises))
	go func() {
		for i, promise := range promises {
			jobs <- Job[T]{promise: promise, index: i}
		}
		close(jobs)
	}()

	resultChan := make(chan Result[T], len(promises))
	errChan := make(chan error, len(promises))
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	for i := 0; i < option.maxConcurrency; i++ {
		wg.Add(1)
		go jobHandler[T](cancelCtx, wg, jobs, resultChan, errChan)
	}

	results := make([]T, len(promises))
	var err error
	doneChan := make(chan bool, 1)
	go func() {
		for errHappen := range errChan {
			err = errHappen
			doneChan <- true
			cancelFunc()
			return
		}

		for result := range resultChan {
			results[result.index] = result.data
		}
		doneChan <- true
	}()

	wg.Wait()
	close(resultChan)
	close(errChan)
	<-doneChan
	cancelFunc()

	return results, err
}

// Any
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/any
func Any[T any](ctx context.Context, promises []Promise[T], optionFuncs ...func(option *Option) *Option) (T, error) {
	option := &Option{
		maxConcurrency: len(promises),
	}
	for _, optionFunc := range optionFuncs {
		option = optionFunc(option)
	}

	wg := &sync.WaitGroup{}

	jobs := make(chan Job[T], len(promises))
	go func() {
		for i, promise := range promises {
			jobs <- Job[T]{promise: promise, index: i}
		}
		close(jobs)
	}()

	resultChan := make(chan Result[T], len(promises))
	errChan := make(chan error, len(promises))
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	for i := 0; i < option.maxConcurrency; i++ {
		wg.Add(1)
		go jobHandler[T](cancelCtx, wg, jobs, resultChan, errChan)
	}

	var result T
	var err error
	doneChan := make(chan bool, 1)
	go func() {
		for r := range resultChan {
			result = r.data
			doneChan <- true
			return
		}

		for errHappen := range errChan {
			if err == nil {
				err = errHappen
			} else {
				err = errors.Wrap(err, errHappen.Error())
			}
		}
		doneChan <- true
	}()

	wg.Wait()
	close(resultChan)
	close(errChan)
	<-doneChan
	cancelFunc()

	if err != nil {
		return result, err
	}
	return result, nil
}

// Race
// https: //developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/race
func Race[T any](ctx context.Context, promises []Promise[T]) (interface{}, error) {
	option := &Option{
		maxConcurrency: len(promises),
	}

	wg := &sync.WaitGroup{}

	jobs := make(chan Job[T], len(promises))
	go func() {
		for i, promise := range promises {
			jobs <- Job[T]{promise: promise, index: i}
		}
		close(jobs)
	}()

	dataOrErrorChan := make(chan DataOrError[T], len(promises))
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	for i := 0; i < option.maxConcurrency; i++ {
		wg.Add(1)
		go jobRaceHandler[T](cancelCtx, wg, jobs, dataOrErrorChan)
	}

	var result interface{}
	var err error
	doneChan := make(chan bool, 1)
	go func() {
		dataOrErr, ok := <-dataOrErrorChan
		if ok {
			if dataOrErr.isData {
				result = dataOrErr.data
			}
			if dataOrErr.isError {
				err = dataOrErr.err
			}
		}
		doneChan <- true
	}()

	wg.Wait()
	close(dataOrErrorChan)
	<-doneChan
	cancelFunc()

	if result != nil {
		return result, nil
	}
	return nil, err
}

func jobHandler[T any](ctx context.Context, wg *sync.WaitGroup, jobs chan Job[T], resultChan chan Result[T], errChan chan error) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}

			resolve := func(data T) {
				resultChan <- Result[T]{
					data:  data,
					index: job.index,
				}
			}

			reject := func(err error) {
				errChan <- err
			}

			job.promise.execute(ctx, resolve, reject)
		case <-ctx.Done():
			return
		}
	}
}

func jobRaceHandler[T any](ctx context.Context, wg *sync.WaitGroup, jobs chan Job[T], chanel chan DataOrError[T]) {
	defer wg.Done()
	for {
		select {
		case job, ok := <-jobs:
			if !ok {
				return
			}

			resolve := func(data T) {
				chanel <- DataOrError[T]{
					data:   data,
					isData: true,
				}
			}

			reject := func(err error) {
				chanel <- DataOrError[T]{
					err:     err,
					isError: true,
				}
			}

			job.promise.execute(ctx, resolve, reject)
		case <-ctx.Done():
			return
		}
	}
}
