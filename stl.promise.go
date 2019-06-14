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
	state int
	executor func(resolve func(interface{}), reject func(error))
	then []func(data interface{}) interface{}
	catch []func(error error) error
	result interface{}
	error error
	mutex *sync.Mutex
	wg *sync.WaitGroup
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
		error:    nil,
		mutex:    &sync.Mutex{},
		wg:       wg,
	}

	go func() {
		defer promise.handlePanic()
		promise.executor(promise.resolve, promise.reject)
	}()

	return promise
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

func (promise *XPPromiseImpl) resolve(resolution interface{}) {
	promise.mutex.Lock()

	if promise.state != pending {
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

func (promise *XPPromiseImpl) reject(error error) {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state != pending {
		return
	}

	promise.error = error

	promise.wg.Done()
	for range promise.then {
		promise.wg.Done()
	}

	for _, fn := range promise.catch {
		promise.error = fn(promise.error)
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

	if promise.state == pending {
		promise.wg.Add(1)
		promise.then = append(promise.then, fulfillment)
	} else if promise.state == fulfilled {
		promise.result = fulfillment(promise.result)
	}

	return promise
}

func (promise *XPPromiseImpl) Catch(rejection func(error error) error) *XPPromiseImpl {
	promise.mutex.Lock()
	defer promise.mutex.Unlock()

	if promise.state == pending {
		promise.wg.Add(1)
		promise.catch = append(promise.catch, rejection)
	} else if promise.state == rejected {
		promise.error = rejection(promise.error)
	}

	return promise
}

func (promise *XPPromiseImpl) Await() (interface{}, error) {
	promise.wg.Wait()
	return promise.result, promise.error
}