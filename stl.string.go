package stl

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type XPStringImpl struct {

}

func (instance *XPStringImpl) ToString(data interface{}) (ok bool, result string) {
	ok = true
	switch data.(type) {
	case string:
		result = data.(string)
		break
	case int:
		result = strconv.Itoa(data.(int))
		break
	case int8:
		result = strconv.Itoa(int(data.(int8)))
		break
	case int16:
		result = strconv.Itoa(int(data.(int16)))
		break
	case int32:
		result = strconv.Itoa(int(data.(int32)))
		break
	case int64:
		result = strconv.Itoa(int(data.(int64)))
		break
	case uint8:
		result = strconv.FormatUint(uint64(data.(uint8)), 10)
		break
	case uint16:
		result = strconv.FormatUint(uint64(data.(uint16)), 10)
		break
	case uint32:
		result = strconv.FormatUint(uint64(data.(uint32)), 10)
		break
	case uint64:
		result = strconv.FormatUint(data.(uint64), 10)
		break
	case float32:
		result = strconv.FormatFloat(data.(float64), 'f', -1, 32)
		break
	case float64:
		result = strconv.FormatFloat(data.(float64), 'f', -1, 64)
		break
	case []byte:
		result = string(data.([]byte))
	case bool:
		if data.(bool) {
			result = "true"
		} else {
			result = "false"
		}
		break
	default:
		ok = false
		result = ""
		break
	}

	return ok, result
}

func (instance *XPStringImpl) ToStringDef(data interface{}, defaultValue ...string) string {
	ok, r := instance.ToString(data)
	if ok {
		return r
	} else {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		} else {
			return ""
		}
	}
}

func (instance *XPStringImpl) SHA1(str string) string {
	hashS := sha1.New()
	hashS.Write([]byte(str))
	return hex.EncodeToString(hashS.Sum(nil))
}

func (instance *XPStringImpl) CRC32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func (instance *XPStringImpl) MD5(str string) string {
	hashS := md5.New()
	hashS.Write([]byte(str))
	return hex.EncodeToString(hashS.Sum(nil))
}

func (instance *XPStringImpl) Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (instance *XPStringImpl) Base64Decode(str string) (result string, err error) {
	r, e := base64.StdEncoding.DecodeString(str)
	if e != nil {
		return "", e
	} else {
		return string(r), nil
	}
}

func (instance *XPStringImpl) UrlEncode(str string) string {
	return url.QueryEscape(str)
}

func (instance *XPStringImpl) UrlDecode(str string) (result string, err error) {
	r, e := url.QueryUnescape(str)
	if e != nil {
		return "", e
	} else {
		return string(r), nil
	}
}

func (instance *XPStringImpl) Uppercase(str string) (result string)  {
	return strings.ToUpper(str)
}

func (instance *XPStringImpl) Lowercase(str string) (result string)  {
	return strings.ToLower(str)
}

func (instance *XPStringImpl) Join(s ...string) string {
	var b strings.Builder
	l:=len(s)
	for i:=0;i<l;i++{
		b.WriteString(s[i])
	}
	return b.String()
}

func (instance *XPStringImpl) JoinWithCap(s []string, cap int) string  {
	var b strings.Builder
	l:=len(s)
	b.Grow(cap)
	for i:=0;i<l;i++{
		b.WriteString(s[i])
	}
	return b.String()
}

func (instance *XPStringImpl) Split(str, sep string) []string {
	return strings.Split(str, sep)
}

func (instance *XPStringImpl) IndexOfSubString(str string, substr string) int {
	// 子串在字符串的字节位置
	result := strings.Index(str,substr)

	if result >= 0 {
		// 获得子串之前的字符串并转换成[]byte
		prefix := []byte(str)[0:result]
		// 将子串之前的字符串转换成[]rune
		rs := []rune(string(prefix))
		// 获得子串之前的字符串的长度，便是子串在字符串的字符位置
		result = len(rs)
	}

	return result
}

