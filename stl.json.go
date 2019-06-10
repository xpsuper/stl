package stl

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)


type JsonItemType int

const (
	Null JsonItemType = iota
	False
	Number
	String
	True
	JSON
)

func (t JsonItemType) String() string {
	switch t {
	default:
		return ""
	case Null:
		return "Null"
	case False:
		return "False"
	case Number:
		return "Number"
	case String:
		return "String"
	case True:
		return "True"
	case JSON:
		return "JSON"
	}
}

type JsonItem struct {
	JsonItemType JsonItemType
	Raw string
	Str string
	Num float64
	Index int
}

func (t JsonItem) String() string {
	switch t.JsonItemType {
	default:
		return ""
	case False:
		return "false"
	case Number:
		if len(t.Raw) == 0 {
			return strconv.FormatFloat(t.Num, 'f', -1, 64)
		}
		var i int
		if t.Raw[0] == '-' {
			i++
		}
		for ; i < len(t.Raw); i++ {
			if t.Raw[i] < '0' || t.Raw[i] > '9' {
				return strconv.FormatFloat(t.Num, 'f', -1, 64)
			}
		}
		return t.Raw
	case String:
		return t.Str
	case JSON:
		return t.Raw
	case True:
		return "true"
	}
}

func (t JsonItem) Bool() bool {
	switch t.JsonItemType {
	default:
		return false
	case True:
		return true
	case String:
		return t.Str != "" && t.Str != "0" && t.Str != "false" && t.Str != "False" && t.Str != "FALSE"
	case Number:
		return t.Num != 0
	}
}

func (t JsonItem) Int() int64 {
	switch t.JsonItemType {
	default:
		return 0
	case True:
		return 1
	case String:
		n, _ := parseInt(t.Str)
		return n
	case Number:
		n, ok := floatToInt(t.Num)
		if !ok {
			n, ok = parseInt(t.Raw)
			if !ok {
				return int64(t.Num)
			}
		}
		return n
	}
}

func (t JsonItem) Uint() uint64 {
	switch t.JsonItemType {
	default:
		return 0
	case True:
		return 1
	case String:
		n, _ := parseUint(t.Str)
		return n
	case Number:
		n, ok := floatToUint(t.Num)
		if !ok {
			n, ok = parseUint(t.Raw)
			if !ok {
				return uint64(t.Num)
			}
		}
		return n
	}
}

func (t JsonItem) Float() float64 {
	switch t.JsonItemType {
	default:
		return 0
	case True:
		return 1
	case String:
		n, _ := strconv.ParseFloat(t.Str, 64)
		return n
	case Number:
		return t.Num
	}
}

func (t JsonItem) Time() time.Time {
	res, _ := time.Parse(time.RFC3339, t.String())
	return res
}

func (t JsonItem) Array() []JsonItem {
	if t.JsonItemType == Null {
		return []JsonItem{}
	}
	if t.JsonItemType != JSON {
		return []JsonItem{t}
	}
	r := t.arrayOrMap('[', false)
	return r.a
}

func (t JsonItem) IsObject() bool {
	return t.JsonItemType == JSON && len(t.Raw) > 0 && t.Raw[0] == '{'
}

func (t JsonItem) IsArray() bool {
	return t.JsonItemType == JSON && len(t.Raw) > 0 && t.Raw[0] == '['
}

func (t JsonItem) ForEach(iterator func(key, value JsonItem) bool) {
	if !t.Exists() {
		return
	}
	if t.JsonItemType != JSON {
		iterator(JsonItem{}, t)
		return
	}
	json := t.Raw
	var keys bool
	var i int
	var key, value JsonItem
	for ; i < len(json); i++ {
		if json[i] == '{' {
			i++
			key.JsonItemType = String
			keys = true
			break
		} else if json[i] == '[' {
			i++
			break
		}
		if json[i] > ' ' {
			return
		}
	}
	var str string
	var vesc bool
	var ok bool
	for ; i < len(json); i++ {
		if keys {
			if json[i] != '"' {
				continue
			}
			s := i
			i, str, vesc, ok = parseString(json, i+1)
			if !ok {
				return
			}
			if vesc {
				key.Str = unescape(str[1 : len(str)-1])
			} else {
				key.Str = str[1 : len(str)-1]
			}
			key.Raw = str
			key.Index = s
		}
		for ; i < len(json); i++ {
			if json[i] <= ' ' || json[i] == ',' || json[i] == ':' {
				continue
			}
			break
		}
		s := i
		i, value, ok = parseAny(json, i, true)
		if !ok {
			return
		}
		value.Index = s
		if !iterator(key, value) {
			return
		}
	}
}

func (t JsonItem) Map() map[string]JsonItem {
	if t.JsonItemType != JSON {
		return map[string]JsonItem{}
	}
	r := t.arrayOrMap('{', false)
	return r.o
}

func (t JsonItem) Value() interface{} {
	if t.JsonItemType == String {
		return t.Str
	}
	switch t.JsonItemType {
	default:
		return nil
	case False:
		return false
	case Number:
		return t.Num
	case JSON:
		r := t.arrayOrMap(0, true)
		if r.vc == '{' {
			return r.oi
		} else if r.vc == '[' {
			return r.ai
		}
		return nil
	case True:
		return true
	}
}

func (t JsonItem) Get(path string) JsonItem {
	return Get(t.Raw, path)
}

// Less return true if a token is less than another token.
// The caseSensitive paramater is used when the tokens are Strings.
// The order when comparing two different JsonItemType is:
//
//  Null < False < Number < String < True < JSON
//
func (t JsonItem) Less(token JsonItem, caseSensitive bool) bool {
	if t.JsonItemType < token.JsonItemType {
		return true
	}
	if t.JsonItemType > token.JsonItemType {
		return false
	}
	if t.JsonItemType == String {
		if caseSensitive {
			return t.Str < token.Str
		}
		return stringLessInsensitive(t.Str, token.Str)
	}
	if t.JsonItemType == Number {
		return t.Num < token.Num
	}
	return t.Raw < token.Raw
}

func (t JsonItem) Exists() bool {
	return t.JsonItemType != Null || len(t.Raw) != 0
}

// Valid returns true if the input is valid json.
//
//  if !gjson.Valid(json) {
//  	return errors.New("invalid json")
//  }
//  value := gjson.Get(json, "name.last")
//
func Valid(json string) bool {
	_, ok := validpayload(stringBytes(json), 0)
	return ok
}

// ValidBytes returns true if the input is valid json.
//
//  if !gjson.Valid(json) {
//  	return errors.New("invalid json")
//  }
//  value := gjson.Get(json, "name.last")
//
// If working with bytes, this method preferred over ValidBytes(string(data))
//
func ValidBytes(json []byte) bool {
	_, ok := validpayload(json, 0)
	return ok
}

func Parse(json string) JsonItem {
	var value JsonItem
	for i := 0; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			value.JsonItemType = JSON
			value.Raw = json[i:] // just take the entire raw
			break
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		default:
			if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
				value.JsonItemType = Number
				value.Raw, value.Num = tonum(json[i:])
			} else {
				return JsonItem{}
			}
		case 'n':
			value.JsonItemType = Null
			value.Raw = tolit(json[i:])
		case 't':
			value.JsonItemType = True
			value.Raw = tolit(json[i:])
		case 'f':
			value.JsonItemType = False
			value.Raw = tolit(json[i:])
		case '"':
			value.JsonItemType = String
			value.Raw, value.Str = tostr(json[i:])
		}
		break
	}
	return value
}

func ParseBytes(json []byte) JsonItem {
	return Parse(string(json))
}

