package eval

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"text/scanner"
	"time"
)

func Evaluate(expression string, parameter interface{}, opts ...Language) (interface{}, error) {
	l := full
	if len(opts) > 0 {
		l = NewLanguage(append([]Language{l}, opts...)...)
	}
	return l.Evaluate(expression, parameter)
}


func Full(extensions ...Language) Language {
	if len(extensions) == 0 {
		return full
	}
	return NewLanguage(append([]Language{full}, extensions...)...)
}


func Arithmetic() Language {
	return arithmetic
}


func Bitmask() Language {
	return bitmask
}


func Text() Language {
	return text
}


func PropositionalLogic() Language {
	return propositionalLogic
}


func JSON() Language {
	return ljson
}


func Base() Language {
	return base
}

var full = NewLanguage(arithmetic, bitmask, text, propositionalLogic, ljson,

	InfixOperator("in", inArray),

	InfixShortCircuit("??", func(a interface{}) (interface{}, bool) {
		return a, a != false && a != nil
	}),
	InfixOperator("??", func(a, b interface{}) (interface{}, error) {
		if a == false || a == nil {
			return b, nil
		}
		return a, nil
	}),

	PostfixOperator("?", parseIf),

	Function("date", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("date() expects exactly one string argument")
		}
		s, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("date() expects exactly one string argument")
		}
		for _, format := range [...]string{
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
			time.Kitchen,
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02",                         // RFC 3339
			"2006-01-02 15:04",                   // RFC 3339 with minutes
			"2006-01-02 15:04:05",                // RFC 3339 with seconds
			"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
			"2006-01-02T15Z0700",                 // ISO8601 with hour
			"2006-01-02T15:04Z0700",              // ISO8601 with minutes
			"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
			"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
		} {
			ret, err := time.ParseInLocation(format, s, time.Local)
			if err == nil {
				return ret, nil
			}
		}
		return nil, fmt.Errorf("date() could not parse %s", s)
	}),
)

var ljson = NewLanguage(
	PrefixExtension('[', parseJSONArray),
	PrefixExtension('{', parseJSONObject),
)

var arithmetic = NewLanguage(
	InfixNumberOperator("+", func(a, b float64) (interface{}, error) { return a + b, nil }),
	InfixNumberOperator("-", func(a, b float64) (interface{}, error) { return a - b, nil }),
	InfixNumberOperator("*", func(a, b float64) (interface{}, error) { return a * b, nil }),
	InfixNumberOperator("/", func(a, b float64) (interface{}, error) { return a / b, nil }),
	InfixNumberOperator("%", func(a, b float64) (interface{}, error) { return math.Mod(a, b), nil }),
	InfixNumberOperator("**", func(a, b float64) (interface{}, error) { return math.Pow(a, b), nil }),

	InfixNumberOperator(">", func(a, b float64) (interface{}, error) { return a > b, nil }),
	InfixNumberOperator(">=", func(a, b float64) (interface{}, error) { return a >= b, nil }),
	InfixNumberOperator("<", func(a, b float64) (interface{}, error) { return a < b, nil }),
	InfixNumberOperator("<=", func(a, b float64) (interface{}, error) { return a <= b, nil }),

	InfixNumberOperator("==", func(a, b float64) (interface{}, error) { return a == b, nil }),
	InfixNumberOperator("!=", func(a, b float64) (interface{}, error) { return a != b, nil }),

	base,
)

var bitmask = NewLanguage(
	InfixNumberOperator("^", func(a, b float64) (interface{}, error) { return float64(int64(a) ^ int64(b)), nil }),
	InfixNumberOperator("&", func(a, b float64) (interface{}, error) { return float64(int64(a) & int64(b)), nil }),
	InfixNumberOperator("|", func(a, b float64) (interface{}, error) { return float64(int64(a) | int64(b)), nil }),
	InfixNumberOperator("<<", func(a, b float64) (interface{}, error) { return float64(int64(a) << uint64(b)), nil }),
	InfixNumberOperator(">>", func(a, b float64) (interface{}, error) { return float64(int64(a) >> uint64(b)), nil }),

	PrefixOperator("~", func(c context.Context, v interface{}) (interface{}, error) {
		i, ok := convertToFloat(v)
		if !ok {
			return nil, fmt.Errorf("unexpected %T expected number", v)
		}
		return float64(^int64(i)), nil
	}),
)

var text = NewLanguage(
	InfixTextOperator("+", func(a, b string) (interface{}, error) { return fmt.Sprintf("%v%v", a, b), nil }),

	InfixTextOperator("<", func(a, b string) (interface{}, error) { return a < b, nil }),
	InfixTextOperator("<=", func(a, b string) (interface{}, error) { return a <= b, nil }),
	InfixTextOperator(">", func(a, b string) (interface{}, error) { return a > b, nil }),
	InfixTextOperator(">=", func(a, b string) (interface{}, error) { return a >= b, nil }),

	InfixEvalOperator("=~", regEx),
	InfixEvalOperator("!~", notRegEx),
	base,
)

var propositionalLogic = NewLanguage(
	PrefixOperator("!", func(c context.Context, v interface{}) (interface{}, error) {
		b, ok := convertToBool(v)
		if !ok {
			return nil, fmt.Errorf("unexpected %T expected bool", v)
		}
		return !b, nil
	}),

	InfixShortCircuit("&&", func(a interface{}) (interface{}, bool) { return false, a == false }),
	InfixBoolOperator("&&", func(a, b bool) (interface{}, error) { return a && b, nil }),
	InfixShortCircuit("||", func(a interface{}) (interface{}, bool) { return true, a == true }),
	InfixBoolOperator("||", func(a, b bool) (interface{}, error) { return a || b, nil }),

	InfixBoolOperator("==", func(a, b bool) (interface{}, error) { return a == b, nil }),
	InfixBoolOperator("!=", func(a, b bool) (interface{}, error) { return a != b, nil }),

	base,
)

var base = NewLanguage(
	PrefixExtension(scanner.Int, parseNumber),
	PrefixExtension(scanner.Float, parseNumber),
	PrefixOperator("-", func(c context.Context, v interface{}) (interface{}, error) {
		i, ok := convertToFloat(v)
		if !ok {
			return nil, fmt.Errorf("unexpected %v(%T) expected number", v, v)
		}
		return -i, nil
	}),

	PrefixExtension(scanner.String, parseString),
	PrefixExtension(scanner.Char, parseString),
	PrefixExtension(scanner.RawString, parseString),

	Constant("true", true),
	Constant("false", false),

	InfixOperator("==", func(a, b interface{}) (interface{}, error) { return reflect.DeepEqual(a, b), nil }),
	InfixOperator("!=", func(a, b interface{}) (interface{}, error) { return !reflect.DeepEqual(a, b), nil }),
	PrefixExtension('(', parseParentheses),

	Precedence("??", 0),

	Precedence("||", 20),
	Precedence("&&", 21),

	Precedence("==", 40),
	Precedence("!=", 40),
	Precedence(">", 40),
	Precedence(">=", 40),
	Precedence("<", 40),
	Precedence("<=", 40),
	Precedence("=~", 40),
	Precedence("!~", 40),
	Precedence("in", 40),

	Precedence("^", 60),
	Precedence("&", 60),
	Precedence("|", 60),

	Precedence("<<", 90),
	Precedence(">>", 90),

	Precedence("+", 120),
	Precedence("-", 120),

	Precedence("*", 150),
	Precedence("/", 150),
	Precedence("%", 150),

	Precedence("**", 200),

	PrefixMetaPrefix(scanner.Ident, parseIdent),
)
