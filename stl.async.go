package stl

import (
	"reflect"
)

// 异步执行所需要的数据
type asyncRun struct {
	Handler reflect.Value
	Params  []reflect.Value
}

// Async 异步执行对象
type XPAsyncImpl struct {
	count int
	tasks map[string]asyncRun
}

// NewAsync 老版本的兼容
func NewAsync() *XPAsyncImpl {
	return New()
}

// New 创建一个新的异步执行对象
func New() *XPAsyncImpl {
	return &XPAsyncImpl{tasks: make(map[string]asyncRun)}
}

// Add 添加异步执行任务
// name 任务名，结果返回时也将放在任务名中
// handler 任务执行函数，将需要被执行的函数导入到程序中
// params 任务执行函数所需要的参数
func (instance *XPAsyncImpl) Add(name string, handler interface{}, params ...interface{}) bool {
	if _, e := instance.tasks[name]; e {
		return false
	}

	handlerValue := reflect.ValueOf(handler)
	if handlerValue.Kind() == reflect.Func {

		paramNum := len(params)

		instance.tasks[name] = asyncRun{
			Handler: handlerValue,
			Params:  make([]reflect.Value, paramNum),
		}

		if paramNum > 0 {
			for k, v := range params {
				instance.tasks[name].Params[k] = reflect.ValueOf(v)
			}
		}

		instance.count++
		return true
	}

	return false
}

// Run 任务执行函数，成功时将返回一个用于接受结果的channel
// 在所有异步任务都运行完成时，结果channel将会返回一个map[string][]interface{}的结果。
func (instance *XPAsyncImpl) Run() (chan map[string][]interface{}, bool) {
	if instance.count < 1 {
		return nil, false
	}
	result := make(chan map[string][]interface{})
	chans := make(chan map[string]interface{}, instance.count)

	go func(result chan map[string][]interface{}, chans chan map[string]interface{}) {
		rs := make(map[string][]interface{})
		defer func(rs map[string][]interface{}) {
			result <- rs
		}(rs)
		for {
			if instance.count < 1 {
				break
			}

			select {
			case res := <-chans:
				instance.count--
				rs[res["name"].(string)] = res["result"].([]interface{})
			}
		}
	}(result, chans)

	for k, v := range instance.tasks {
		go func(name string, chans chan map[string]interface{}, async asyncRun) {
			result := make([]interface{}, 0)
			defer func(name string, chans chan map[string]interface{}) {
				chans <- map[string]interface{}{"name": name, "result": result}
			}(name, chans)

			values := async.Handler.Call(async.Params)

			if valuesNum := len(values); valuesNum > 0 {
				resultItems := make([]interface{}, valuesNum)
				for k, v := range values {
					resultItems[k] = v.Interface()
				}
				result = resultItems
				return
			}
		}(k, chans, v)
	}

	return result, true
}

// Clean 清空任务队列.
func (instance *XPAsyncImpl) Clean() {
	instance.tasks = make(map[string]asyncRun)
}
