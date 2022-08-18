package linq

// Zip 以短集合的长度为标准实现压缩
// resultSelector 用来自两个集合的元素返回一个新的元素
func (q LinQuery) Zip(q2 LinQuery,
	resultSelector func(interface{}, interface{}) interface{}) LinQuery {

	return LinQuery{
		Iterate: func() Iterator {
			next1 := q.Iterate()
			next2 := q2.Iterate()

			return func() (item interface{}, ok bool) {
				item1, ok1 := next1()
				item2, ok2 := next2()

				if ok1 && ok2 {
					return resultSelector(item1, item2), true
				}

				return nil, false
			}
		},
	}
}

func (q LinQuery) ZipT(q2 LinQuery,
	resultSelectorFn interface{}) LinQuery {
	resultSelectorGenericFunc, err := newGenericFunc(
		"ZipT", "resultSelectorFn", resultSelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	resultSelectorFunc := func(item1 interface{}, item2 interface{}) interface{} {
		return resultSelectorGenericFunc.Call(item1, item2)
	}

	return q.Zip(q2, resultSelectorFunc)
}
