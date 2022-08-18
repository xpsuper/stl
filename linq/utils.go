package linq

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func isEmpty(data interface{}) bool {
	if data == nil {
		return true
	}
	dataRef := reflect.ValueOf(data)
	for dataRef.Kind() == reflect.Ptr {
		dataRef = dataRef.Elem()
	}
	switch dataRef.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return dataRef.Len() == 0
	case reflect.String:
		return strings.Trim(data.(string), " ") == ""
	}
	return false
}

func equal(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	return reflect.DeepEqual(expected, actual)
}

func grow(s reflect.Value) (v reflect.Value, newCap int) {
	cap := s.Cap()
	if cap == 0 {
		cap = 1
	} else {
		cap *= 2
	}
	newSlice := reflect.MakeSlice(s.Type(), cap, cap)
	reflect.Copy(newSlice, s)
	return newSlice, cap
}

type intConverter func(interface{}) int64

func getIntConverter(data interface{}) intConverter {
	switch data.(type) {
	case (int):
		return func(i interface{}) int64 {
			return int64(i.(int))
		}
	case (int8):
		return func(i interface{}) int64 {
			return int64(i.(int8))
		}
	case (int16):
		return func(i interface{}) int64 {
			return int64(i.(int16))
		}
	case (int32):
		return func(i interface{}) int64 {
			return int64(i.(int32))
		}
	}

	return func(i interface{}) int64 {
		return i.(int64)
	}
}

type uintConverter func(interface{}) uint64

func getUIntConverter(data interface{}) uintConverter {
	switch data.(type) {
	case (uint):
		return func(i interface{}) uint64 {
			return uint64(i.(uint))
		}
	case (uint8):
		return func(i interface{}) uint64 {
			return uint64(i.(uint8))
		}
	case (uint16):
		return func(i interface{}) uint64 {
			return uint64(i.(uint16))
		}
	case (uint32):
		return func(i interface{}) uint64 {
			return uint64(i.(uint32))
		}
	}

	return func(i interface{}) uint64 {
		return i.(uint64)
	}
}

type floatConverter func(interface{}) float64

func getFloatConverter(data interface{}) floatConverter {
	switch data.(type) {
	case (float32):
		return func(i interface{}) float64 {
			return float64(i.(float32))
		}
	}

	return func(i interface{}) float64 {
		return i.(float64)
	}
}

type comparer func(interface{}, interface{}) int

// Comparable 自定义比较接口, 如果调用 linq 的 OrderBy 等排序方法, 自定义对象需实现此接口
//
// Example:
// 	func (f foo) CompareTo(c Comparable) int {
// 		a, b := f.f1, c.(foo).f1
//
// 		if a < b {
// 			return -1
// 		} else if a > b {
// 			return 1
// 		}
//
// 		return 0
// 	}
type Comparable interface {
	CompareTo(Comparable) int
}

