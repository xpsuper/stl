package stl

import (
	"errors"
	"fmt"
	"math/rand"
)

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

type GoArray struct {
	data []interface{}
	size int
}

func NewGoArray(capacity ...int) (array *GoArray) {
	if len(capacity) >= 1 && capacity[0] != 0 {
		array = &GoArray{
			data: make([]interface{}, capacity[0]),
			size: 0,
		}
	} else {
		array = &GoArray{
			data: make([]interface{}, 10),
			size: 0,
		}
	}

	return
}

//判断索引是否越界
func (array *GoArray) checkIndex(index int) bool {
	if index < 0 || index >= array.size {
		return true
	}

	return false
}

//数组扩容
func (array *GoArray) resize(capacity int) {
	newArray := make([]interface{}, capacity)
	for i := 0; i < array.size; i ++ {
		newArray[i] = array.data[i]
	}
	array.data = newArray
	newArray = nil
}

//获取数组容量
func (array *GoArray) GetCapacity() int {
	return cap(array.data)
}

//获取数组长度
func (array *GoArray) GetSize() int {
	return array.size
}

//判断数组是否为空
func (array *GoArray) IsEmpty() bool {
	return array.size == 0
}

//向数组头插入元素
func (array *GoArray) AddFirst(value interface{}) error {
	return array.Add(0, value)
}

//向数组尾插入元素
func (array *GoArray) AddLast(value interface{}) error {
	return array.Add(array.size, value)
}

//在 index 位置，插入元素e, 时间复杂度 O(m+n)
func (array *GoArray) Add(index int, value interface{}) (err error) {
	if index < 0 || index > array.size {
		err = errors.New("Add failed. Require index >= 0 and index <= size.")
		return
	}

	// 如果当前元素个数等于数组容量，则将数组扩容为原来的2倍
	cap := array.GetCapacity()
	if array.size == cap {
		array.resize(cap * 2)
	}

	for i := array.size - 1; i >= index; i-- {
		array.data[i+1] = array.data[i]
	}

	array.data[index] = value
	array.size++
	return
}

//获取对应 index 位置的元素
func (array *GoArray) Get(index int) (value interface{}, err error) {
	if array.checkIndex(index) {
		err = errors.New("Get failed. Index is illegal.")
		return
	}

	value = array.data[index]
	return
}

//修改 index 位置的元素
func (array *GoArray) Set(index int, value interface{}) (err error) {
	if array.checkIndex(index) {
		err = errors.New("Set failed. Index is illegal.")
		return
	}

	array.data[index] = value
	return
}

//查找数组中是否有元素
func (array *GoArray) Contains(value interface{}) bool {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return true
		}
	}

	return false
}

//通过索引查找数组，索引范围[0,n-1](未找到，返回 -1)
func (array *GoArray) Find(value interface{}) int {
	for i := 0; i < array.size; i++ {
		if array.data[i] == value {
			return i
		}
	}

	return -1
}

// 删除 index 位置的元素，并返回
func (array *GoArray) Remove(index int) (value interface{}, err error) {
	if array.checkIndex(index) {
		err = errors.New("Remove failed. Index is illegal.")
		return
	}

	value = array.data[index]
	for i := index + 1; i < array.size; i++ {
		//数据全部往前挪动一位,覆盖需要删除的元素
		array.data[i-1] = array.data[i]
	}

	array.size--
	array.data[array.size] = nil //loitering objects != memory leak

	cap := array.GetCapacity()
	if array.size == cap/4 && cap/2 != 0 {
		array.resize(cap / 2)
	}
	return
}

//删除数组首个元素
func (array *GoArray) RemoveFirst() (interface{}, error) {
	return array.Remove(0)
}

//删除末尾元素
func (array *GoArray) RemoveLast() (interface{}, error) {
	return array.Remove(int(array.size - 1))
}

//从数组中删除指定元素
func (array *GoArray) RemoveElement(value interface{}) (e interface{}, err error) {
	index := array.Find(value)
	if index != -1 {
		e, err = array.Remove(index)
	}
	return
}

//清空数组
func (array *GoArray) Clear() {
	array.data = make([]interface{}, array.size)
	array.size = 0
}

//打印数列
func (array *GoArray) PrintIn() {
	var format string
	format = fmt.Sprintf("Array: size = %d , capacity = %d\n",array.size, cap(array.data))
	format += "["
	for i := 0; i < array.GetSize(); i++ {
		format += fmt.Sprintf("%+v", array.data[i])
		if i != array.size -1 {
			format += ", "
		}
	}
	format += "]"
	fmt.Println(format)
}