//路径是由点连接的一系列键。
//键可以包含特殊的通配符“*”和“?”。
//要访问数组值，请使用索引作为键。
//要获取数组中的元素数量或访问子路径，请使用'#'字符。
//点和通配符可以用'\'转义。
//
//  {
//    "name": {"first": "Tom", "last": "Anderson"},
//    "age":37,
//    "children": ["Sara","Alex","Jack"],
//    "friends": [
//      {"first": "James", "last": "Murphy"},
//      {"first": "Roger", "last": "Craig"}
//    ]
//  }
//  "name.last"          >> "Anderson"
//  "age"                >> 37
//  "children"           >> ["Sara","Alex","Jack"]
//  "children.#"         >> 3
//  "children.1"         >> "Alex"
//  "child*.2"           >> "Jack"
//  "c?ildren.0"         >> "Sara"
//  "friends.#.first"    >> ["James","Roger"]
func Get(json, path string) JsonItem {
	if !DisableModifiers {
		if len(path) > 1 && path[0] == '@' {
			// possible modifier
			var ok bool
			var rjson string
			path, rjson, ok = execModifier(json, path)
			if ok {
				if len(path) > 0 && path[0] == '|' {
					res := Get(rjson, path[1:])
					res.Index = 0
					return res
				}
				return Parse(rjson)
			}
		}
	}
	var i int
	var c = &parseContext{json: json}
	if len(path) >= 2 && path[0] == '.' && path[1] == '.' {
		c.lines = true
		parseArray(c, 0, path[2:])
	} else {
		for ; i < len(c.json); i++ {
			if c.json[i] == '{' {
				i++
				parseObject(c, i, path)
				break
			}
			if c.json[i] == '[' {
				i++
				parseArray(c, i, path)
				break
			}
		}
	}
	if c.piped {
		res := c.value.Get(c.pipe)
		res.Index = 0
		return res
	}
	fillIndex(json, c)
	return c.value
}

// GetBytes searches json for the specified path.
// If working with bytes, this method preferred over Get(string(data), path)
func GetBytes(json []byte, path string) JsonItem {
	return getBytes(json, path)
}

// GetMany searches json for the multiple paths.
// The return value is a JsonItem array where the number of JsonItems
// will be equal to the number of input paths.
func GetMany(json string, path ...string) []JsonItem {
	res := make([]JsonItem, len(path))
	for i, path := range path {
		res[i] = Get(json, path)
	}
	return res
}

// GetManyBytes searches json for the multiple paths.
// The return value is a JsonItem array where the number of JsonItems
// will be equal to the number of input paths.
func GetManyBytes(json []byte, path ...string) []JsonItem {
	res := make([]JsonItem, len(path))
	for i, path := range path {
		res[i] = GetBytes(json, path)
	}
	return res
}

func ForEachLine(json string, iterator func(line JsonItem) bool) {
	var res JsonItem
	var i int
	for {
		i, res, _ = parseAny(json, i, true)
		if !res.Exists() {
			break
		}
		if !iterator(res) {
			return
		}
	}
}

//=========================================

type arrayOrMapJsonItem struct {
	a  []JsonItem
	ai []interface{}
	o  map[string]JsonItem
	oi map[string]interface{}
	vc byte
}

func (t JsonItem) arrayOrMap(vc byte, valueize bool) (r arrayOrMapJsonItem) {
	var json = t.Raw
	var i int
	var value JsonItem
	var count int
	var key JsonItem
	if vc == 0 {
		for ; i < len(json); i++ {
			if json[i] == '{' || json[i] == '[' {
				r.vc = json[i]
				i++
				break
			}
			if json[i] > ' ' {
				goto end
			}
		}
	} else {
		for ; i < len(json); i++ {
			if json[i] == vc {
				i++
				break
			}
			if json[i] > ' ' {
				goto end
			}
		}
		r.vc = vc
	}
	if r.vc == '{' {
		if valueize {
			r.oi = make(map[string]interface{})
		} else {
			r.o = make(map[string]JsonItem)
		}
	} else {
		if valueize {
			r.ai = make([]interface{}, 0)
		} else {
			r.a = make([]JsonItem, 0)
		}
	}
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		// get next value
		if json[i] == ']' || json[i] == '}' {
			break
		}
		switch json[i] {
		default:
			if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
				value.JsonItemType = Number
				value.Raw, value.Num = tonum(json[i:])
				value.Str = ""
			} else {
				continue
			}
		case '{', '[':
			value.JsonItemType = JSON
			value.Raw = squash(json[i:])
			value.Str, value.Num = "", 0
		case 'n':
			value.JsonItemType = Null
			value.Raw = tolit(json[i:])
			value.Str, value.Num = "", 0
		case 't':
			value.JsonItemType = True
			value.Raw = tolit(json[i:])
			value.Str, value.Num = "", 0
		case 'f':
			value.JsonItemType = False
			value.Raw = tolit(json[i:])
			value.Str, value.Num = "", 0
		case '"':
			value.JsonItemType = String
			value.Raw, value.Str = tostr(json[i:])
			value.Num = 0
		}
		i += len(value.Raw) - 1

		if r.vc == '{' {
			if count%2 == 0 {
				key = value
			} else {
				if valueize {
					if _, ok := r.oi[key.Str]; !ok {
						r.oi[key.Str] = value.Value()
					}
				} else {
					if _, ok := r.o[key.Str]; !ok {
						r.o[key.Str] = value
					}
				}
			}
			count++
		} else {
			if valueize {
				r.ai = append(r.ai, value.Value())
			} else {
				r.a = append(r.a, value)
			}
		}
	}
end:
	return
}

func squash(json string) string {
	depth := 1
	for i := 1; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					return json[:i+1]
				}
			}
		}
	}
	return json
}

func tonum(json string) (raw string, num float64) {
	for i := 1; i < len(json); i++ {
		// less than dash might have valid characters
		if json[i] <= '-' {
			if json[i] <= ' ' || json[i] == ',' {
				// break on whitespace and comma
				raw = json[:i]
				num, _ = strconv.ParseFloat(raw, 64)
				return
			}
			// could be a '+' or '-'. let's assume so.
			continue
		}
		if json[i] < ']' {
			// probably a valid number
			continue
		}
		if json[i] == 'e' || json[i] == 'E' {
			// allow for exponential numbers
			continue
		}
		// likely a ']' or '}'
		raw = json[:i]
		num, _ = strconv.ParseFloat(raw, 64)
		return
	}
	raw = json
	num, _ = strconv.ParseFloat(raw, 64)
	return
}

func tolit(json string) (raw string) {
	for i := 1; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return json[:i]
		}
	}
	return json
}

func tostr(json string) (raw string, str string) {
	// expects that the lead character is a '"'
	for i := 1; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return json[:i+1], json[1:i]
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					break
				}
			}
			var ret string
			if i+1 < len(json) {
				ret = json[:i+1]
			} else {
				ret = json[:i]
			}
			return ret, unescape(json[1:i])
		}
	}
	return json, json[1:]
}

func parseString(json string, i int) (int, string, bool, bool) {
	var s = i
	for ; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return i + 1, json[s-1 : i+1], false, true
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return i + 1, json[s-1 : i+1], true, true
				}
			}
			break
		}
	}
	return i, json[s-1:], false, false
}

func parseNumber(json string, i int) (int, string) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' || json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

func parseLiteral(json string, i int) (int, string) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

type arrayPathJsonItem struct {
	part    string
	path    string
	pipe    string
	piped   bool
	more    bool
	alogok  bool
	arrch   bool
	alogkey string
	query   struct {
		on    bool
		path  string
		op    string
		value string
		all   bool
	}
}

