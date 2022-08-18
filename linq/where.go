package linq

// Where 过滤符合 predicate 的元素
func (q LinQuery) Where(predicate func(interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if predicate(item) {
						return
					}
				}

				return
			}
		},
	}
}

func (q LinQuery) WhereT(predicateFn interface{}) LinQuery {

	predicateGenericFunc, err := newGenericFunc(
		"WhereT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.Where(predicateFunc)
}

// WhereIndexed 过滤符合 predicate 的元素, predicate 方法中会提供索引
func (q LinQuery) WhereIndexed(predicate func(int, interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			index := 0

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if predicate(index, item) {
						index++
						return
					}

					index++
				}

				return
			}
		},
	}
}

func (q LinQuery) WhereIndexedT(predicateFn interface{}) LinQuery {
	predicateGenericFunc, err := newGenericFunc(
		"WhereIndexedT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(int), new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(index int, item interface{}) bool {
		return predicateGenericFunc.Call(index, item).(bool)
	}

	return q.WhereIndexed(predicateFunc)
}
