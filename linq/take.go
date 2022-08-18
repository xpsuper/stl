package linq

// Take 从头开始选取指定数量的元素
func (q LinQuery) Take(count int) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			n := count

			return func() (item interface{}, ok bool) {
				if n <= 0 {
					return
				}

				n--
				return next()
			}
		},
	}
}

// TakeWhile 从头开始选取元素直到 满足 predicate 条件(predicate 返回 true)
func (q LinQuery) TakeWhile(predicate func(interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			done := false

			return func() (item interface{}, ok bool) {
				if done {
					return
				}

				item, ok = next()
				if !ok {
					done = true
					return
				}

				if predicate(item) {
					return
				}

				done = true
				return nil, false
			}
		},
	}
}

func (q LinQuery) TakeWhileT(predicateFn interface{}) LinQuery {

	predicateGenericFunc, err := newGenericFunc(
		"TakeWhileT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.TakeWhile(predicateFunc)
}

// TakeWhileIndexed 从头开始选取元素直到 满足 predicate 条件(predicate 返回 true) predicate 中会返回索引
func (q LinQuery) TakeWhileIndexed(predicate func(int, interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			done := false
			index := 0

			return func() (item interface{}, ok bool) {
				if done {
					return
				}

				item, ok = next()
				if !ok {
					done = true
					return
				}

				if predicate(index, item) {
					index++
					return
				}

				done = true
				return nil, false
			}
		},
	}
}

func (q LinQuery) TakeWhileIndexedT(predicateFn interface{}) LinQuery {
	whereFunc, err := newGenericFunc(
		"TakeWhileIndexedT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(int), new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(index int, item interface{}) bool {
		return whereFunc.Call(index, item).(bool)
	}

	return q.TakeWhileIndexed(predicateFunc)
}