func parseArrayPath(path string) (r arrayPathJsonItem) {
	for i := 0; i < len(path); i++ {
		if !DisableChaining {
			if path[i] == '|' {
				r.part = path[:i]
				r.pipe = path[i+1:]
				r.piped = true
				return
			}
		}
		if path[i] == '.' {
			r.part = path[:i]
			r.path = path[i+1:]
			r.more = true
			return
		}
		if path[i] == '#' {
			r.arrch = true
			if i == 0 && len(path) > 1 {
				if path[1] == '.' {
					r.alogok = true
					r.alogkey = path[2:]
					r.path = path[:1]
				} else if path[1] == '[' {
					r.query.on = true
					// query
					i += 2
					// whitespace
					for ; i < len(path); i++ {
						if path[i] > ' ' {
							break
						}
					}
					s := i
					for ; i < len(path); i++ {
						if path[i] <= ' ' ||
							path[i] == '!' ||
							path[i] == '=' ||
							path[i] == '<' ||
							path[i] == '>' ||
							path[i] == '%' ||
							path[i] == ']' {
							break
						}
					}
					r.query.path = path[s:i]
					// whitespace
					for ; i < len(path); i++ {
						if path[i] > ' ' {
							break
						}
					}
					if i < len(path) {
						s = i
						if path[i] == '!' {
							if i < len(path)-1 && (path[i+1] == '=' || path[i+1] == '%') {
								i++
							}
						} else if path[i] == '<' || path[i] == '>' {
							if i < len(path)-1 && path[i+1] == '=' {
								i++
							}
						} else if path[i] == '=' {
							if i < len(path)-1 && path[i+1] == '=' {
								s++
								i++
							}
						}
						i++
						r.query.op = path[s:i]
						// whitespace
						for ; i < len(path); i++ {
							if path[i] > ' ' {
								break
							}
						}
						s = i
						for ; i < len(path); i++ {
							if path[i] == '"' {
								i++
								s2 := i
								for ; i < len(path); i++ {
									if path[i] > '\\' {
										continue
									}
									if path[i] == '"' {
										// look for an escaped slash
										if path[i-1] == '\\' {
											n := 0
											for j := i - 2; j > s2-1; j-- {
												if path[j] != '\\' {
													break
												}
												n++
											}
											if n%2 == 0 {
												continue
											}
										}
										break
									}
								}
							} else if path[i] == ']' {
								if i+1 < len(path) && path[i+1] == '#' {
									r.query.all = true
								}
								break
							}
						}
						if i > len(path) {
							i = len(path)
						}
						v := path[s:i]
						for len(v) > 0 && v[len(v)-1] <= ' ' {
							v = v[:len(v)-1]
						}
						r.query.value = v
					}
				}
			}
			continue
		}
	}
	r.part = path
	r.path = ""
	return
}

var DisableChaining = false

type objectPathJsonItem struct {
	part  string
	path  string
	pipe  string
	piped bool
	wild  bool
	more  bool
}

func parseObjectPath(path string) (r objectPathJsonItem) {
	for i := 0; i < len(path); i++ {
		if !DisableChaining {
			if path[i] == '|' {
				r.part = path[:i]
				r.pipe = path[i+1:]
				r.piped = true
				return
			}
		}
		if path[i] == '.' {
			r.part = path[:i]
			r.path = path[i+1:]
			r.more = true
			return
		}
		if path[i] == '*' || path[i] == '?' {
			r.wild = true
			continue
		}
		if path[i] == '\\' {
			// go into escape mode. this is a slower path that
			// strips off the escape character from the part.
			epart := []byte(path[:i])
			i++
			if i < len(path) {
				epart = append(epart, path[i])
				i++
				for ; i < len(path); i++ {
					if path[i] == '\\' {
						i++
						if i < len(path) {
							epart = append(epart, path[i])
						}
						continue
					} else if path[i] == '.' {
						r.part = string(epart)
						r.path = path[i+1:]
						r.more = true
						return
					} else if path[i] == '*' || path[i] == '?' {
						r.wild = true
					}
					epart = append(epart, path[i])
				}
			}
			// append the last part
			r.part = string(epart)
			return
		}
	}
	r.part = path
	return
}

func parseSquash(json string, i int) (int, string) {
	s := i
	i++
	depth := 1
	for ; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[':
				depth++
			case '}', ']':
				depth--
				if depth == 0 {
					i++
					return i, json[s:i]
				}
			}
		}
	}
	return i, json[s:]
}

