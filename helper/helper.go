package helper

import (
	"fmt"
	"reflect"
)

// Helper contains all tools which can be chained.
type Helper interface {
	Chunk(size int) Helper
	Compact() Helper
	Drop(n int) Helper
	Filter(predicate interface{}) Helper
	FlattenDeep() Helper
	Initial() Helper
	Intersect(y interface{}) Helper
	Map(mapFunc interface{}) Helper
	Reverse() Helper
	Shuffle() Helper
	Tail() Helper
	Uniq() Helper

	All() bool
	Any() bool
	Contains(elem interface{}) bool
	Every(elements ...interface{}) bool
	Find(predicate interface{}) interface{}
	ForEach(predicate interface{})
	ForEachRight(predicate interface{})
	Head() interface{}
	Keys() interface{}
	IndexOf(elem interface{}) int
	IsEmpty() bool
	Last() interface{}
	LastIndexOf(elem interface{}) int
	NotEmpty() bool
	Product() float64
	Reduce(reduceFunc, acc interface{}) float64
	Sum() float64
	Type() reflect.Type
	Value() interface{}
	Values() interface{}
}

// Chain creates a simple new helper.Helper from a collection. Each method 
// call generate a new Helper containing the previous result.
func Chain(v interface{}) Helper {
	isNotNil(v, "Chain")

	valueType := reflect.TypeOf(v)
	if isValidHelperEntry(valueType) ||
		(valueType.Kind() == reflect.Ptr && isValidHelperEntry(valueType.Elem())) {
		return &chainHelper{v}
	}

	panic(fmt.Sprintf("Type %s is not supported by Chain", valueType.String()))
}

// LazyChain creates a lazy helper.Helper from a collection. Each method call
// generate a new Helper containing a method generating the previous value.
// With that, all data are only generated when we call a tailling method like All or Find.
func LazyChain(v interface{}) Helper {
	isNotNil(v, "LazyChain")

	valueType := reflect.TypeOf(v)
	if isValidHelperEntry(valueType) ||
		(valueType.Kind() == reflect.Ptr && isValidHelperEntry(valueType.Elem())) {
		return &lazyHelper{func() interface{} { return v }}
	}

	panic(fmt.Sprintf("Type %s is not supported by LazyChain", valueType.String()))

}

// LazyChainWith creates a lzy helper.Helper from a generator. Like LazyChain, each 
// method call generate a new Helper containing a method generating the previous value.
// But, instead of using a collection, it takes a generator which can generate values.
// With LazyChainWith, to can create a generic pipeline of collection transformation and, 
// throw the generator, sending different collection.
func LazyChainWith(generator func() interface{}) Helper {
	isNotNil(generator, "LazyChainWith")
	return &lazyHelper{func() interface{} {
		isNotNil(generator, "LazyChainWith")

		v := generator()
		valueType := reflect.TypeOf(v)
		if isValidHelperEntry(valueType) ||
			(valueType.Kind() == reflect.Ptr && isValidHelperEntry(valueType.Elem())) {
			return v
		}

		panic(fmt.Sprintf("Type %s is not supported by LazyChainWith generator", valueType.String()))
	}}
}

func isNotNil(v interface{}, from string) {
	if v == nil {
		panic(fmt.Sprintf("nil value is not supported by %s", from))
	}
}

func isValidHelperEntry(valueType reflect.Type) bool {
	return valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array ||
		valueType.Kind() == reflect.Map ||
		valueType.Kind() == reflect.String
}
