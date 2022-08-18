package linq

// Sort 通过提供 Less 比较函数来排序,返回一个排序后的 Query, 性能比 OrderBy,
// OrderByDescending, ThenBy 和 ThenByDescending 要好.
func (q LinQuery) Sort(less func(i, j interface{}) bool) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			items := q.lessSort(less)
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
	}
}

func (q LinQuery) SortT(lessFn interface{}) LinQuery {
	lessGenericFunc, err := newGenericFunc(
		"SortT", "lessFn", lessFn,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(bool))),
	)
	if err != nil {
		panic(err)
	}

	lessFunc := func(i, j interface{}) bool {
		return lessGenericFunc.Call(i, j).(bool)
	}

	return q.Sort(lessFunc)
}