func parseObject(c *parseContext, i int, path string) (int, bool) {
	var pmatch, kesc, vesc, ok, hit bool
	var key, val string
	rp := parseObjectPath(path)
	if !rp.more && rp.piped {
		c.pipe = rp.pipe
		c.piped = true
	}
	for i < len(c.json) {
		for ; i < len(c.json); i++ {
			if c.json[i] == '"' {
				// parse_key_string
				// this is slightly different from getting s string value
				// because we don't need the outer quotes.
				i++
				var s = i
				for ; i < len(c.json); i++ {
					if c.json[i] > '\\' {
						continue
					}
					if c.json[i] == '"' {
						i, key, kesc, ok = i+1, c.json[s:i], false, true
						goto parse_key_string_done
					}
					if c.json[i] == '\\' {
						i++
						for ; i < len(c.json); i++ {
							if c.json[i] > '\\' {
								continue
							}
							if c.json[i] == '"' {
								// look for an escaped slash
								if c.json[i-1] == '\\' {
									n := 0
									for j := i - 2; j > 0; j-- {
										if c.json[j] != '\\' {
											break
										}
										n++
									}
									if n%2 == 0 {
										continue
									}
								}
								i, key, kesc, ok = i+1, c.json[s:i], true, true
								goto parse_key_string_done
							}
						}
						break
					}
				}
				key, kesc, ok = c.json[s:], false, false
			parse_key_string_done:
				break
			}
			if c.json[i] == '}' {
				return i + 1, false
			}
		}
		if !ok {
			return i, false
		}
		if rp.wild {
			if kesc {
				pmatch = Match(unescape(key), rp.part)
			} else {
				pmatch = Match(key, rp.part)
			}
		} else {
			if kesc {
				pmatch = rp.part == unescape(key)
			} else {
				pmatch = rp.part == key
			}
		}
		hit = pmatch && !rp.more
		for ; i < len(c.json); i++ {
			switch c.json[i] {
			default:
				continue
			case '"':
				i++
				i, val, vesc, ok = parseString(c.json, i)
				if !ok {
					return i, false
				}
				if hit {
					if vesc {
						c.value.Str = unescape(val[1 : len(val)-1])
					} else {
						c.value.Str = val[1 : len(val)-1]
					}
					c.value.Raw = val
					c.value.JsonItemType = String
					return i, true
				}
			case '{':
				if pmatch && !hit {
					i, hit = parseObject(c, i+1, rp.path)
					if hit {
						return i, true
					}
				} else {
					i, val = parseSquash(c.json, i)
					if hit {
						c.value.Raw = val
						c.value.JsonItemType = JSON
						return i, true
					}
				}
			case '[':
				if pmatch && !hit {
					i, hit = parseArray(c, i+1, rp.path)
					if hit {
						return i, true
					}
				} else {
					i, val = parseSquash(c.json, i)
					if hit {
						c.value.Raw = val
						c.value.JsonItemType = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if hit {
					c.value.Raw = val
					c.value.JsonItemType = Number
					c.value.Num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if hit {
					c.value.Raw = val
					switch vc {
					case 't':
						c.value.JsonItemType = True
					case 'f':
						c.value.JsonItemType = False
					}
					return i, true
				}
			}
			break
		}
	}
	return i, false
}

func queryMatches(rp *arrayPathJsonItem, value JsonItem) bool {
	rpv := rp.query.value
	if len(rpv) > 2 && rpv[0] == '"' && rpv[len(rpv)-1] == '"' {
		rpv = rpv[1 : len(rpv)-1]
	}
	switch value.JsonItemType {
	case String:
		switch rp.query.op {
		case "=":
			return value.Str == rpv
		case "!=":
			return value.Str != rpv
		case "<":
			return value.Str < rpv
		case "<=":
			return value.Str <= rpv
		case ">":
			return value.Str > rpv
		case ">=":
			return value.Str >= rpv
		case "%":
			return Match(value.Str, rpv)
		case "!%":
			return !Match(value.Str, rpv)
		}
	case Number:
		rpvn, _ := strconv.ParseFloat(rpv, 64)
		switch rp.query.op {
		case "=":
			return value.Num == rpvn
		case "!=":
			return value.Num != rpvn
		case "<":
			return value.Num < rpvn
		case "<=":
			return value.Num <= rpvn
		case ">":
			return value.Num > rpvn
		case ">=":
			return value.Num >= rpvn
		}
	case True:
		switch rp.query.op {
		case "=":
			return rpv == "true"
		case "!=":
			return rpv != "true"
		case ">":
			return rpv == "false"
		case ">=":
			return true
		}
	case False:
		switch rp.query.op {
		case "=":
			return rpv == "false"
		case "!=":
			return rpv != "false"
		case "<":
			return rpv == "true"
		case "<=":
			return true
		}
	}
	return false
}

func parseArray(c *parseContext, i int, path string) (int, bool) {
	var pmatch, vesc, ok, hit bool
	var val string
	var h int
	var alog []int
	var partidx int
	var multires []byte
	rp := parseArrayPath(path)
	if !rp.arrch {
		n, ok := parseUint(rp.part)
		if !ok {
			partidx = -1
		} else {
			partidx = int(n)
		}
	}
	if !rp.more && rp.piped {
		c.pipe = rp.pipe
		c.piped = true
	}
	for i < len(c.json)+1 {
		if !rp.arrch {
			pmatch = partidx == h
			hit = pmatch && !rp.more
		}
		h++
		if rp.alogok {
			alog = append(alog, i)
		}
		for ; ; i++ {
			var ch byte
			if i > len(c.json) {
				break
			} else if i == len(c.json) {
				ch = ']'
			} else {
				ch = c.json[i]
			}
			switch ch {
			default:
				continue
			case '"':
				i++
				i, val, vesc, ok = parseString(c.json, i)
				if !ok {
					return i, false
				}
				if hit {
					if rp.alogok {
						break
					}
					if vesc {
						c.value.Str = unescape(val[1 : len(val)-1])
					} else {
						c.value.Str = val[1 : len(val)-1]
					}
					c.value.Raw = val
					c.value.JsonItemType = String
					return i, true
				}
			case '{':
				if pmatch && !hit {
					i, hit = parseObject(c, i+1, rp.path)
					if hit {
						if rp.alogok {
							break
						}
						return i, true
					}
				} else {
					i, val = parseSquash(c.json, i)
					if rp.query.on {
						res := Get(val, rp.query.path)
						if queryMatches(&rp, res) {
							if rp.more {
								res = Get(val, rp.path)
							} else {
								res = JsonItem{Raw: val, JsonItemType: JSON}
							}
							if rp.query.all {
								if len(multires) == 0 {
									multires = append(multires, '[')
								} else {
									multires = append(multires, ',')
								}
								multires = append(multires, res.Raw...)
							} else {
								c.value = res
								return i, true
							}
						}
					} else if hit {
						if rp.alogok {
							break
						}
						c.value.Raw = val
						c.value.JsonItemType = JSON
						return i, true
					}
				}
			case '[':
				if pmatch && !hit {
					i, hit = parseArray(c, i+1, rp.path)
					if hit {
						if rp.alogok {
							break
						}
						return i, true
					}
				} else {
					i, val = parseSquash(c.json, i)
					if hit {
						if rp.alogok {
							break
						}
						c.value.Raw = val
						c.value.JsonItemType = JSON
						return i, true
					}
				}
			case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				i, val = parseNumber(c.json, i)
				if hit {
					if rp.alogok {
						break
					}
					c.value.Raw = val
					c.value.JsonItemType = Number
					c.value.Num, _ = strconv.ParseFloat(val, 64)
					return i, true
				}
			case 't', 'f', 'n':
				vc := c.json[i]
				i, val = parseLiteral(c.json, i)
				if hit {
					if rp.alogok {
						break
					}
					c.value.Raw = val
					switch vc {
					case 't':
						c.value.JsonItemType = True
					case 'f':
						c.value.JsonItemType = False
					}
					return i, true
				}
			case ']':
				if rp.arrch && rp.part == "#" {
					if rp.alogok {
						var jsons = make([]byte, 0, 64)
						jsons = append(jsons, '[')

						for j, k := 0, 0; j < len(alog); j++ {
							_, res, ok := parseAny(c.json, alog[j], true)
							if ok {
								res := res.Get(rp.alogkey)
								if res.Exists() {
									if k > 0 {
										jsons = append(jsons, ',')
									}
									jsons = append(jsons, []byte(res.Raw)...)
									k++
								}
							}
						}
						jsons = append(jsons, ']')
						c.value.JsonItemType = JSON
						c.value.Raw = string(jsons)
						return i + 1, true
					}
					if rp.alogok {
						break
					}
					c.value.Raw = ""
					c.value.JsonItemType = Number
					c.value.Num = float64(h - 1)
					c.calcd = true
					return i + 1, true
				}
				if len(multires) > 0 && !c.value.Exists() {
					c.value = JsonItem{
						Raw:  string(append(multires, ']')),
						JsonItemType: JSON,
					}
				}
				return i + 1, false
			}
			break
		}
	}
	return i, false
}

type parseContext struct {
	json  string
	value JsonItem
	pipe  string
	piped bool
	calcd bool
	lines bool
}

// runeit returns the rune from the the \uXXXX
func runeit(json string) rune {
	n, _ := strconv.ParseUint(json[:4], 16, 64)
	return rune(n)
}

// unescape unescapes a string
func unescape(json string) string { //, error) {
	var str = make([]byte, 0, len(json))
	for i := 0; i < len(json); i++ {
		switch {
		default:
			str = append(str, json[i])
		case json[i] < ' ':
			return string(str)
		case json[i] == '\\':
			i++
			if i >= len(json) {
				return string(str)
			}
			switch json[i] {
			default:
				return string(str)
			case '\\':
				str = append(str, '\\')
			case '/':
				str = append(str, '/')
			case 'b':
				str = append(str, '\b')
			case 'f':
				str = append(str, '\f')
			case 'n':
				str = append(str, '\n')
			case 'r':
				str = append(str, '\r')
			case 't':
				str = append(str, '\t')
			case '"':
				str = append(str, '"')
			case 'u':
				if i+5 > len(json) {
					return string(str)
				}
				r := runeit(json[i+1:])
				i += 5
				if utf16.IsSurrogate(r) {
					// need another code
					if len(json[i:]) >= 6 && json[i] == '\\' && json[i+1] == 'u' {
						// we expect it to be correct so just consume it
						r = utf16.DecodeRune(r, runeit(json[i+2:]))
						i += 6
					}
				}
				// provide enough space to encode the largest utf8 possible
				str = append(str, 0, 0, 0, 0, 0, 0, 0, 0)
				n := utf8.EncodeRune(str[len(str)-8:], r)
				str = str[:len(str)-8+n]
				i-- // backtrack index by one
			}
		}
	}
	return string(str)
}

func stringLessInsensitive(a, b string) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] >= 'A' && a[i] <= 'Z' {
			if b[i] >= 'A' && b[i] <= 'Z' {
				// both are uppercase, do nothing
				if a[i] < b[i] {
					return true
				} else if a[i] > b[i] {
					return false
				}
			} else {
				// a is uppercase, convert a to lowercase
				if a[i]+32 < b[i] {
					return true
				} else if a[i]+32 > b[i] {
					return false
				}
			}
		} else if b[i] >= 'A' && b[i] <= 'Z' {
			// b is uppercase, convert b to lowercase
			if a[i] < b[i]+32 {
				return true
			} else if a[i] > b[i]+32 {
				return false
			}
		} else {
			// neither are uppercase
			if a[i] < b[i] {
				return true
			} else if a[i] > b[i] {
				return false
			}
		}
	}
	return len(a) < len(b)
}

