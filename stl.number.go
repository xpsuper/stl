package stl

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

type XPNumberImpl struct {

}

func (instance *XPNumberImpl) Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func (instance *XPNumberImpl) Max(a, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}

func (instance *XPNumberImpl) Min(a, b int64) int64 {
	if a <= b {
		return a
	} else {
		return b
	}
}

func (instance *XPNumberImpl) Round(x float64) (result int64) {
	return int64(math.Floor(x + 0.5))
}

func (instance *XPNumberImpl) EqualsFloat32(a, b float32) bool {
	const ePSINON float32 = 0.00001;
	c := a - b;
	return (c >= - ePSINON) && (c <= ePSINON);
}

func (instance *XPNumberImpl) EqualsFloat64(a, b float64) bool {
	const ePSINON float64 = 0.00000001;
	c := a - b;
	return (c >= - ePSINON) && (c <= ePSINON);
}

// 获取范围为[0.0, 1.0)，类型为float32的随机小数

func (instance *XPNumberImpl) RandomFloat32() float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float32()
}

// 获取范围为[0.0, 1.0)，类型为float64的随机小数

func (instance *XPNumberImpl) RandomFloat64() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float64()
}

func (instance *XPNumberImpl) RandomInt(max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max)
}

func (instance *XPNumberImpl) RandomInt32(max int32) int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int31n(max)
}

func (instance *XPNumberImpl) RandomInt64(max int64) int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63n(max)
}

func (instance *XPNumberImpl) ToInt64(data interface{}) (ok bool, result int64) {
	var err error
	ok = true
	switch data.(type) {
	case string:
		result, err = strconv.ParseInt(data.(string), 10, 64)
		if err != nil {
			ok = false
		}
		break
	case int:
		result = int64(data.(int))
		break
	case int8:
		result = int64(data.(int8))
		break
	case int16:
		result = int64(data.(int16))
		break
	case int32:
		result = int64(data.(int32))
		break
	case int64:
		result = int64(data.(int64))
		break
	case uint8:
		result = int64(data.(uint8))
		break
	case uint16:
		result = int64(data.(uint16))
		break
	case uint32:
		result = int64(data.(uint32))
		break
	case uint64:
		result = int64(data.(uint64))
		break
	case float32:
		result = int64(data.(float32))
		break
	case float64:
		result = int64(data.(float64))
		break
	case []byte:
		//1
		// b_tmp := data.([]byte)
		// b_buf := bytes.NewBuffer(b_tmp)
		// var result_tmp int32
		// err = binary.Read(b_buf, binary.BigEndian, &result_tmp)
		// if err != nil {
		// 	ok = false
		// 	result = 0
		// } else {
		// 	result = int64(result_tmp)
		// }

		//2
		rStr := string(data.([]byte))
		result, err = strconv.ParseInt(rStr, 10, 64)
		if err != nil {
			ok = false
		}
		break
	case bool:
		if data.(bool) {
			result = 1
		} else {
			result = 0
		}
		break
	default:
		ok = false
		result = 0
		break
	}

	return ok, result
}

func (instance *XPNumberImpl) ToInt64Def(data interface{}, defaultValue ...int64) int64 {
	ok, r := instance.ToInt64(data)
	if ok {
		return r
	} else {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
}

func (instance *XPNumberImpl) ToInt(data interface{}) (bool, int) {
	ok, result := instance.ToInt64(data)
	return ok, int(result)
}

func (instance *XPNumberImpl) ToIntDef(data interface{}, defaultValue ...int) int {
	ok, r := instance.ToInt(data)
	if ok {
		return r
	} else {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
}

func (instance *XPNumberImpl) ToFloat64(data interface{}) (ok bool, result float64) {
	var err error
	ok = true
	switch data.(type) {
	case string:
		result, err = strconv.ParseFloat(data.(string), 64)
		if err != nil {
			ok = false
		}
		break
	case int:
		result = float64(data.(int))
		break
	case int8:
		result = float64(data.(int8))
		break
	case int16:
		result = float64(data.(int16))
		break
	case int32:
		result = float64(data.(int32))
		break
	case int64:
		result = float64(data.(int64))
		break
	case uint8:
		result = float64(data.(uint8))
		break
	case uint16:
		result = float64(data.(uint16))
		break
	case uint32:
		result = float64(data.(uint32))
		break
	case uint64:
		result = float64(data.(uint64))
		break
	case float32:
		result = float64(data.(float32))
		break
	case float64:
		result = data.(float64)
		break
	case []byte:
		rStr := string(data.([]byte))
		result, err = strconv.ParseFloat(rStr, 64)
		if err != nil {
			ok = false
		}
		break
	case bool:
		if data.(bool) {
			result = 1
		} else {
			result = 0
		}
		break
	default:
		ok = false
		result = 0
		break
	}
	return ok, result
}

func (instance *XPNumberImpl) ToFloat64Def(data interface{}, defaultValue ...float64) float64 {
	ok, r := instance.ToFloat64(data)
	if ok {
		return r
	} else {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
}

func (instance *XPNumberImpl) ToFloat32(data interface{}) (bool, float32) {
	ok, result := instance.ToFloat64(data)
	return ok, float32(result)
}

func (instance *XPNumberImpl) ToFloat32Def(data interface{}, defaultValue ...float32) float32 {
	ok, r := instance.ToFloat32(data)
	if ok {
		return r
	} else {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
}
