package linq

// Skip 跳过 count 个元素，返回剩下的元素
func (q LinQuery) Skip(count int) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			n := count

			return func() (item interface{}, ok bool) {
				for ; n > 0; n-- {
					item, ok = next()
					if !ok {
						return
					}
				}

				return next()
			}
		},
	}
}

// SkipWhile 跳过元素直到 predicate 返回true，返回剩下的元素
func (q LinQuery) SkipWhile(predicate func(interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			ready := false

			return func() (item interface{}, ok bool) {
				for !ready {
					item, ok = next()
					if !ok {
						return
					}

					ready = !predicate(item)
					if ready {
						return
					}
				}

				return next()
			}
		},
	}
}

func (q LinQuery) SkipWhileT(predicateFn interface{}) LinQuery {

	predicateGenericFunc, err := newGenericFunc(
		"SkipWhileT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(item interface{}) bool {
		return predicateGenericFunc.Call(item).(bool)
	}

	return q.SkipWhile(predicateFunc)
}

// SkipWhileIndexed 跳过元素直到 predicate 返回true，返回剩下的元素, predicate 方法会提供索引
func (q LinQuery) SkipWhileIndexed(predicate func(int, interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			ready := false
			index := 0

			return func() (item interface{}, ok bool) {
				for !ready {
					item, ok = next()
					if !ok {
						return
					}

					ready = !predicate(index, item)
					if ready {
						return
					}

					index++
				}

				return next()
			}
		},
	}
}

func (q LinQuery) SkipWhileIndexedT(predicateFn interface{}) LinQuery {
	predicateGenericFunc, err := newGenericFunc(
		"SkipWhileIndexedT", "predicateFn", predicateFn,
		simpleParamValidator(newElemTypeSlice(new(int), new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	predicateFunc := func(index int, item interface{}) bool {
		return predicateGenericFunc.Call(index, item).(bool)
	}

	return q.SkipWhileIndexed(predicateFunc)
}