// parseAny parses the next value from a json string.
// A JsonItem is returned when the hit param is set.
// The return values are (i int, res JsonItem, ok bool)
func parseAny(json string, i int, hit bool) (int, JsonItem, bool) {
	var res JsonItem
	var val string
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			i, val = parseSquash(json, i)
			if hit {
				res.Raw = val
				res.JsonItemType = JSON
			}
			return i, res, true
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		case '"':
			i++
			var vesc bool
			var ok bool
			i, val, vesc, ok = parseString(json, i)
			if !ok {
				return i, res, false
			}
			if hit {
				res.JsonItemType = String
				res.Raw = val
				if vesc {
					res.Str = unescape(val[1 : len(val)-1])
				} else {
					res.Str = val[1 : len(val)-1]
				}
			}
			return i, res, true
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i, val = parseNumber(json, i)
			if hit {
				res.Raw = val
				res.JsonItemType = Number
				res.Num, _ = strconv.ParseFloat(val, 64)
			}
			return i, res, true
		case 't', 'f', 'n':
			vc := json[i]
			i, val = parseLiteral(json, i)
			if hit {
				res.Raw = val
				switch vc {
				case 't':
					res.JsonItemType = True
				case 'f':
					res.JsonItemType = False
				}
				return i, res, true
			}
		}
	}
	return i, res, false
}

var ( // used for testing
	testWatchForFallback bool
	testLastWasFallback  bool
)

var fieldsmu sync.RWMutex
var fields = make(map[string]map[string]int)

func assign(jsval JsonItem, goval reflect.Value) {
	if jsval.JsonItemType == Null {
		return
	}
	switch goval.Kind() {
	default:
	case reflect.Ptr:
		if !goval.IsNil() {
			newval := reflect.New(goval.Elem().Type())
			assign(jsval, newval.Elem())
			goval.Elem().Set(newval.Elem())
		} else {
			newval := reflect.New(goval.Type().Elem())
			assign(jsval, newval.Elem())
			goval.Set(newval)
		}
	case reflect.Struct:
		fieldsmu.RLock()
		sf := fields[goval.Type().String()]
		fieldsmu.RUnlock()
		if sf == nil {
			fieldsmu.Lock()
			sf = make(map[string]int)
			for i := 0; i < goval.Type().NumField(); i++ {
				f := goval.Type().Field(i)
				tag := strings.Split(f.Tag.Get("json"), ",")[0]
				if tag != "-" {
					if tag != "" {
						sf[tag] = i
						sf[f.Name] = i
					} else {
						sf[f.Name] = i
					}
				}
			}
			fields[goval.Type().String()] = sf
			fieldsmu.Unlock()
		}
		jsval.ForEach(func(key, value JsonItem) bool {
			if idx, ok := sf[key.Str]; ok {
				f := goval.Field(idx)
				if f.CanSet() {
					assign(value, f)
				}
			}
			return true
		})
	case reflect.Slice:
		if goval.Type().Elem().Kind() == reflect.Uint8 && jsval.JsonItemType == String {
			data, _ := base64.StdEncoding.DecodeString(jsval.String())
			goval.Set(reflect.ValueOf(data))
		} else {
			jsvals := jsval.Array()
			slice := reflect.MakeSlice(goval.Type(), len(jsvals), len(jsvals))
			for i := 0; i < len(jsvals); i++ {
				assign(jsvals[i], slice.Index(i))
			}
			goval.Set(slice)
		}
	case reflect.Array:
		i, n := 0, goval.Len()
		jsval.ForEach(func(_, value JsonItem) bool {
			if i == n {
				return false
			}
			assign(value, goval.Index(i))
			i++
			return true
		})
	case reflect.Map:
		if goval.Type().Key().Kind() == reflect.String && goval.Type().Elem().Kind() == reflect.Interface {
			goval.Set(reflect.ValueOf(jsval.Value()))
		}
	case reflect.Interface:
		goval.Set(reflect.ValueOf(jsval.Value()))
	case reflect.Bool:
		goval.SetBool(jsval.Bool())
	case reflect.Float32, reflect.Float64:
		goval.SetFloat(jsval.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		goval.SetInt(jsval.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		goval.SetUint(jsval.Uint())
	case reflect.String:
		goval.SetString(jsval.String())
	}
	if len(goval.Type().PkgPath()) > 0 {
		v := goval.Addr()
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				u.UnmarshalJSON([]byte(jsval.Raw))
			}
		}
	}
}

var validate uintptr = 1

// UnmarshalValidationEnabled provides the option to disable JSON validation
// during the Unmarshal routine. Validation is enabled by default.
//
// Deprecated: Use encoder/json.Unmarshal instead
func UnmarshalValidationEnabled(enabled bool) {
	if enabled {
		atomic.StoreUintptr(&validate, 1)
	} else {
		atomic.StoreUintptr(&validate, 0)
	}
}

// Unmarshal loads the JSON data into the value pointed to by v.
//
// This function works almost identically to json.Unmarshal except  that
// gjson.Unmarshal will automatically attempt to convert JSON values to any Go
// JsonItemType. For example, the JSON string "100" or the JSON number 100 can be equally
// assigned to Go string, int, byte, uint64, etc. This rule applies to all JsonItemTypes.
//
// Deprecated: Use encoder/json.Unmarshal instead
func Unmarshal(data []byte, v interface{}) error {
	if atomic.LoadUintptr(&validate) == 1 {
		_, ok := validpayload(data, 0)
		if !ok {
			return errors.New("invalid json")
		}
	}
	if v := reflect.ValueOf(v); v.Kind() == reflect.Ptr {
		assign(ParseBytes(data), v)
	}
	return nil
}

func validpayload(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			i, ok = validany(data, i)
			if !ok {
				return i, false
			}
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false
				case ' ', '\t', '\n', '\r':
					continue
				}
			}
			return i, true
		case ' ', '\t', '\n', '\r':
			continue
		}
	}
	return i, false
}
func validany(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			return validobject(data, i+1)
		case '[':
			return validarray(data, i+1)
		case '"':
			return validstring(data, i+1)
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return validnumber(data, i+1)
		case 't':
			return validtrue(data, i+1)
		case 'f':
			return validfalse(data, i+1)
		case 'n':
			return validnull(data, i+1)
		}
	}
	return i, false
}
func validobject(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case '}':
			return i + 1, true
		case '"':
		key:
			if i, ok = validstring(data, i+1); !ok {
				return i, false
			}
			if i, ok = validcolon(data, i); !ok {
				return i, false
			}
			if i, ok = validany(data, i); !ok {
				return i, false
			}
			if i, ok = validcomma(data, i, '}'); !ok {
				return i, false
			}
			if data[i] == '}' {
				return i + 1, true
			}
			i++
			for ; i < len(data); i++ {
				switch data[i] {
				default:
					return i, false
				case ' ', '\t', '\n', '\r':
					continue
				case '"':
					goto key
				}
			}
			return i, false
		}
	}
	return i, false
}
func validcolon(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ':':
			return i + 1, true
		}
	}
	return i, false
}
func validcomma(data []byte, i int, end byte) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			return i, false
		case ' ', '\t', '\n', '\r':
			continue
		case ',':
			return i, true
		case end:
			return i, true
		}
	}
	return i, false
}
func validarray(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		switch data[i] {
		default:
			for ; i < len(data); i++ {
				if i, ok = validany(data, i); !ok {
					return i, false
				}
				if i, ok = validcomma(data, i, ']'); !ok {
					return i, false
				}
				if data[i] == ']' {
					return i + 1, true
				}
			}
		case ' ', '\t', '\n', '\r':
			continue
		case ']':
			return i + 1, true
		}
	}
	return i, false
}
func validstring(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		if data[i] < ' ' {
			return i, false
		} else if data[i] == '\\' {
			i++
			if i == len(data) {
				return i, false
			}
			switch data[i] {
			default:
				return i, false
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for j := 0; j < 4; j++ {
					i++
					if i >= len(data) {
						return i, false
					}
					if !((data[i] >= '0' && data[i] <= '9') ||
						(data[i] >= 'a' && data[i] <= 'f') ||
						(data[i] >= 'A' && data[i] <= 'F')) {
						return i, false
					}
				}
			}
		} else if data[i] == '"' {
			return i + 1, true
		}
	}
	return i, false
}
func validnumber(data []byte, i int) (outi int, ok bool) {
	i--
	// sign
	if data[i] == '-' {
		i++
	}
	// int
	if i == len(data) {
		return i, false
	}
	if data[i] == '0' {
		i++
	} else {
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// frac
	if i == len(data) {
		return i, true
	}
	if data[i] == '.' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// exp
	if i == len(data) {
		return i, true
	}
	if data[i] == 'e' || data[i] == 'E' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] == '+' || data[i] == '-' {
			i++
		}
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	return i, true
}