func (instance *XPStringImpl) SubString(str string, start, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs  := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if start < 0 {
		start = 0
	}

	if start >= lth {
		start = lth
	}

	end := start + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[start:end])
}

func (instance *XPStringImpl) Format(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a ...)
}

func (instance *XPStringImpl) Replace(str, old, new string) string {
	return strings.ReplaceAll(str, old, new)
}

func (instance *XPStringImpl) Trim(str string) string {
	return strings.Trim(str, " ")
}

func (instance *XPStringImpl) TrimLeft(str string) string {
	return strings.TrimLeft(str, " ")
}

func (instance *XPStringImpl) TrimRight(str string) string {
	return strings.TrimRight(str, " ")
}

func (instance *XPStringImpl) Equal(source, target string) bool {
	return strings.EqualFold(source, target)
}

func (instance *XPStringImpl) EqualIgnoreCase(source, target string) bool {
	return strings.EqualFold(strings.ToLower(source), strings.ToLower(target))
}

func (instance *XPStringImpl) StartWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

func (instance *XPStringImpl) EndWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

func (instance *XPStringImpl) Contain(str, sub string) bool {
	return strings.Contains(str, sub)
}

func (instance *XPStringImpl) ContainHan(str string) bool {
	result := false
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			result = true
			break
		}
	}
	return result
}

func (instance *XPStringImpl) IsEmpty(str string) bool {
	return strings.Trim(str, " ") == ""
}

func (instance *XPStringImpl) IsNumeric(str string) bool {
	pattern := "\\d+"
	if r, err := getRegexp(pattern); err == nil {
		return r.Match([]byte(str))
	}
	return false
}

func (instance *XPStringImpl) Random(length int) string {
	if length == 0 {
		return ""
	}

	var seed = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

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

func (instance *XPStringImpl) RandomOnlyNumber(length int) string {
	if length == 0 {
		return ""
	}

	var seed = []byte("0123456789")

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

func (instance *XPStringImpl) RandomWithSeed(length int, seed string) string {
	if length == 0 {
		return ""
	}

	var seedByte = []byte(seed)

	seedLen := len(seedByte)
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
			b[i] = seedByte[c%seedLen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}

/**
 * 字符串掩码处理
 * @param str string 需要处理的字符串
 * @param code string 掩码符号如 （*）
 * @param start int 开始处理位置 (负数表示从尾部开始)
 * @param length int 掩饰长度
 * @return string 处理过后的字符串
 *
 * 示例  str = 18888888888  code = *  start = -4 length = 4  返回  1888888****
 */
func (instance *XPStringImpl) HideStr(str, code string, start, length int) string {
	l := len(str)
	if start < 0 {
		start = -start
		start = l - start
	}
	var end string
	if length < l-start {
		end = str[start+length : l]
	}
	if l-start < length {
		length = l - start
	}
	hide := strings.Repeat(code, length)
	return str[:start] + hide + end
}

// 分词
// 将字符串分隔为一个单词列表，支持数字、驼峰风格等
func (instance *XPStringImpl) Segment(str string) (entries []string) {
	if !utf8.ValidString(str) {
		return []string{str}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0

	for _, r := range str {
		switch true {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case unicode.IsDigit(r):
			class = 3
		default:
			class = 4
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}

	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}

	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}

	return
}

func (instance *XPStringImpl) ToInt(str string, def int) int {
	out, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return def
	}
	return int(out)
}

func (instance *XPStringImpl) ToInt32(str string, def int32) int32 {
	out, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return def
	}
	return int32(out)
}

func (instance *XPStringImpl) ToInt64(str string, def int64) int64 {
	out, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return def
	}

	return out
}

func (instance *XPStringImpl) ToFloat32(str string, def float32) float32 {
	out, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return def
	}

	return float32(out)
}

func (instance *XPStringImpl) ToFloat64(str string, def float64) float64 {
	out, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return def
	}

	return out
}

func (instance *XPStringImpl) ToBool(str string, def bool) bool {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "YES", "Yes":
		return true
	case "0", "f", "F", "false", "FALSE", "False", "NO", "No":
		return false
	}

	return def
}
