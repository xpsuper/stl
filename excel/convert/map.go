package convert

import (
	"reflect"
)

// 把map转成map[string]interface{}，key的值使用MustString计算。
// 如果子项中也有map，则继续递归执行直到全部转换为map[string]interface{}
// 如果子项有[]interface{}，则要继续判定slice的元素中的类型
// 常用于各种xml\yaml\json转换为map的结果的统一处理。
func MustMapStringInterfaceRecursions(leafMap interface{}) map[string]interface{} {
	leafType := reflect.TypeOf(leafMap)
	if leafType.Kind() != reflect.Map {
		return nil
	}
	leafValue := reflect.ValueOf(leafMap)
	if leafValue.Len() == 0 {
		return nil
	}

	resMap := make(map[string]interface{})
	leafKeyValues := leafValue.MapKeys()
	// key的value
	for _, leafKeyValue := range leafKeyValues {
		// node的value
		nodeValue := leafValue.MapIndex(leafKeyValue)

		// 获得实际的key和node
		k := leafKeyValue.Interface()
		node := nodeValue.Interface()
		if nodeValue.IsNil() {
			continue
		}

		strKey := MustString(k)
		nodeType := reflect.TypeOf(node)

		switch nodeType.Kind() {
		case reflect.Map:
			temp := MustMapStringInterfaceRecursions(node)
			if temp != nil {
				resMap[strKey] = temp
			}
		case reflect.Slice, reflect.Array:
			temp := MustMapStringInterfaceRecursionsInArrayInterface(node)
			if temp != nil {
				resMap[strKey] = temp
			}
		default:
			resMap[strKey] = node
		}
	}
	return resMap
}

// 协助处理[]interface{}中的map[interface{}]interface{}为map[string]interface{}
func MustMapStringInterfaceRecursionsInArrayInterface(leafAry interface{}) []interface{} {
	leafType := reflect.TypeOf(leafAry)
	if leafType.Kind() != reflect.Array &&
		leafType.Kind() != reflect.Slice {
		return nil
	}
	leafValue := reflect.ValueOf(leafAry)
	if leafValue.Len() == 0 {
		return nil
	}
	resAry := make([]interface{}, 0)

	for i := 0; i < leafValue.Len(); i++ {
		nodeValue := leafValue.Index(i)
		// 获得实际的key和node

		node := nodeValue.Interface()
		nodeType := reflect.TypeOf(node)

		switch nodeType.Kind() {
		case reflect.Array, reflect.Slice:
			temp := MustMapStringInterfaceRecursionsInArrayInterface(node)
			if temp != nil {
				resAry = append(resAry, temp)
			}
		case reflect.Map:
			temp := MustMapStringInterfaceRecursions(node)
			if temp != nil {
				resAry = append(resAry, temp)
			}
		default:
			resAry = append(resAry, node)
		}
	}

	return resAry
}
