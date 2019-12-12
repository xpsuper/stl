package taskbus

import (
	"fmt"
	"reflect"
)

// Interface fot async to handle user user functions
type Taskier interface {
	GetFuncs() (interface{}, error)
}

// Type used as a slisce of tasks
type Tasks []interface{}

func (t Tasks) GetFuncs() (interface{}, error) {
	var (
		l   = len(t)
		fns = make([]reflect.Value, l)
	)

	for i := 0; i < l; i++ {
		f := reflect.Indirect(reflect.ValueOf(t[i]))

		if f.Kind() != reflect.Func {
			return fns, fmt.Errorf("%T must be a Function ", f)
		}

		fns[i] = f
	}

	return fns, nil
}

// Type used as a map of tasks
type MapTasks map[string]interface{}

func (mt MapTasks) GetFuncs() (interface{}, error) {
	var fns = map[string]reflect.Value{}

	for k, v := range mt {
		f := reflect.Indirect(reflect.ValueOf(v))

		if f.Kind() != reflect.Func {
			return fns, fmt.Errorf("%T must be a Function ", f)
		}

		fns[k] = f
	}

	return fns, nil
}


func Chain(stack Tasks, firstArgs ...interface{}) ([]interface{}, error) {
	var (
		err  error
		args []reflect.Value
		f    = &funcs{}
	)
	// Checks if the Tasks passed are valid functions.
	f.Stack, err = stack.GetFuncs()

	if err != nil {
		panic(err.Error())
	}

	// transform interface{} to reflect.Value for execution
	for i := 0; i < len(firstArgs); i++ {
		args = append(args, reflect.ValueOf(firstArgs[i]))
	}

	return f.ExecInSeries(args...)
}

func Max(stack Taskier) (Results, error) {
	return execConcurrentStack(stack, true)
}

func All(stack Taskier) (Results, error) {
	return execConcurrentStack(stack, false)
}

func execConcurrentStack(stack Taskier, parallel bool) (Results, error) {
	var (
		err error
		f   = &funcs{}
	)

	// Checks if the Tasks passed are valid functions.
	f.Stack, err = stack.GetFuncs()

	if err != nil {
		panic(err)
	}
	return f.ExecConcurrent(parallel)
}
