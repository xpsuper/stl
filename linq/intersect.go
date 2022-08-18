package linq

// Intersect 返回交集 (同时包含在第一个集合和第二个集合中的元素)
func (q LinQuery) Intersect(q2 LinQuery) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			next2 := q2.Iterate()

			set := make(map[interface{}]bool)
			for item, ok := next2(); ok; item, ok = next2() {
				set[item] = true
			}

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if _, has := set[item]; has {
						delete(set, item)
						return
					}
				}

				return
			}
		},
	}
}

// IntersectBy 返回交集 (同时包含在第一个集合和第二个集合中的元素) 以 selector 提供的值作为判断依据
func (q LinQuery) IntersectBy(q2 LinQuery,
	selector func(interface{}) interface{}) LinQuery {

	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			next2 := q2.Iterate()

			set := make(map[interface{}]bool)
			for item, ok := next2(); ok; item, ok = next2() {
				s := selector(item)
				set[s] = true
			}

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					s := selector(item)
					if _, has := set[s]; has {
						delete(set, s)
						return
					}
				}

				return
			}
		},
	}
}

func (q LinQuery) IntersectByT(q2 LinQuery,
	selectorFn interface{}) LinQuery {
	selectorGenericFunc, err := newGenericFunc(
		"IntersectByT", "selectorFn", selectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	selectorFunc := func(item interface{}) interface{} {
		return selectorGenericFunc.Call(item)
	}

	return q.IntersectBy(q2, selectorFunc)
}
