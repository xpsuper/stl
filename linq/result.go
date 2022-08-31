package linq

import (
	"math"
	"reflect"
	"strings"
)

// IsEmpty 判断是否为空
func (q LinQuery) IsEmpty() bool {
	_, ok := q.Iterate()()
	return !ok
}

// All 检查集合的所有元素是否都符合给定条件
func (q LinQuery) All(predicate func(interface{}) bool) bool {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if !predicate(item) {
			return false
		}
	}

	return true
}

func (q LinQuery) AllT(predicateFn interface{}) bool {

	predicateGenericFunc, err := newGenericFunc(
		"AllT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}
	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.All(predicateFunc)
}

// Any 检查集合的元素是否有一个符合给定条件
func (q LinQuery) Any(predicate func(interface{}) bool) bool {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if predicate(item) {
			return true
		}
	}

	return false
}

func (q LinQuery) AnyT(predicateFn interface{}) bool {

	predicateGenericFunc, err := newGenericFunc(
		"AnyT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.Any(predicateFunc)
}

// Average 计算数值集合的平均值
func (q LinQuery) Average() (r float64) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return math.NaN()
	}

	n := 1
	switch item.(type) {
	case int, int8, int16, int32, int64:
		conv := getIntConverter(item)
		sum := conv(item)

		for item, ok = next(); ok; item, ok = next() {
			sum += conv(item)
			n++
		}

		r = float64(sum)
	case uint, uint8, uint16, uint32, uint64:
		conv := getUIntConverter(item)
		sum := conv(item)

		for item, ok = next(); ok; item, ok = next() {
			sum += conv(item)
			n++
		}

		r = float64(sum)
	default:
		conv := getFloatConverter(item)
		r = conv(item)

		for item, ok = next(); ok; item, ok = next() {
			r += conv(item)
			n++
		}
	}

	return r / float64(n)
}

// Contains 确定集合是否包含指定的元素
func (q LinQuery) Contains(value interface{}) bool {
	if !q.Source.IsValid() {
		return false
	}

	if q.Source.Type().Kind() == reflect.String {
		if reflect.ValueOf(value).Type().Kind() == reflect.String {
			str := q.Source.Interface().(string)
			return strings.Contains(str, value.(string))
		}
	}

	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if item == value {
			return true
		}
	}

	return false
}

// ContainsKey 确定Map中是否包含指定的键
func (q LinQuery) ContainsKey(key interface{}) bool {
	if q.Source.IsValid() && q.Source.Type().Kind() == reflect.Map {
		next := q.Iterate()

		for item, ok := next(); ok; item, ok = next() {
			kv := item.(KeyValue)
			if equal(kv.Key, key) {
				return true
			}
		}
	}

	return false
}

// Count 返回集合中的元素个数
func (q LinQuery) Count() (r int) {
	next := q.Iterate()

	for _, ok := next(); ok; _, ok = next() {
		r++
	}

	return
}

// CountWith 返回集合中满足条件的元素个数
func (q LinQuery) CountWith(predicate func(interface{}) bool) (r int) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if predicate(item) {
			r++
		}
	}

	return
}

func (q LinQuery) CountWithT(predicateFn interface{}) int {

	predicateGenericFunc, err := newGenericFunc(
		"CountWithT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.CountWith(predicateFunc)
}

// First 返回集合的第一个元素
func (q LinQuery) First() interface{} {
	item, _ := q.Iterate()()
	return item
}

// FirstWith 返回集合中第一个满足条件的元素
func (q LinQuery) FirstWith(predicate func(interface{}) bool) interface{} {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if predicate(item) {
			return item
		}
	}

	return nil
}

func (q LinQuery) FirstWithT(predicateFn interface{}) interface{} {

	predicateGenericFunc, err := newGenericFunc(
		"FirstWithT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.FirstWith(predicateFunc)
}

// ForEach 以集合的每个元素为基础执行指定的操作
func (q LinQuery) ForEach(action func(interface{})) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		action(item)
	}
}

func (q LinQuery) ForEachT(actionFn interface{}) {
	actionGenericFunc, err := newGenericFunc(
		"ForEachT", "actionFn", actionFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), nil),
	)

	if err != nil {
		panic(err)
	}

	actionFunc := func(item interface{}) {
		actionGenericFunc.Call(item)
	}

	q.ForEach(actionFunc)
}

// ForEachIndexed 以集合的每个元素为基础执行指定的操作, 并且提供元素的索引
func (q LinQuery) ForEachIndexed(action func(int, interface{})) {
	next := q.Iterate()
	index := 0

	for item, ok := next(); ok; item, ok = next() {
		action(index, item)
		index++
	}
}

func (q LinQuery) ForEachIndexedT(actionFn interface{}) {
	actionGenericFunc, err := newGenericFunc(
		"ForEachIndexedT", "actionFn", actionFn,
		simpleParamValidator(newElemTypeSlice(new(int), new(genericType)), nil),
	)

	if err != nil {
		panic(err)
	}

	actionFunc := func(index int, item interface{}) {
		actionGenericFunc.Call(index, item)
	}

	q.ForEachIndexed(actionFunc)
}

