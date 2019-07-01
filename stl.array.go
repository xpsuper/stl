package stl

import "math/rand"

type XPArrayImpl struct {

}

func randomInt64(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}

	return rand.Int63n(max-min) + min
}

func (instance *XPArrayImpl) Merge(arr ...[]interface{}) (array []interface{}) {
	switch len(arr) {
	case 0:
		break
	case 1:
		array = arr[0]
		break
	default:
		arr1 := arr[0]
		arr2 := instance.Merge(arr[1:]...)//...将数组元素打散
		array = make([]interface{}, len(arr1)+len(arr2))
		copy(array, arr1)
		copy(array[len(arr1):], arr2)
		break
	}

	return
}

func (instance *XPArrayImpl) Random(array []interface{}) []interface{} {
	for i := len(array) - 1; i >= 0; i-- {
		p := randomInt64(0, int64(i))
		a := array[i]
		array[i] = array[p]
		array[p] = a
	}
	return array
}