package stl

import (
	"regexp"
	"sync"
)

type XPRegexpImpl struct {

}

// Quote quotes <s> by replacing special chars in <s>
// to match the rules of regular expression pattern.
// And returns the copy.
//
// Eg: Quote(`[foo]`) returns `\[foo\]`.
func (instance *XPRegexpImpl) Quote(s string) string {
	return regexp.QuoteMeta(s)
}

// Validate checks whether given regular expression pattern <pattern> valid.
func (instance *XPRegexpImpl) Validate(pattern string) error {
	_, err := getRegexp(pattern)
	return err
}

// IsMatch checks whether given bytes <src> matches <pattern>.
func (instance *XPRegexpImpl) IsMatch(pattern string, src []byte) bool {
	if r, err := getRegexp(pattern); err == nil {
		return r.Match(src)
	}
	return false
}

// IsMatchString checks whether given string <src> matches <pattern>.
func (instance *XPRegexpImpl) IsMatchString(pattern string, src string) bool {
	return instance.IsMatch(pattern, []byte(src))
}

// MatchString return bytes slice that matched <pattern>.
func (instance *XPRegexpImpl) Match(pattern string, src []byte) ([][]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.FindSubmatch(src), nil
	} else {
		return nil, err
	}
}

// MatchString return strings that matched <pattern>.
func (instance *XPRegexpImpl) MatchString(pattern string, src string) ([]string, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.FindStringSubmatch(src), nil
	} else {
		return nil, err
	}
}

// MatchAll return all bytes slices that matched <pattern>.
func (instance *XPRegexpImpl) MatchAll(pattern string, src []byte) ([][][]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.FindAllSubmatch(src, -1), nil
	} else {
		return nil, err
	}
}

// MatchAllString return all strings that matched <pattern>.
func (instance *XPRegexpImpl) MatchAllString(pattern string, src string) ([][]string, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.FindAllStringSubmatch(src, -1), nil
	} else {
		return nil, err
	}
}

// ReplaceString replace all matched <pattern> in bytes <src> with bytes <replace>.
func (instance *XPRegexpImpl) Replace(pattern string, replace, src []byte) ([]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.ReplaceAll(src, replace), nil
	} else {
		return nil, err
	}
}

// ReplaceString replace all matched <pattern> in string <src> with string <replace>.
func (instance *XPRegexpImpl) ReplaceString(pattern, replace, src string) (string, error) {
	r, e := instance.Replace(pattern, []byte(replace), []byte(src))
	return string(r), e
}

// ReplaceFunc replace all matched <pattern> in bytes <src>
// with custom replacement function <replaceFunc>.
func (instance *XPRegexpImpl) ReplaceFunc(pattern string, src []byte, replaceFunc func(b []byte) []byte) ([]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.ReplaceAllFunc(src, replaceFunc), nil
	} else {
		return nil, err
	}
}

// ReplaceFunc replace all matched <pattern> in bytes <src>
// with custom replacement function <replaceFunc>.
// The parameter <match> type for <replaceFunc> is [][]byte,
// which is the result contains all sub-patterns of <pattern> using Match function.
func (instance *XPRegexpImpl) ReplaceFuncMatch(pattern string, src []byte, replaceFunc func(match [][]byte) []byte) ([]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.ReplaceAllFunc(src, func(bytes []byte) []byte {
			match, _ := instance.Match(pattern, src)
			return replaceFunc(match)
		}), nil
	} else {
		return nil, err
	}
}

// ReplaceStringFunc replace all matched <pattern> in string <src>
// with custom replacement function <replaceFunc>.
func (instance *XPRegexpImpl) ReplaceStringFunc(pattern string, src string, replaceFunc func(s string) string) (string, error) {
	bytes, err := instance.ReplaceFunc(pattern, []byte(src), func(bytes []byte) []byte {
		return []byte(replaceFunc(string(bytes)))
	})
	return string(bytes), err
}

// ReplaceStringFuncMatch replace all matched <pattern> in string <src>
// with custom replacement function <replaceFunc>.
// The parameter <match> type for <replaceFunc> is []string,
// which is the result contains all sub-patterns of <pattern> using MatchString function.
func (instance *XPRegexpImpl) ReplaceStringFuncMatch(pattern string, src string, replaceFunc func(match []string) string) (string, error) {
	if r, err := getRegexp(pattern); err == nil {
		return string(r.ReplaceAllFunc([]byte(src), func(bytes []byte) []byte {
			match, _ := instance.MatchString(pattern, src)
			return []byte(replaceFunc(match))
		})), nil
	} else {
		return "", err
	}
}

// Split slices <src> into substrings separated by the expression and returns a slice of
// the substrings between those expression matches.
func (instance *XPRegexpImpl) Split(pattern string, src string) []string {
	if r, err := getRegexp(pattern); err == nil {
		return r.Split(src, -1)
	}
	return nil
}

var (
	regexMu  = sync.RWMutex{}
	regexMap = make(map[string]*regexp.Regexp)
)

func getRegexp(pattern string) (*regexp.Regexp, error) {
	if r := getCache(pattern); r != nil {
		return r, nil
	}
	if r, err := regexp.Compile(pattern); err == nil {
		setCache(pattern, r)
		return r, nil
	} else {
		return nil, err
	}
}

// getCache returns *regexp.Regexp object from cache by given <pattern>, for internal usage.
func getCache(pattern string) (regex *regexp.Regexp) {
	regexMu.RLock()
	regex = regexMap[pattern]
	regexMu.RUnlock()
	return
}

// setCache stores *regexp.Regexp object into cache, for internal usage.
func setCache(pattern string, regex *regexp.Regexp) {
	regexMu.Lock()
	regexMap[pattern] = regex
	regexMu.Unlock()
}