func validtrue(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'r' && data[i+1] == 'u' && data[i+2] == 'e' {
		return i + 3, true
	}
	return i, false
}
func validfalse(data []byte, i int) (outi int, ok bool) {
	if i+4 <= len(data) && data[i] == 'a' && data[i+1] == 'l' && data[i+2] == 's' && data[i+3] == 'e' {
		return i + 4, true
	}
	return i, false
}
func validnull(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'u' && data[i+1] == 'l' && data[i+2] == 'l' {
		return i + 3, true
	}
	return i, false
}

func parseUint(s string) (n uint64, ok bool) {
	var i int
	if i == len(s) {
		return 0, false
	}
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + uint64(s[i]-'0')
		} else {
			return 0, false
		}
	}
	return n, true
}

func parseInt(s string) (n int64, ok bool) {
	var i int
	var sign bool
	if len(s) > 0 && s[0] == '-' {
		sign = true
		i++
	}
	if i == len(s) {
		return 0, false
	}
	for ; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int64(s[i]-'0')
		} else {
			return 0, false
		}
	}
	if sign {
		return n * -1, true
	}
	return n, true
}

const minUint53 = 0
const maxUint53 = 4503599627370495
const minInt53 = -2251799813685248
const maxInt53 = 2251799813685247

func floatToUint(f float64) (n uint64, ok bool) {
	n = uint64(f)
	if float64(n) == f && n >= minUint53 && n <= maxUint53 {
		return n, true
	}
	return 0, false
}

func floatToInt(f float64) (n int64, ok bool) {
	n = int64(f)
	if float64(n) == f && n >= minInt53 && n <= maxInt53 {
		return n, true
	}
	return 0, false
}

// execModifier parses the path to find a matching modifier function.
// then input expects that the path already starts with a '@'
func execModifier(json, path string) (pathOut, res string, ok bool) {
	name := path[1:]
	var hasArgs bool
	for i := 1; i < len(path); i++ {
		if path[i] == ':' {
			pathOut = path[i+1:]
			name = path[1:i]
			hasArgs = len(pathOut) > 0
			break
		}
		if !DisableChaining {
			if path[i] == '|' {
				pathOut = path[i:]
				name = path[1:i]
				break
			}
		}
	}
	if fn, ok := modifiers[name]; ok {
		var args string
		if hasArgs {
			var parsedArgs bool
			switch pathOut[0] {
			case '{', '[', '"':
				res := Parse(pathOut)
				if res.Exists() {
					_, args = parseSquash(pathOut, 0)
					pathOut = pathOut[len(args):]
					parsedArgs = true
				}
			}
			if !parsedArgs {
				idx := -1
				if !DisableChaining {
					idx = strings.IndexByte(pathOut, '|')
				}
				if idx == -1 {
					args = pathOut
					pathOut = ""
				} else {
					args = pathOut[:idx]
					pathOut = pathOut[idx:]
				}
			}
		}
		return pathOut, fn(json, args), true
	}
	return pathOut, res, false
}

// DisableModifiers will disable the modifier syntax
var DisableModifiers = false

var modifiers = map[string]func(json, arg string) string{
	"pretty":  modPretty,
	"ugly":    modUgly,
	"reverse": modReverse,
}

// AddModifier binds a custom modifier command to the GJSON syntax.
// This operation is not thread safe and should be executed prior to
// using all other gjson function.
func AddModifier(name string, fn func(json, arg string) string) {
	modifiers[name] = fn
}

// ModifierExists returns true when the specified modifier exists.
func ModifierExists(name string, fn func(json, arg string) string) bool {
	_, ok := modifiers[name]
	return ok
}

// @pretty modifier makes the json look nice.
func modPretty(json, arg string) string {
	if len(arg) > 0 {
		opts := *DefaultOptions
		Parse(arg).ForEach(func(key, value JsonItem) bool {
			switch key.String() {
			case "sortKeys":
				opts.SortKeys = value.Bool()
			case "indent":
				opts.Indent = value.String()
			case "prefix":
				opts.Prefix = value.String()
			case "width":
				opts.Width = int(value.Int())
			}
			return true
		})
		return bytesString(PrettyOptions(stringBytes(json), &opts))
	}
	return bytesString(Pretty(stringBytes(json)))
}

// @ugly modifier removes all whitespace.
func modUgly(json, arg string) string {
	return bytesString(Ugly(stringBytes(json)))
}

// @reverse reverses array elements or root object members.
func modReverse(json, arg string) string {
	res := Parse(json)
	if res.IsArray() {
		var values []JsonItem
		res.ForEach(func(_, value JsonItem) bool {
			values = append(values, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '[')
		for i, j := len(values)-1, 0; i >= 0; i, j = i-1, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, values[i].Raw...)
		}
		out = append(out, ']')
		return bytesString(out)
	}
	if res.IsObject() {
		var keyValues []JsonItem
		res.ForEach(func(key, value JsonItem) bool {
			keyValues = append(keyValues, key, value)
			return true
		})
		out := make([]byte, 0, len(json))
		out = append(out, '{')
		for i, j := len(keyValues)-2, 0; i >= 0; i, j = i-2, j+1 {
			if j > 0 {
				out = append(out, ',')
			}
			out = append(out, keyValues[i+0].Raw...)
			out = append(out, ':')
			out = append(out, keyValues[i+1].Raw...)
		}
		out = append(out, '}')
		return bytesString(out)
	}
	return json
}

/*********** Match ***********/
func Match(str, pattern string) bool {
	if pattern == "*" {
		return true
	}
	return deepMatch(str, pattern)
}
func deepMatch(str, pattern string) bool {
	for len(pattern) > 0 {
		if pattern[0] > 0x7f {
			return deepMatchRune(str, pattern)
		}
		switch pattern[0] {
		default:
			if len(str) == 0 {
				return false
			}
			if str[0] > 0x7f {
				return deepMatchRune(str, pattern)
			}
			if str[0] != pattern[0] {
				return false
			}
		case '?':
			if len(str) == 0 {
				return false
			}
		case '*':
			return deepMatch(str, pattern[1:]) ||
				(len(str) > 0 && deepMatch(str[1:], pattern))
		}
		str = str[1:]
		pattern = pattern[1:]
	}
	return len(str) == 0 && len(pattern) == 0
}

func deepMatchRune(str, pattern string) bool {
	var sr, pr rune
	var srsz, prsz int

	// read the first rune ahead of time
	if len(str) > 0 {
		if str[0] > 0x7f {
			sr, srsz = utf8.DecodeRuneInString(str)
		} else {
			sr, srsz = rune(str[0]), 1
		}
	} else {
		sr, srsz = utf8.RuneError, 0
	}
	if len(pattern) > 0 {
		if pattern[0] > 0x7f {
			pr, prsz = utf8.DecodeRuneInString(pattern)
		} else {
			pr, prsz = rune(pattern[0]), 1
		}
	} else {
		pr, prsz = utf8.RuneError, 0
	}
	// done reading
	for pr != utf8.RuneError {
		switch pr {
		default:
			if srsz == utf8.RuneError {
				return false
			}
			if sr != pr {
				return false
			}
		case '?':
			if srsz == utf8.RuneError {
				return false
			}
		case '*':
			return deepMatchRune(str, pattern[prsz:]) ||
				(srsz > 0 && deepMatchRune(str[srsz:], pattern))
		}
		str = str[srsz:]
		pattern = pattern[prsz:]
		// read the next runes
		if len(str) > 0 {
			if str[0] > 0x7f {
				sr, srsz = utf8.DecodeRuneInString(str)
			} else {
				sr, srsz = rune(str[0]), 1
			}
		} else {
			sr, srsz = utf8.RuneError, 0
		}
		if len(pattern) > 0 {
			if pattern[0] > 0x7f {
				pr, prsz = utf8.DecodeRuneInString(pattern)
			} else {
				pr, prsz = rune(pattern[0]), 1
			}
		} else {
			pr, prsz = utf8.RuneError, 0
		}
		// done reading
	}

	return srsz == 0 && prsz == 0
}

