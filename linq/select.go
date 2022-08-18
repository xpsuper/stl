package linq

// Select 将每个元素通过 selector 函数转换为新的元素，并返回一个新的 Query
func (q LinQuery) Select(selector func(interface{}) interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			return func() (item interface{}, ok bool) {
				var it interface{}
				it, ok = next()
				if ok {
					item = selector(it)
				}

				return
			}
		},
	}
}

func (q LinQuery) SelectT(selectorFn interface{}) LinQuery {

	selectGenericFunc, err := newGenericFunc(
		"SelectT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectGenericFunc.Call(item)
	}

	return q.Select(selectorFunc)
}

// SelectIndexed 将每个元素通过 selector 函数转换为新的元素，并返回一个新的 Query, selector 函数会按照索引顺序返回
func (q LinQuery) SelectIndexed(selector func(int, interface{}) interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			index := 0

			return func() (item interface{}, ok bool) {
				var it interface{}
				it, ok = next()
				if ok {
					item = selector(index, it)
					index++
				}

				return
			}
		},
	}
}

func (q LinQuery) SelectIndexedT(selectorFn interface{}) LinQuery {
	selectGenericFunc, err := newGenericFunc(
		"SelectIndexedT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(int), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(index int, item interface{}) interface{} {
		return selectGenericFunc.Call(index, item)
	}

	return q.SelectIndexed(selectorFunc)
}
