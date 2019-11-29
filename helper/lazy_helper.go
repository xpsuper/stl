package helper

import "reflect"

type lazyHelper struct {
	exec func() interface{}
}

func (b *lazyHelper) Chunk(size int) Helper {
	return &lazyHelper{func() interface{} { return Chunk(b.exec(), size) }}
}
func (b *lazyHelper) Compact() Helper {
	return &lazyHelper{func() interface{} { return Compact(b.exec()) }}
}
func (b *lazyHelper) Drop(n int) Helper {
	return &lazyHelper{func() interface{} { return Drop(b.exec(), n) }}
}
func (b *lazyHelper) Filter(predicate interface{}) Helper {
	return &lazyHelper{func() interface{} { return Filter(b.exec(), predicate) }}
}
func (b *lazyHelper) FlattenDeep() Helper {
	return &lazyHelper{func() interface{} { return FlattenDeep(b.exec()) }}
}
func (b *lazyHelper) Initial() Helper {
	return &lazyHelper{func() interface{} { return Initial(b.exec()) }}
}
func (b *lazyHelper) Intersect(y interface{}) Helper {
	return &lazyHelper{func() interface{} { return Intersect(b.exec(), y) }}
}
func (b *lazyHelper) Map(mapFunc interface{}) Helper {
	return &lazyHelper{func() interface{} { return Map(b.exec(), mapFunc) }}
}
func (b *lazyHelper) Reverse() Helper {
	return &lazyHelper{func() interface{} { return Reverse(b.exec()) }}
}
func (b *lazyHelper) Shuffle() Helper {
	return &lazyHelper{func() interface{} { return Shuffle(b.exec()) }}
}
func (b *lazyHelper) Tail() Helper {
	return &lazyHelper{func() interface{} { return Tail(b.exec()) }}
}
func (b *lazyHelper) Uniq() Helper {
	return &lazyHelper{func() interface{} { return Uniq(b.exec()) }}
}

func (b *lazyHelper) All() bool {
	return (&chainHelper{b.exec()}).All()
}
func (b *lazyHelper) Any() bool {
	return (&chainHelper{b.exec()}).Any()
}
func (b *lazyHelper) Contains(elem interface{}) bool {
	return Contains(b.exec(), elem)
}
func (b *lazyHelper) Every(elements ...interface{}) bool {
	return Every(b.exec(), elements...)
}
func (b *lazyHelper) Find(predicate interface{}) interface{} {
	return Find(b.exec(), predicate)
}
func (b *lazyHelper) ForEach(predicate interface{}) {
	ForEach(b.exec(), predicate)
}
func (b *lazyHelper) ForEachRight(predicate interface{}) {
	ForEachRight(b.exec(), predicate)
}
func (b *lazyHelper) Head() interface{} {
	return Head(b.exec())
}
func (b *lazyHelper) Keys() interface{} {
	return Keys(b.exec())
}
func (b *lazyHelper) IndexOf(elem interface{}) int {
	return IndexOf(b.exec(), elem)
}
func (b *lazyHelper) IsEmpty() bool {
	return IsEmpty(b.exec())
}
func (b *lazyHelper) Last() interface{} {
	return Last(b.exec())
}
func (b *lazyHelper) LastIndexOf(elem interface{}) int {
	return LastIndexOf(b.exec(), elem)
}
func (b *lazyHelper) NotEmpty() bool {
	return NotEmpty(b.exec())
}
func (b *lazyHelper) Product() float64 {
	return Product(b.exec())
}
func (b *lazyHelper) Reduce(reduceFunc, acc interface{}) float64 {
	return Reduce(b.exec(), reduceFunc, acc)
}
func (b *lazyHelper) Sum() float64 {
	return Sum(b.exec())
}
func (b *lazyHelper) Type() reflect.Type {
	return reflect.TypeOf(b.exec())
}
func (b *lazyHelper) Value() interface{} {
	return b.exec()
}
func (b *lazyHelper) Values() interface{} {
	return Values(b.exec())
}
