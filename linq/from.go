package linq

import "reflect"

// Iterator 迭代器函数别名
type Iterator func() (item interface{}, ok bool)

// LinQuery Linq 结构
type LinQuery struct {
	Source  reflect.Value
	Iterate func() Iterator
}

// KeyValue 用于在 map 上的迭代数据类型,ToMap 方法也通过此类型输出数据
type KeyValue struct {
	Key   interface{}
	Value interface{}
}

// Iterable 自定义类型中需实现此迭代器接口以支持 Linq 查询
type Iterable interface {
	Iterate() Iterator
}

// From 以 source 为基础初始化一个 LinQuery 查询
func From(source interface{}) LinQuery {
	src := reflect.ValueOf(source)

	switch src.Kind() {
	case reflect.Slice, reflect.Array:
		len := src.Len()

		return LinQuery{
			Source: src,
			Iterate: func() Iterator {
				index := 0

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						item = src.Index(index).Interface()
						index++
					}

					return
				}
			},
		}
	case reflect.Map:
		len := src.Len()

		return LinQuery{
			Source: src,
			Iterate: func() Iterator {
				index := 0
				keys := src.MapKeys()

				return func() (item interface{}, ok bool) {
					ok = index < len
					if ok {
						key := keys[index]
						item = KeyValue{
							Key:   key.Interface(),
							Value: src.MapIndex(key).Interface(),
						}

						index++
					}

					return
				}
			},
		}
	case reflect.String:
		return FromString(source.(string))
	case reflect.Chan:
		if _, ok := source.(chan interface{}); ok {
			return FromChannel(source.(chan interface{}))
		} else {
			return FromChannelT(source)
		}
	default:
		if isEmpty(source) {
			return FromString("")
		}
		return FromIterable(source.(Iterable))
	}
}

func FromChannel(source <-chan interface{}) LinQuery {
	return LinQuery{
		Source: reflect.ValueOf(source),
		Iterate: func() Iterator {
			return func() (item interface{}, ok bool) {
				item, ok = <-source
				return
			}
		},
	}
}

func FromChannelT(source interface{}) LinQuery {
	src := reflect.ValueOf(source)
	return LinQuery{
		Source: src,
		Iterate: func() Iterator {
			return func() (interface{}, bool) {
				value, ok := src.Recv()
				return value.Interface(), ok
			}
		},
	}
}

func FromString(source string) LinQuery {
	runes := []rune(source)
	l := len(runes)

	return LinQuery{
		Source: reflect.ValueOf(source),
		Iterate: func() Iterator {
			index := 0

			return func() (item interface{}, ok bool) {
				ok = index < l
				if ok {
					item = runes[index]
					index++
				}

				return
			}
		},
	}
}

func FromIterable(source Iterable) LinQuery {
	return LinQuery{
		Iterate: source.Iterate,
	}
}

// Range 生成指定范围内的整数序列
func Range(start, count int) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			index := 0
			current := start

			return func() (item interface{}, ok bool) {
				if index >= count {
					return nil, false
				}

				item, ok = current, true

				index++
				current++
				return
			}
		},
	}
}

// Repeat 生成包含一个重复值的序列
func Repeat(value interface{}, count int) LinQuery {
	return LinQuery{
		Iterate: func() Iterator {
			index := 0

			return func() (item interface{}, ok bool) {
				if index >= count {
					return nil, false
				}

				item, ok = value, true

				index++
				return
			}
		},
	}
}
