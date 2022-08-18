package linq

// Aggregate 在序列上应用累加寄存器方法
//
// Example:
//
// fruits := []string{"apple", "mango", "orange", "passion-fruit", "grape"}
//
// Determine which string in the slice is the longest.
// longestName := From(fruits).Aggregate(
// 		func(r interface{}, i interface{}) interface{} {
//			if len(r.(string)) > len(i.(string)) {
//				return r
//			}
//			return i
//		},
//	)
//
//	fmt.Println(longestName)
//  Output: passion-fruit
//
func (q LinQuery) Aggregate(f func(interface{}, interface{}) interface{}) interface{} {
	next := q.Iterate()

	result, any := next()
	if !any {
		return nil
	}

	for current, ok := next(); ok; current, ok = next() {
		result = f(result, current)
	}

	return result
}

func (q LinQuery) AggregateT(f interface{}) interface{} {
	fGenericFunc, err := newGenericFunc(
		"AggregateT", "f", f,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	fFunc := func(result interface{}, current interface{}) interface{} {
		return fGenericFunc.Call(result, current)
	}

	return q.Aggregate(fFunc)
}

// AggregateWithSeed 在序列上应用累加寄存器方法，可设置初始值 seed
func (q LinQuery) AggregateWithSeed(seed interface{},
	f func(interface{}, interface{}) interface{}) interface{} {

	next := q.Iterate()
	result := seed

	for current, ok := next(); ok; current, ok = next() {
		result = f(result, current)
	}

	return result
}

func (q LinQuery) AggregateWithSeedT(seed interface{},
	f interface{}) interface{} {
	fGenericFunc, err := newGenericFunc(
		"AggregateWithSeed", "f", f,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	fFunc := func(result interface{}, current interface{}) interface{} {
		return fGenericFunc.Call(result, current)
	}

	return q.AggregateWithSeed(seed, fFunc)
}

// AggregateWithSeedBy 在序列上应用累加寄存器方法，可设置初始值 seed，并通过 resultSelector 返回结果
//
// Example:
//
//	input := []string{"apple", "mango", "orange", "passion-fruit", "grape"}
//
//	//确定数组中的任何字符串是否长于“banana”
//	longestName := From(input).
//	AggregateWithSeedBy("banana",
//		func(longest interface{}, next interface{}) interface{} {
//			if len(longest.(string)) > len(next.(string)) {
//				return longest
//			}
//			return next
//		},
//		// 返回最终结果
//		func(result interface{}) interface{} {
//			return fmt.Sprintf("The fruit with the longest name is %s.", result)
//		},
//	)
//
//	fmt.Println(longestName)
//  Output: The fruit with the longest name is passion-fruit.
//
func (q LinQuery) AggregateWithSeedBy(seed interface{},
	f func(interface{}, interface{}) interface{},
	resultSelector func(interface{}) interface{}) interface{} {

	next := q.Iterate()
	result := seed

	for current, ok := next(); ok; current, ok = next() {
		result = f(result, current)
	}

	return resultSelector(result)
}

func (q LinQuery) AggregateWithSeedByT(seed interface{},
	f interface{},
	resultSelectorFn interface{}) interface{} {
	fGenericFunc, err := newGenericFunc(
		"AggregateWithSeedByT", "f", f,
		simpleParamValidator(newElemTypeSlice(new(genericType), new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	fFunc := func(result interface{}, current interface{}) interface{} {
		return fGenericFunc.Call(result, current)
	}

	resultSelectorGenericFunc, err := newGenericFunc(
		"AggregateWithSeedByT", "resultSelectorFn", resultSelectorFn,
		simpleParamValidator(newElemTypeSlice(new(genericType)), newElemTypeSlice(new(genericType))),
	)
	if err != nil {
		panic(err)
	}

	resultSelectorFunc := func(result interface{}) interface{} {
		return resultSelectorGenericFunc.Call(result)
	}

	return q.AggregateWithSeedBy(seed, fFunc, resultSelectorFunc)
}