func getComparer(data interface{}) comparer {
	switch data.(type) {
	case int:
		return func(x, y interface{}) int {
			a, b := x.(int), y.(int)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case int8:
		return func(x, y interface{}) int {
			a, b := x.(int8), y.(int8)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case int16:
		return func(x, y interface{}) int {
			a, b := x.(int16), y.(int16)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case int32:
		return func(x, y interface{}) int {
			a, b := x.(int32), y.(int32)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case int64:
		return func(x, y interface{}) int {
			a, b := x.(int64), y.(int64)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case uint:
		return func(x, y interface{}) int {
			a, b := x.(uint), y.(uint)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case uint8:
		return func(x, y interface{}) int {
			a, b := x.(uint8), y.(uint8)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case uint16:
		return func(x, y interface{}) int {
			a, b := x.(uint16), y.(uint16)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case uint32:
		return func(x, y interface{}) int {
			a, b := x.(uint32), y.(uint32)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case uint64:
		return func(x, y interface{}) int {
			a, b := x.(uint64), y.(uint64)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case float32:
		return func(x, y interface{}) int {
			a, b := x.(float32), y.(float32)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case float64:
		return func(x, y interface{}) int {
			a, b := x.(float64), y.(float64)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case string:
		return func(x, y interface{}) int {
			a, b := x.(string), y.(string)
			switch {
			case a > b:
				return 1
			case b > a:
				return -1
			default:
				return 0
			}
		}
	case bool:
		return func(x, y interface{}) int {
			a, b := x.(bool), y.(bool)
			switch {
			case a == b:
				return 0
			case a:
				return 1
			default:
				return -1
			}
		}
	default:
		return func(x, y interface{}) int {
			a, b := x.(Comparable), y.(Comparable)
			return a.CompareTo(b)
		}
	}
}

// 内部排序对象
type sorter struct {
	items []interface{}
	less  func(i, j interface{}) bool
}

func (s sorter) Len() int {
	return len(s.items)
}

func (s sorter) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

func (s sorter) Less(i, j int) bool {
	return s.less(s.items[i], s.items[j])
}

func (q LinQuery) sort(orders []order) (r []interface{}) {
	next := q.Iterate()
	for item, ok := next(); ok; item, ok = next() {
		r = append(r, item)
	}

	if len(r) == 0 {
		return
	}

	for i, j := range orders {
		orders[i].compare = getComparer(j.selector(r[0]))
	}

	s := sorter{
		items: r,
		less: func(i, j interface{}) bool {
			for _, order := range orders {
				x, y := order.selector(i), order.selector(j)
				switch order.compare(x, y) {
				case 0:
					continue
				case -1:
					return !order.desc
				default:
					return order.desc
				}
			}

			return false
		}}

	sort.Sort(s)
	return
}

func (q LinQuery) lessSort(less func(i, j interface{}) bool) (r []interface{}) {
	next := q.Iterate()
	for item, ok := next(); ok; item, ok = next() {
		r = append(r, item)
	}

	s := sorter{items: r, less: less}

	sort.Sort(s)
	return
}

// genericType represents a any reflect.Type.
type genericType int

var genericTp = reflect.TypeOf(new(genericType)).Elem()

// functionCache keeps genericFunc reflection objects in cache.
type functionCache struct {
	MethodName string
	ParamName  string
	FnValue    reflect.Value
	FnType     reflect.Type
	TypesIn    []reflect.Type
	TypesOut   []reflect.Type
}

// genericFunc is a type used to validate and call dynamic functions.
type genericFunc struct {
	Cache *functionCache
}

// Call calls a dynamic function.
func (g *genericFunc) Call(params ...interface{}) interface{} {
	paramsIn := make([]reflect.Value, len(params))
	for i, param := range params {
		paramsIn[i] = reflect.ValueOf(param)
	}
	paramsOut := g.Cache.FnValue.Call(paramsIn)
	if len(paramsOut) >= 1 {
		return paramsOut[0].Interface()
	}
	return nil
}

// newGenericFunc instantiates a new genericFunc pointer
func newGenericFunc(methodName, paramName string, fn interface{}, validateFunc func(*functionCache) error) (*genericFunc, error) {
	cache := &functionCache{}
	cache.FnValue = reflect.ValueOf(fn)

	if cache.FnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("%s: parameter [%s] is not a function type. It is a '%s'", methodName, paramName, cache.FnValue.Type())
	}
	cache.MethodName = methodName
	cache.ParamName = paramName
	cache.FnType = cache.FnValue.Type()
	numTypesIn := cache.FnType.NumIn()
	cache.TypesIn = make([]reflect.Type, numTypesIn)
	for i := 0; i < numTypesIn; i++ {
		cache.TypesIn[i] = cache.FnType.In(i)
	}

	numTypesOut := cache.FnType.NumOut()
	cache.TypesOut = make([]reflect.Type, numTypesOut)
	for i := 0; i < numTypesOut; i++ {
		cache.TypesOut[i] = cache.FnType.Out(i)
	}
	if err := validateFunc(cache); err != nil {
		return nil, err
	}

	return &genericFunc{Cache: cache}, nil
}

// simpleParamValidator creates a function to validate genericFunc based in the
// In and Out function parameters.
func simpleParamValidator(In []reflect.Type, Out []reflect.Type) func(cache *functionCache) error {
	return func(cache *functionCache) error {
		var isValid = func() bool {
			if In != nil {
				if len(In) != len(cache.TypesIn) {
					return false
				}
				for i, paramIn := range In {
					if paramIn != genericTp && paramIn != cache.TypesIn[i] {
						return false
					}
				}
			}
			if Out != nil {
				if len(Out) != len(cache.TypesOut) {
					return false
				}
				for i, paramOut := range Out {
					if paramOut != genericTp && paramOut != cache.TypesOut[i] {
						return false
					}
				}
			}
			return true
		}

		if !isValid() {
			return fmt.Errorf("%s: parameter [%s] has a invalid function signature. Expected: '%s', actual: '%s'", cache.MethodName, cache.ParamName, formatFnSignature(In, Out), formatFnSignature(cache.TypesIn, cache.TypesOut))
		}
		return nil
	}
}

// newElemTypeSlice creates a slice of items elem types.
func newElemTypeSlice(items ...interface{}) []reflect.Type {
	typeList := make([]reflect.Type, len(items))
	for i, item := range items {
		typeItem := reflect.TypeOf(item)
		if typeItem.Kind() == reflect.Ptr {
			typeList[i] = typeItem.Elem()
		}
	}
	return typeList
}

// formatFnSignature formats the func signature based in the parameters types.
func formatFnSignature(In []reflect.Type, Out []reflect.Type) string {
	paramInNames := make([]string, len(In))
	for i, typeIn := range In {
		if typeIn == genericTp {
			paramInNames[i] = "T"
		} else {
			paramInNames[i] = typeIn.String()
		}

	}
	paramOutNames := make([]string, len(Out))
	for i, typeOut := range Out {
		if typeOut == genericTp {
			paramOutNames[i] = "T"
		} else {
			paramOutNames[i] = typeOut.String()
		}
	}
	return fmt.Sprintf("func(%s)%s", strings.Join(paramInNames, ","), strings.Join(paramOutNames, ","))
}