var maxRuneBytes = func() []byte {
	b := make([]byte, 4)
	if utf8.EncodeRune(b, '\U0010FFFF') != 4 {
		panic("invalid rune encoding")
	}
	return b
}()

func Allowable(pattern string) (min, max string) {
	if pattern == "" || pattern[0] == '*' {
		return "", ""
	}

	minb := make([]byte, 0, len(pattern))
	maxb := make([]byte, 0, len(pattern))
	var wild bool
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '*' {
			wild = true
			break
		}
		if pattern[i] == '?' {
			minb = append(minb, 0)
			maxb = append(maxb, maxRuneBytes...)
		} else {
			minb = append(minb, pattern[i])
			maxb = append(maxb, pattern[i])
		}
	}
	if wild {
		r, n := utf8.DecodeLastRune(maxb)
		if r != utf8.RuneError {
			if r < utf8.MaxRune {
				r++
				if r > 0x7f {
					b := make([]byte, 4)
					nn := utf8.EncodeRune(b, r)
					maxb = append(maxb[:len(maxb)-n], b[:nn]...)
				} else {
					maxb = append(maxb[:len(maxb)-n], byte(r))
				}
			}
		}
	}
	return string(minb), string(maxb)
}

// IsPattern returns true if the string is a pattern.
func IsPattern(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] == '*' || str[i] == '?' {
			return true
		}
	}
	return false
}

/*********** Json Pretty ***********/
type Options struct {
	// Width is an max column width for single line arrays
	// Default is 80
	Width int
	// Prefix is a prefix for all lines
	// Default is an empty string
	Prefix string
	// Indent is the nested indentation
	// Default is two spaces
	Indent string
	// SortKeys will sort the keys alphabetically
	// Default is false
	SortKeys bool
}

var DefaultOptions = &Options{Width: 80, Prefix: "", Indent: "  ", SortKeys: false}

func Pretty(json []byte) []byte { return PrettyOptions(json, nil) }

func PrettyOptions(json []byte, opts *Options) []byte {
	if opts == nil {
		opts = DefaultOptions
	}
	buf := make([]byte, 0, len(json))
	if len(opts.Prefix) != 0 {
		buf = append(buf, opts.Prefix...)
	}
	buf, _, _, _ = appendPrettyAny(buf, json, 0, true,
		opts.Width, opts.Prefix, opts.Indent, opts.SortKeys,
		0, 0, -1)
	if len(buf) > 0 {
		buf = append(buf, '\n')
	}
	return buf
}

func Ugly(json []byte) []byte {
	buf := make([]byte, 0, len(json))
	return ugly(buf, json)
}

func UglyInPlace(json []byte) []byte { return ugly(json, json) }

func ugly(dst, src []byte) []byte {
	dst = dst[:0]
	for i := 0; i < len(src); i++ {
		if src[i] > ' ' {
			dst = append(dst, src[i])
			if src[i] == '"' {
				for i = i + 1; i < len(src); i++ {
					dst = append(dst, src[i])
					if src[i] == '"' {
						j := i - 1
						for ; ; j-- {
							if src[j] != '\\' {
								break
							}
						}
						if (j-i)%2 != 0 {
							break
						}
					}
				}
			}
		}
	}
	return dst
}

