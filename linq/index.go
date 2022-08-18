package linq

// IndexOf 返回符合 predicate 条件的第一个元素的索引，如果没有找到，则返回 -1
func (q LinQuery) IndexOf(predicate func(interface{}) bool) int {
	index := 0
	next := q.Iterate()

	for item, ok := next(); ok; item, ok = next() {
		if predicate(item) {
			return index
		}
		index++
	}

	return -1
}

func (q LinQuery) IndexOfT(predicateFn interface{}) int {

	predicateGenericFunc, err := newGenericFunc(
		"IndexOfT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.IndexOf(predicateFunc)
}
