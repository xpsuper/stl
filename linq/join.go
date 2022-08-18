package linq

// Join 连接两个集合
// outerKeySelector outer 的连接键
// innerKeySelector inner 的连接键
// resultSelector 连接后的结果
func (q LinQuery) Join(inner LinQuery,
	outerKeySelector func(interface{}) interface{},
	innerKeySelector func(interface{}) interface{},
	resultSelector func(outer interface{}, inner interface{}) interface{}) LinQuery {

	return LinQuery{
		Iterate: func() Iterator {
			outernext := q.Iterate()
			innernext := inner.Iterate()

			innerLookup := make(map[interface{}][]interface{})
			for innerItem, ok := innernext(); ok; innerItem, ok = innernext() {
				innerKey := innerKeySelector(innerItem)
				innerLookup[innerKey] = append(innerLookup[innerKey], innerItem)
			}

			var outerItem interface{}
			var innerGroup []interface{}
			innerLen, innerIndex := 0, 0

			return func() (item interface{}, ok bool) {
				if innerIndex >= innerLen {
					has := false
					for !has {
						outerItem, ok = outernext()
						if !ok {
							return
						}

						innerGroup, has = innerLookup[outerKeySelector(outerItem)]
						innerLen = len(innerGroup)
						innerIndex = 0
					}
				}

				item = resultSelector(outerItem, innerGroup[innerIndex])
				innerIndex++
				return item, true
			}
		},
	}
}

func (q LinQuery) JoinT(inner LinQuery,
	outerKeySelectorFn interface{},
	innerKeySelectorFn interface{},
	resultSelectorFn interface{}) LinQuery {
	outerKeySelectorGenericFunc, err := newGenericFunc(
		"JoinT", "outerKeySelectorFn", outerKeySelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	outerKeySelectorFunc := func(item interface{}) interface{} {
		return outerKeySelectorGenericFunc.Call(item)
	}

	innerKeySelectorFuncGenericFunc, err := newGenericFunc(
		"JoinT", "innerKeySelectorFn",
		innerKeySelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	innerKeySelectorFunc := func(item interface{}) interface{} {
		return innerKeySelectorFuncGenericFunc.Call(item)
	}

	resultSelectorGenericFunc, err := newGenericFunc(
		"JoinT", "resultSelectorFn", resultSelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	resultSelectorFunc := func(outer interface{}, inner interface{}) interface{} {
		return resultSelectorGenericFunc.Call(outer, inner)
	}

	return q.Join(inner, outerKeySelectorFunc, innerKeySelectorFunc, resultSelectorFunc)
}