func appendPrettyAny(buf, json []byte, i int, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == '"' {
			return appendPrettyString(buf, json, i, nl)
		}
		if (json[i] >= '0' && json[i] <= '9') || json[i] == '-' {
			return appendPrettyNumber(buf, json, i, nl)
		}
		if json[i] == '{' {
			return appendPrettyObject(buf, json, i, '{', '}', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		if json[i] == '[' {
			return appendPrettyObject(buf, json, i, '[', ']', pretty, width, prefix, indent, sortkeys, tabs, nl, max)
		}
		switch json[i] {
		case 't':
			return append(buf, 't', 'r', 'u', 'e'), i + 4, nl, true
		case 'f':
			return append(buf, 'f', 'a', 'l', 's', 'e'), i + 5, nl, true
		case 'n':
			return append(buf, 'n', 'u', 'l', 'l'), i + 4, nl, true
		}
	}
	return buf, i, nl, true
}

type pair struct {
	kstart, kend int
	vstart, vend int
}

type byKey struct {
	sorted bool
	json   []byte
	pairs  []pair
}

func (arr *byKey) Len() int {
	return len(arr.pairs)
}
func (arr *byKey) Less(i, j int) bool {
	key1 := arr.json[arr.pairs[i].kstart+1 : arr.pairs[i].kend-1]
	key2 := arr.json[arr.pairs[j].kstart+1 : arr.pairs[j].kend-1]
	return string(key1) < string(key2)
}
func (arr *byKey) Swap(i, j int) {
	arr.pairs[i], arr.pairs[j] = arr.pairs[j], arr.pairs[i]
	arr.sorted = true
}

func appendPrettyObject(buf, json []byte, i int, open, close byte, pretty bool, width int, prefix, indent string, sortkeys bool, tabs, nl, max int) ([]byte, int, int, bool) {
	var ok bool
	if width > 0 {
		if pretty && open == '[' && max == -1 {
			// here we try to create a single line array
			max := width - (len(buf) - nl)
			if max > 3 {
				s1, s2 := len(buf), i
				buf, i, _, ok = appendPrettyObject(buf, json, i, '[', ']', false, width, prefix, "", sortkeys, 0, 0, max)
				if ok && len(buf)-s1 <= max {
					return buf, i, nl, true
				}
				buf = buf[:s1]
				i = s2
			}
		} else if max != -1 && open == '{' {
			return buf, i, nl, false
		}
	}
	buf = append(buf, open)
	i++
	var pairs []pair
	if open == '{' && sortkeys {
		pairs = make([]pair, 0, 8)
	}
	var n int
	for ; i < len(json); i++ {
		if json[i] <= ' ' {
			continue
		}
		if json[i] == close {
			if pretty {
				if open == '{' && sortkeys {
					buf = sortPairs(json, buf, pairs)
				}
				if n > 0 {
					nl = len(buf)
					buf = append(buf, '\n')
				}
				if buf[len(buf)-1] != open {
					buf = appendTabs(buf, prefix, indent, tabs)
				}
			}
			buf = append(buf, close)
			return buf, i + 1, nl, open != '{'
		}
		if open == '[' || json[i] == '"' {
			if n > 0 {
				buf = append(buf, ',')
				if width != -1 && open == '[' {
					buf = append(buf, ' ')
				}
			}
			var p pair
			if pretty {
				nl = len(buf)
				buf = append(buf, '\n')
				if open == '{' && sortkeys {
					p.kstart = i
					p.vstart = len(buf)
				}
				buf = appendTabs(buf, prefix, indent, tabs+1)
			}
			if open == '{' {
				buf, i, nl, _ = appendPrettyString(buf, json, i, nl)
				if sortkeys {
					p.kend = i
				}
				buf = append(buf, ':')
				if pretty {
					buf = append(buf, ' ')
				}
			}
			buf, i, nl, ok = appendPrettyAny(buf, json, i, pretty, width, prefix, indent, sortkeys, tabs+1, nl, max)
			if max != -1 && !ok {
				return buf, i, nl, false
			}
			if pretty && open == '{' && sortkeys {
				p.vend = len(buf)
				if p.kstart > p.kend || p.vstart > p.vend {
					// bad data. disable sorting
					sortkeys = false
				} else {
					pairs = append(pairs, p)
				}
			}
			i--
			n++
		}
	}
	return buf, i, nl, open != '{'
}

func sortPairs(json, buf []byte, pairs []pair) []byte {
	if len(pairs) == 0 {
		return buf
	}
	vstart := pairs[0].vstart
	vend := pairs[len(pairs)-1].vend
	arr := byKey{false, json, pairs}
	sort.Sort(&arr)
	if !arr.sorted {
		return buf
	}
	nbuf := make([]byte, 0, vend-vstart)
	for i, p := range pairs {
		nbuf = append(nbuf, buf[p.vstart:p.vend]...)
		if i < len(pairs)-1 {
			nbuf = append(nbuf, ',')
			nbuf = append(nbuf, '\n')
		}
	}
	return append(buf[:vstart], nbuf...)
}

func appendPrettyString(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] == '"' {
			var sc int
			for j := i - 1; j > s; j-- {
				if json[j] == '\\' {
					sc++
				} else {
					break
				}
			}
			if sc%2 == 1 {
				continue
			}
			i++
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendPrettyNumber(buf, json []byte, i, nl int) ([]byte, int, int, bool) {
	s := i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ':' || json[i] == ']' || json[i] == '}' {
			break
		}
	}
	return append(buf, json[s:i]...), i, nl, true
}

func appendTabs(buf []byte, prefix, indent string, tabs int) []byte {
	if len(prefix) != 0 {
		buf = append(buf, prefix...)
	}
	if len(indent) == 2 && indent[0] == ' ' && indent[1] == ' ' {
		for i := 0; i < tabs; i++ {
			buf = append(buf, ' ', ' ')
		}
	} else {
		for i := 0; i < tabs; i++ {
			buf = append(buf, indent...)
		}
	}
	return buf
}

// Style is the color style
type Style struct {
	Key, String, Number [2]string
	True, False, Null   [2]string
	Append              func(dst []byte, c byte) []byte
}

func hexp(p byte) byte {
	switch {
	case p < 10:
		return p + '0'
	default:
		return (p - 10) + 'a'
	}
}

var TerminalStyle = &Style{
	Key:    [2]string{"\x1B[94m", "\x1B[0m"},
	String: [2]string{"\x1B[92m", "\x1B[0m"},
	Number: [2]string{"\x1B[93m", "\x1B[0m"},
	True:   [2]string{"\x1B[96m", "\x1B[0m"},
	False:  [2]string{"\x1B[96m", "\x1B[0m"},
	Null:   [2]string{"\x1B[91m", "\x1B[0m"},
	Append: func(dst []byte, c byte) []byte {
		if c < ' ' && (c != '\r' && c != '\n' && c != '\t' && c != '\v') {
			dst = append(dst, "\\u00"...)
			dst = append(dst, hexp((c>>4)&0xF))
			return append(dst, hexp((c)&0xF))
		}
		return append(dst, c)
	},
}

// Color will colorize the json. The style parma is used for customizing
// the colors. Passing nil to the style param will use the default
// TerminalStyle.
func Color(src []byte, style *Style) []byte {
	if style == nil {
		style = TerminalStyle
	}
	apnd := style.Append
	if apnd == nil {
		apnd = func(dst []byte, c byte) []byte {
			return append(dst, c)
		}
	}
	type stackt struct {
		kind byte
		key  bool
	}
	var dst []byte
	var stack []stackt
	for i := 0; i < len(src); i++ {
		if src[i] == '"' {
			key := len(stack) > 0 && stack[len(stack)-1].key
			if key {
				dst = append(dst, style.Key[0]...)
			} else {
				dst = append(dst, style.String[0]...)
			}
			dst = apnd(dst, '"')
			for i = i + 1; i < len(src); i++ {
				dst = apnd(dst, src[i])
				if src[i] == '"' {
					j := i - 1
					for ; ; j-- {
						if src[j] != '\\' {
							break
						}
					}
					if (j-i)%2 != 0 {
						break
					}
				}
			}
			if key {
				dst = append(dst, style.Key[1]...)
			} else {
				dst = append(dst, style.String[1]...)
			}
		} else if src[i] == '{' || src[i] == '[' {
			stack = append(stack, stackt{src[i], src[i] == '{'})
			dst = apnd(dst, src[i])
		} else if (src[i] == '}' || src[i] == ']') && len(stack) > 0 {
			stack = stack[:len(stack)-1]
			dst = apnd(dst, src[i])
		} else if (src[i] == ':' || src[i] == ',') && len(stack) > 0 && stack[len(stack)-1].kind == '{' {
			stack[len(stack)-1].key = !stack[len(stack)-1].key
			dst = apnd(dst, src[i])
		} else {
			var kind byte
			if (src[i] >= '0' && src[i] <= '9') || src[i] == '-' {
				kind = '0'
				dst = append(dst, style.Number[0]...)
			} else if src[i] == 't' {
				kind = 't'
				dst = append(dst, style.True[0]...)
			} else if src[i] == 'f' {
				kind = 'f'
				dst = append(dst, style.False[0]...)
			} else if src[i] == 'n' {
				kind = 'n'
				dst = append(dst, style.Null[0]...)
			} else {
				dst = apnd(dst, src[i])
			}
			if kind != 0 {
				for ; i < len(src); i++ {
					if src[i] <= ' ' || src[i] == ',' || src[i] == ':' || src[i] == ']' || src[i] == '}' {
						i--
						break
					}
					dst = apnd(dst, src[i])
				}
				if kind == '0' {
					dst = append(dst, style.Number[1]...)
				} else if kind == 't' {
					dst = append(dst, style.True[1]...)
				} else if kind == 'f' {
					dst = append(dst, style.False[1]...)
				} else if kind == 'n' {
					dst = append(dst, style.Null[1]...)
				}
			}
		}
	}
	return dst
}

/*********** Private Method ***********/
func getBytes(json []byte, path string) JsonItem {
	var JsonItem JsonItem
	if json != nil {
		// unsafe cast to string
		JsonItem = Get(*(*string)(unsafe.Pointer(&json)), path)
		// safely get the string headers
		rawhi := *(*reflect.StringHeader)(unsafe.Pointer(&JsonItem.Raw))
		strhi := *(*reflect.StringHeader)(unsafe.Pointer(&JsonItem.Str))
		// create byte slice headers
		rawh := reflect.SliceHeader{Data: rawhi.Data, Len: rawhi.Len}
		strh := reflect.SliceHeader{Data: strhi.Data, Len: strhi.Len}
		if strh.Data == 0 {
			// str is nil
			if rawh.Data == 0 {
				// raw is nil
				JsonItem.Raw = ""
			} else {
				// raw has data, safely copy the slice header to a string
				JsonItem.Raw = string(*(*[]byte)(unsafe.Pointer(&rawh)))
			}
			JsonItem.Str = ""
		} else if rawh.Data == 0 {
			// raw is nil
			JsonItem.Raw = ""
			// str has data, safely copy the slice header to a string
			JsonItem.Str = string(*(*[]byte)(unsafe.Pointer(&strh)))
		} else if strh.Data >= rawh.Data &&
			int(strh.Data)+strh.Len <= int(rawh.Data)+rawh.Len {
			// Str is a substring of Raw.
			start := int(strh.Data - rawh.Data)
			// safely copy the raw slice header
			JsonItem.Raw = string(*(*[]byte)(unsafe.Pointer(&rawh)))
			// substring the raw
			JsonItem.Str = JsonItem.Raw[start : start+strh.Len]
		} else {
			// safely copy both the raw and str slice headers to strings
			JsonItem.Raw = string(*(*[]byte)(unsafe.Pointer(&rawh)))
			JsonItem.Str = string(*(*[]byte)(unsafe.Pointer(&strh)))
		}
	}
	return JsonItem
}

func fillIndex(json string, c *parseContext) {
	if len(c.value.Raw) > 0 && !c.calcd {
		jhdr := *(*reflect.StringHeader)(unsafe.Pointer(&json))
		rhdr := *(*reflect.StringHeader)(unsafe.Pointer(&(c.value.Raw)))
		c.value.Index = int(rhdr.Data - jhdr.Data)
		if c.value.Index < 0 || c.value.Index >= len(json) {
			c.value.Index = 0
		}
	}
}

func stringBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: (*reflect.StringHeader)(unsafe.Pointer(&s)).Data,
		Len:  len(s),
		Cap:  len(s),
	}))
}

func bytesString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
