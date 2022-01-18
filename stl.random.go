package stl

import (
	"math/rand"
	"time"
)

type XPRandomImpl struct{}

type RandomStringOption func(options *RandomStringOptions)
type RandomStringOptions struct {
	numberOnly bool
	seed       string
}

func loadRandomStringOpts(opts []RandomStringOption) *RandomStringOptions {
	options := &RandomStringOptions{
		numberOnly: false,
		seed:       "",
	}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func (instance *XPRandomImpl) WithNumOnly(numberOnly bool) RandomStringOption {
	return func(options *RandomStringOptions) {
		options.numberOnly = numberOnly
	}
}

func (instance *XPRandomImpl) WithCustomSeed(seed string) RandomStringOption {
	return func(options *RandomStringOptions) {
		options.seed = seed
	}
}

func (instance *XPRandomImpl) RandomString(length int, opts ...RandomStringOption) string {
	if length == 0 {
		return ""
	}

	options := loadRandomStringOpts(opts)

	var seed = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	if options.numberOnly {
		seed = []byte("0123456789")
	} else if options.seed != "" {
		seed = []byte(options.seed)
	}

	seedLen := len(seed)
	if seedLen < 2 || seedLen > 256 {
		panic("Wrong charset length for NewLenChars()")
	}

	max := 255 - (256 % seedLen)
	b := make([]byte, length)
	r := make([]byte, length+(length/4))
	i := 0

	for {
		if _, e := rand.Read(r); e != nil {
			return ""
		}

		for _, rb := range r {
			c := int(rb)
			if c > max {
				continue
			}
			b[i] = seed[c%seedLen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

// 获取范围为[0.0, 1.0)，类型为float32的随机小数

func (instance *XPRandomImpl) RandomFloat32() float32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float32()
}

// 获取范围为[0.0, 1.0)，类型为float64的随机小数

func (instance *XPRandomImpl) RandomFloat64() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Float64()
}

func (instance *XPRandomImpl) RandomInt(max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max)
}

func (instance *XPRandomImpl) RandomIntRange(min, max int) int {
	if min >= max || max == 0 {
		return max
	}

	return rand.Intn(max-min) + min
}

func (instance *XPRandomImpl) RandomInt32(max int32) int32 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int31n(max)
}

func (instance *XPRandomImpl) RandomInt32Range(min, max int32) int32 {
	if min >= max || max == 0 {
		return max
	}

	return rand.Int31n(max-min) + min
}

func (instance *XPRandomImpl) RandomInt64(max int64) int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Int63n(max)
}

func (instance *XPRandomImpl) RandomInt64Range(min, max int64) int64 {
	if min >= max || max == 0 {
		return max
	}

	return rand.Int63n(max-min) + min
}

func (instance *XPRandomImpl) RandomArray(array []interface{}) []interface{} {
	for i := len(array) - 1; i >= 0; i-- {
		p := instance.RandomInt64Range(0, int64(i))
		a := array[i]
		array[i] = array[p]
		array[p] = a
	}
	return array
}
