package linq

// Except 返回集差 (第一个集合中不出现在第二个集合中的元素)
func (q LinQuery) Except(q2 LinQuery) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			next2 := q2.Iterate()
			set := make(map[interface{}]bool)
			for i, ok := next2(); ok; i, ok = next2() {
				set[i] = true
			}

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if _, has := set[item]; !has {
						return
					}
				}

				return
			}
		},
	}
}

// ExceptBy 返回集差 (第一个集合中不出现在第二个集合中的元素) 以 selector 提供的值作为去重依据
func (q LinQuery) ExceptBy(q2 LinQuery,
	selector func(interface{}) interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()

			next2 := q2.Iterate()
			set := make(map[interface{}]bool)
			for i, ok := next2(); ok; i, ok = next2() {
				s := selector(i)
				set[s] = true
			}

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					s := selector(item)
					if _, has := set[s]; !has {
						return
					}
				}

				return
			}
		},
	}
}

func (q LinQuery) ExceptByT(q2 LinQuery,
	selectorFn interface{}) LinQuery {
	selectorGenericFunc, err := newGenericFunc(
		"ExceptByT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectorGenericFunc.Call(item)
	}

	return q.ExceptBy(q2, selectorFunc)
}
