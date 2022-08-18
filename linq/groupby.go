package linq

// Group 用于存储GroupBy方法结果的类型
type Group struct {
	Key   interface{}
	Group []interface{}
}

// GroupBy 根据指定的值对集合的元素进行分组
// keySelector 键选择器提供分组依据
// elementSelector 值选择器提供并入分组的元素
// Example:
//		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
//		q := From(input).GroupBy(
//			func(i interface{}) interface{} { return i.(int) % 2 },
//			func(i interface{}) interface{} { return i.(int) })
//
//		fmt.Println(q.OrderBy(func(i interface{}) interface{} {
//			return i.(Group).Key
//		}).Results())
//
//		Output:
//		[{0 [2 4 6 8]} {1 [1 3 5 7 9]}]

func (q LinQuery) GroupBy(keySelector func(interface{}) interface{},
	elementSelector func(interface{}) interface{}) LinQuery {
	return LinQuery{
		q.Source,
		func() Iterator {
			next := q.Iterate()
			set := make(map[interface{}][]interface{})

			for item, ok := next(); ok; item, ok = next() {
				key := keySelector(item)
				set[key] = append(set[key], elementSelector(item))
			}

			len := len(set)
			idx := 0
			groups := make([]Group, len)
			for k, v := range set {
				groups[idx] = Group{k, v}
				idx++
			}

			index := 0

			return func() (item interface{}, ok bool) {
				ok = index < len
				if ok {
					item = groups[index]
					index++
				}

				return
			}
		},
	}
}

func (q LinQuery) GroupByT(keySelectorFn interface{},
	elementSelectorFn interface{}) LinQuery {
	keySelectorGenericFunc, err := newGenericFunc(
		"GroupByT", "keySelectorFn", keySelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	keySelectorFunc := func(item interface{}) interface{} {
		return keySelectorGenericFunc.Call(item)
	}

	elementSelectorGenericFunc, err := newGenericFunc(
		"GroupByT", "elementSelectorFn", elementSelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	elementSelectorFunc := func(item interface{}) interface{} {
		return elementSelectorGenericFunc.Call(item)

	}

	return q.GroupBy(keySelectorFunc, elementSelectorFunc)
}