// Last 返回集合的最后一个元素
func (q LinQuery) Last() (r interface{}) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		r = item
	}

	return
}

// LastWith 返回集合中最后一个满足条件的元素
func (q LinQuery) LastWith(predicate func(interface{}) bool) (r interface{}) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if predicate(item) {
			r = item
		}
	}

	return
}

func (q LinQuery) LastWithT(predicateFn interface{}) interface{} {

	predicateGenericFunc, err := newGenericFunc(
		"LastWithT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.LastWith(predicateFunc)
}

// Max 返回值集合中的最大值
func (q LinQuery) Max() (r interface{}) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return nil
	}

	compare := getComparer(item)
	r = item

	for item, ok := next(); ok; item, ok = next() {
		if compare(item, r) > 0 {
			r = item
		}
	}

	return
}

// Min 返回值集合中的最小值
func (q LinQuery) Min() (r interface{}) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return nil
	}

	compare := getComparer(item)
	r = item

	for item, ok := next(); ok; item, ok = next() {
		if compare(item, r) < 0 {
			r = item
		}
	}

	return
}

// Results 迭代集合并返回SLICE
func (q LinQuery) Results() (r []interface{}) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		r = append(r, item)
	}

	return
}

// SequenceEqual 通过迭代判断两个集合是否相等
func (q LinQuery) SequenceEqual(q2 LinQuery) bool {
	next := q.Iterate()
	next2 := q2.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		item2, ok2 := next2()
		if !ok2 || !equal(item, item2) {
			return false
		}
	}

	_, ok2 := next2()
	return !ok2
}

// SumInt 计算和
func (q LinQuery) SumInt() (r int64) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return 0
	}

	conv := getIntConverter(item)
	r = conv(item)

	for item, ok = next(); ok; item, ok = next() {
		r += conv(item)
	}

	return
}

// SumUInt 计算和
func (q LinQuery) SumUInt() (r uint64) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return 0
	}

	conv := getUIntConverter(item)
	r = conv(item)

	for item, ok = next(); ok; item, ok = next() {
		r += conv(item)
	}

	return
}

// SumFloat 计算和
func (q LinQuery) SumFloat() (r float64) {
	next := q.Iterate()
	item, ok := next()
	if !ok {
		return 0
	}

	conv := getFloatConverter(item)
	r = conv(item)

	for item, ok = next(); ok; item, ok = next() {
		r += conv(item)
	}

	return
}

// ToChannel 迭代集合并返回CHANNEL
func (q LinQuery) ToChannel(result chan<- interface{}) {
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		result <- item
	}

	close(result)
}

func (q LinQuery) ToChannelT(result interface{}) {
	r := reflect.ValueOf(result)
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		r.Send(reflect.ValueOf(item))
	}

	r.Close()
}

// ToMap 迭代集合并返回 map
func (q LinQuery) ToMap(result interface{}) {
	q.ToMapBy(
		result,
		func(i interface{}) interface{} {
			return i.(KeyValue).Key
		},
		func(i interface{}) interface{} {
			return i.(KeyValue).Value
		})
}

// ToMapBy 迭代集合并返回 map
func (q LinQuery) ToMapBy(result interface{},
	keySelector func(interface{}) interface{},
	valueSelector func(interface{}) interface{}) {
	res := reflect.ValueOf(result)
	m := reflect.Indirect(res)
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		key := reflect.ValueOf(keySelector(item))
		value := reflect.ValueOf(valueSelector(item))

		m.SetMapIndex(key, value)
	}

	res.Elem().Set(m)
}

func (q LinQuery) ToMapByT(result interface{},
	keySelectorFn interface{}, valueSelectorFn interface{}) {
	keySelectorGenericFunc, err := newGenericFunc(
		"ToMapByT", "keySelectorFn", keySelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	keySelectorFunc := func(item interface{}) interface{} {
		return keySelectorGenericFunc.Call(item)
	}

	valueSelectorGenericFunc, err := newGenericFunc(
		"ToMapByT", "valueSelectorFn", valueSelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	valueSelectorFunc := func(item interface{}) interface{} {
		return valueSelectorGenericFunc.Call(item)
	}

	q.ToMapBy(result, keySelectorFunc, valueSelectorFunc)
}

// ToSlice 迭代集合并返回 SLICE
func (q LinQuery) ToSlice(v interface{}) {
	res := reflect.ValueOf(v)
	slice := reflect.Indirect(res)

	cap := slice.Cap()
	res.Elem().Set(slice.Slice(0, cap)) // make len(slice)==cap(slice) from now on

	next := q.Iterate()
	index := 0
	for item, ok := next(); ok; item, ok = next() {
		if index >= cap {
			slice, cap = grow(slice)
		}
		slice.Index(index).Set(reflect.ValueOf(item))
		index++
	}

	// reslice the len(res)==cap(res) actual res size
	res.Elem().Set(slice.Slice(0, index))
}
