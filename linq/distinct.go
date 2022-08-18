package linq

// Distinct 去重,返回不包含重复值的无序集合
func (q LinQuery) Distinct() LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			set := make(map[interface{}]bool)

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					if _, has := set[item]; !has {
						set[item] = true
						return
					}
				}

				return
			}
		},
	}
}

func (oq OrderedQuery) Distinct() OrderedQuery {
	return OrderedQuery{
		orders: oq.orders,
		LinQuery: LinQuery{
			Iterate: func() Iterator {
				next := oq.Iterate()
				var prev interface{}

				return func() (item interface{}, ok bool) {
					for item, ok = next(); ok; item, ok = next() {
						if item != prev {
							prev = item
							return
						}
					}

					return
				}
			},
		},
	}
}

// DistinctBy 去重,返回不包含重复值的无序集合,以 selector 提供的值作为去重依据
func (q LinQuery) DistinctBy(selector func(interface{}) interface{}) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			next := q.Iterate()
			set := make(map[interface{}]bool)

			return func() (item interface{}, ok bool) {
				for item, ok = next(); ok; item, ok = next() {
					s := selector(item)
					if _, has := set[s]; !has {
						set[s] = true
						return
					}
				}

				return
			}
		},
	}
}

func (q LinQuery) DistinctByT(selectorFn interface{}) LinQuery {
	selectorFunc, ok := selectorFn.(func(interface{}) interface{})
	if !ok {
		selectorGenericFunc, err := newGenericFunc(
			"DistinctByT", "selectorFn", selectorFn,
			simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
		)
		if err != nil {
			panic(err)
		}

		selectorFunc = func(item interface{}) interface{} {
			return selectorGenericFunc.Call(item)
		}
	}
	return q.DistinctBy(selectorFunc)
}
