package linq

type order struct {
	selector func(interface{}) interface{}
	compare  comparer
	desc     bool
}

type OrderedQuery struct {
	LinQuery
	original LinQuery
	orders   []order
}

// OrderBy 以 selector 提供的值升序排序
func (q LinQuery) OrderBy(selector func(interface{}) interface{}) OrderedQuery {
	return OrderedQuery{
		orders:   []order{{selector: selector}},
		original: q,
		LinQuery: LinQuery{
			Iterate: func() Iterator {
				items := q.sort([]order{{selector: selector}})
				len := len(items)
				index := 0

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						item = items[index]
						index++
					}

					return
				}
			},
		},
	}
}

func (q LinQuery) OrderByT(selectorFn interface{}) OrderedQuery {
	selectorGenericFunc, err := newGenericFunc(
		"OrderByT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectorGenericFunc.Call(item)
	}

	return q.OrderBy(selectorFunc)
}

// OrderByDescending 以 selector 提供的值升序排序
func (q LinQuery) OrderByDescending(selector func(interface{}) interface{}) OrderedQuery {
	return OrderedQuery{
		orders:   []order{{selector: selector, desc: true}},
		original: q,
		LinQuery: LinQuery{
			Iterate: func() Iterator {
				items := q.sort([]order{{selector: selector, desc: true}})
				len := len(items)
				index := 0

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						item = items[index]
						index++
					}

					return
				}
			},
		},
	}
}

func (q LinQuery) OrderByDescendingT(selectorFn interface{}) OrderedQuery {
	selectorGenericFunc, err := newGenericFunc(
		"OrderByDescendingT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectorGenericFunc.Call(item)
	}

	return q.OrderByDescending(selectorFunc)
}

// ThenBy 继续以 selector 提供的值升序排序
func (oq OrderedQuery) ThenBy(
	selector func(interface{}) interface{}) OrderedQuery {
	return OrderedQuery{
		orders:   append(oq.orders, order{selector: selector}),
		original: oq.original,
		LinQuery: LinQuery{
			Iterate: func() Iterator {
				items := oq.original.sort(append(oq.orders, order{selector: selector}))
				len := len(items)
				index := 0

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						item = items[index]
						index++
					}

					return
				}
			},
		},
	}
}

func (oq OrderedQuery) ThenByT(selectorFn interface{}) OrderedQuery {
	selectorGenericFunc, err := newGenericFunc(
		"ThenByT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectorGenericFunc.Call(item)
	}

	return oq.ThenBy(selectorFunc)
}

// ThenByDescending 继续以 selector 提供的值降序排序
func (oq OrderedQuery) ThenByDescending(selector func(interface{}) interface{}) OrderedQuery {
	return OrderedQuery{
		orders:   append(oq.orders, order{selector: selector, desc: true}),
		original: oq.original,
		LinQuery: LinQuery{
			Iterate: func() Iterator {
				items := oq.original.sort(append(oq.orders, order{selector: selector, desc: true}))
				len := len(items)
				index := 0

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						item = items[index]
						index++
					}

					return
				}
			},
		},
	}
}

func (oq OrderedQuery) ThenByDescendingT(selectorFn interface{}) OrderedQuery {
	selectorFunc, ok := selectorFn.(func(interface{}) interface{})
	if !ok {
		selectorGenericFunc, err := newGenericFunc(
			"ThenByDescending", "selectorFn", selectorFn,
			simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
		)
		if err != nil {
			panic(err)
		}

		selectorFunc = func(item interface{}) interface{} {
			return selectorGenericFunc.Call(item)
		}
	}
	return oq.ThenByDescending(selectorFunc)
}
