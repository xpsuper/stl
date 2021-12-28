package stl

import (
	"errors"
	"sync"
)

const (
	pending = iota
	fulfilled
	rejected
)

type XPPromiseImpl struct {
	state    int
	executor func(resolve func(interface{}), reject func(error))
	then     []func(data interface{}) interface{}
	catch    []func(err error) error
	result   interface{}
	err      error
	mutex    *sync.Mutex
	wg       *sync.WaitGroup
}

func NewPromise(executor func(resolve func(interface{}), reject func(error))) *XPPromiseImpl {
	var wg = &sync.WaitGroup{}
	wg.Add(1)

	var promise = &XPPromiseImpl{
		state:    pending,
		executor: executor,
		then:     make([]func(interface{}) interface{}, 0),
		catch:    make([]func(error) error, 0),
		result:   nil,
		err:      nil,
		mutex:    &sync.Mutex{},
		wg:       wg,
	}

	go func() {
		defer promise.handlePanic()
		promise.executor(promise.resolve, promise.reject)
	}()

	return promise
}

func (promise *XPPromiseImpl) resolve(resolution interface{}) {
	promise.mutex.Lock()

	if promise.state != pending {
		promise.mutex.Unlock()
		return
	}

	switch result := resolution.(type) {
	case *XPPromiseImpl:
		res, err := result.Await()
		if err != nil {
			promise.mutex.Unlock()
			promise.reject(err)
			return
		}
		promise.result = res
	default:
		promise.result = result
	}

	promise.wg.Done()
	for range promise.catch {
		promise.wg.Done()
	}

	for _, fn := range promise.then {
		switch result := fn(promise.result).(type) {
		case *XPPromiseImpl:
			res, err := result.Await()
			if err != nil {
				promise.mutex.Unlock()
				promise.reject(err)
				return
			}
			promise.result = res
		default:
			promise.result = result
		}
		promise.wg.Done()
	}

	promise.state = fulfilled

	promise.mutex.Unlock()
}

func (promise *XPPromiseImpl) reject(err error) {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state != pending {
		return
	}

	promise.err = err

	promise.wg.Done()
	for range promise.then {
		promise.wg.Done()
	}

	for _, fn := range promise.catch {
		promise.err = fn(promise.err)
		promise.wg.Done()
	}

	promise.state = rejected
}

func (promise *XPPromiseImpl) handlePanic() {
	var r = recover()
	if r != nil {
		promise.reject(errors.New(r.(string)))
	}
}

func (promise *XPPromiseImpl) Then(fulfillment func(data interface{}) interface{}) *XPPromiseImpl {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	switch promise.state {
	case pending:
		promise.wg.Add(1)
		promise.then = append(promise.then, fulfillment)
	case fulfilled:
		promise.result = fulfillment(promise.result)
	}

	return promise
}

func (promise *XPPromiseImpl) Catch(rejection func(error error) error) *XPPromiseImpl {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	switch promise.state {
	case pending:
		promise.wg.Add(1)
		promise.catch = append(promise.catch, rejection)
	case rejected:
		promise.err = rejection(promise.err)
	}

	return promise
}

func (promise *XPPromiseImpl) Await() (interface{}, error) {
	promise.wg.Wait()
	return promise.result, promise.err
}

func PromiseAll(promises ...*XPPromiseImpl) *XPPromiseImpl {
	psLen := len(promises)
	if psLen == 0 {
		return ResolvePromise(make([]interface{}, 0))
	}

	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolutionsChan := make(chan []interface{}, psLen)
		errorChan := make(chan error, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data interface{}) interface{} {
					resolutionsChan <- []interface{}{i, data}
					return data
				}).Catch(func(err error) error {
					errorChan <- err
					return err
				})
			}(index)
		}

		resolutions := make([]interface{}, psLen)
		for x := 0; x < psLen; x++ {
			select {
			case resolution := <-resolutionsChan:
				resolutions[resolution[0].(int)] = resolution[1]

			case err := <-errorChan:
				reject(err)
				return
			}
		}
		resolve(resolutions)
	})
}

func PromiseRace(promises ...*XPPromiseImpl) *XPPromiseImpl {
	psLen := len(promises)
	if psLen == 0 {
		return ResolvePromise(nil)
	}

	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolutionsChan := make(chan interface{}, psLen)
		errorChan := make(chan error, psLen)

		for _, promise := range promises {
			promise.Then(func(data interface{}) interface{} {
				resolutionsChan <- data
				return data
			}).Catch(func(err error) error {
				errorChan <- err
				return err
			})
		}

		select {
		case resolution := <-resolutionsChan:
			resolve(resolution)

		case err := <-errorChan:
			reject(err)
		}
	})
}

func AllSettled(promises ...*XPPromiseImpl) *XPPromiseImpl {
	psLen := len(promises)
	if psLen == 0 {
		return ResolvePromise(make([]interface{}, 0))
	}

	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolutionsChan := make(chan []interface{}, psLen)

		for index, promise := range promises {
			func(i int) {
				promise.Then(func(data interface{}) interface{} {
					resolutionsChan <- []interface{}{i, data}
					return data
				}).Catch(func(err error) error {
					resolutionsChan <- []interface{}{i, err}
					return err
				})
			}(index)
		}

		resolutions := make([]interface{}, psLen)
		for x := 0; x < psLen; x++ {
			resolution := <-resolutionsChan
			resolutions[resolution[0].(int)] = resolution[1]
		}
		resolve(resolutions)
	})
}

func ResolvePromise(resolution interface{}) *XPPromiseImpl {
	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		resolve(resolution)
	})
}

func RejectPromise(err error) *XPPromiseImpl {
	return NewPromise(func(resolve func(interface{}), reject func(error)) {
		reject(err)
	})
}
