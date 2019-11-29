package helper

import (
	"fmt"
	"reflect"
)

type chainHelper struct {
	collection interface{}
}

func (b *chainHelper) Chunk(size int) Helper {
	return &chainHelper{Chunk(b.collection, size)}
}
func (b *chainHelper) Compact() Helper {
	return &chainHelper{Compact(b.collection)}
}
func (b *chainHelper) Drop(n int) Helper {
	return &chainHelper{Drop(b.collection, n)}
}
func (b *chainHelper) Filter(predicate interface{}) Helper {
	return &chainHelper{Filter(b.collection, predicate)}
}
func (b *chainHelper) FlattenDeep() Helper {
	return &chainHelper{FlattenDeep(b.collection)}
}
func (b *chainHelper) Initial() Helper {
	return &chainHelper{Initial(b.collection)}
}
func (b *chainHelper) Intersect(y interface{}) Helper {
	return &chainHelper{Intersect(b.collection, y)}
}
func (b *chainHelper) Map(mapFunc interface{}) Helper {
	return &chainHelper{Map(b.collection, mapFunc)}
}
func (b *chainHelper) Reverse() Helper {
	return &chainHelper{Reverse(b.collection)}
}
func (b *chainHelper) Shuffle() Helper {
	return &chainHelper{Shuffle(b.collection)}
}
func (b *chainHelper) Tail() Helper {
	return &chainHelper{Tail(b.collection)}
}
func (b *chainHelper) Uniq() Helper {
	return &chainHelper{Uniq(b.collection)}
}

func (b *chainHelper) All() bool {
	v := reflect.ValueOf(b.collection)
	t := v.Type()

	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Type %s is not supported by Chain.All", t.String()))
	}

	c := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		c[i] = v.Index(i).Interface()
	}
	return All(c...)
}
func (b *chainHelper) Any() bool {
	v := reflect.ValueOf(b.collection)
	t := v.Type()

	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Type %s is not supported by Chain.Any", t.String()))
	}

	c := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		c[i] = v.Index(i).Interface()
	}
	return Any(c...)
}
func (b *chainHelper) Contains(elem interface{}) bool {
	return Contains(b.collection, elem)
}
func (b *chainHelper) Every(elements ...interface{}) bool {
	return Every(b.collection, elements...)
}
func (b *chainHelper) Find(predicate interface{}) interface{} {
	return Find(b.collection, predicate)
}
func (b *chainHelper) ForEach(predicate interface{}) {
	ForEach(b.collection, predicate)
}
func (b *chainHelper) ForEachRight(predicate interface{}) {
	ForEachRight(b.collection, predicate)
}
func (b *chainHelper) Head() interface{} {
	return Head(b.collection)
}
func (b *chainHelper) Keys() interface{} {
	return Keys(b.collection)
}
func (b *chainHelper) IndexOf(elem interface{}) int {
	return IndexOf(b.collection, elem)
}
func (b *chainHelper) IsEmpty() bool {
	return IsEmpty(b.collection)
}
func (b *chainHelper) Last() interface{} {
	return Last(b.collection)
}
func (b *chainHelper) LastIndexOf(elem interface{}) int {
	return LastIndexOf(b.collection, elem)
}
func (b *chainHelper) NotEmpty() bool {
	return NotEmpty(b.collection)
}
func (b *chainHelper) Product() float64 {
	return Product(b.collection)
}
func (b *chainHelper) Reduce(reduceFunc, acc interface{}) float64 {
	return Reduce(b.collection, reduceFunc, acc)
}
func (b *chainHelper) Sum() float64 {
	return Sum(b.collection)
}
func (b *chainHelper) Type() reflect.Type {
	return reflect.TypeOf(b.collection)
}
func (b *chainHelper) Value() interface{} {
	return b.collection
}
func (b *chainHelper) Values() interface{} {
	return Values(b.collection)
}
