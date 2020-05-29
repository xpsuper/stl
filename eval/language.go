package eval

import (
	"context"
	"fmt"
	"text/scanner"
	"unicode"
)

type Language struct {
	prefixes        map[interface{}]prefix
	operators       map[string]operator
	operatorSymbols map[rune]struct{}
	selector        func(Evaluables) Evaluable
}

func NewLanguage(bases ...Language) Language {
	l := newLanguage()
	for _, base := range bases {
		for i, e := range base.prefixes {
			l.prefixes[i] = e
		}
		for i, e := range base.operators {
			l.operators[i] = e.merge(l.operators[i])
			l.operators[i].initiate(i)
		}
		for i := range base.operatorSymbols {
			l.operatorSymbols[i] = struct{}{}
		}
		if base.selector != nil {
			l.selector = base.selector
		}
	}
	return l
}

func newLanguage() Language {
	return Language{
		prefixes:        map[interface{}]prefix{},
		operators:       map[string]operator{},
		operatorSymbols: map[rune]struct{}{},
	}
}

func (l Language) NewEvaluable(expression string) (Evaluable, error) {
	p := newParser(expression, l)

	eval, err := p.ParseExpression(context.Background())

	if err == nil && p.isCamouflaged() && p.lastScan != scanner.EOF {
		err = p.camouflage
	}

	if err != nil {
		pos := p.scanner.Pos()
		return nil, fmt.Errorf("parsing error: %s - %d:%d %s", p.scanner.Position, pos.Line, pos.Column, err)
	}
	return eval, nil
}

func (l Language) Evaluate(expression string, parameter interface{}) (interface{}, error) {
	eval, err := l.NewEvaluable(expression)
	if err != nil {
		return nil, err
	}
	v, err := eval(context.Background(), parameter)
	if err != nil {
		return nil, fmt.Errorf("can not evaluate %s: %v", expression, err)
	}
	return v, nil
}

func Function(name string, function interface{}) Language {
	l := newLanguage()
	l.prefixes[name] = func(c context.Context, p *Parser) (eval Evaluable, err error) {
		args := []Evaluable{}
		scan := p.Scan()
		switch scan {
		case '(':
			args, err = p.parseArguments(c)
			if err != nil {
				return nil, err
			}
		default:
			p.Camouflage("function call", '(')
		}
		return p.callFunc(toFunc(function), args...), nil
	}
	return l
}

func Constant(name string, value interface{}) Language {
	l := newLanguage()
	l.prefixes[l.makePrefixKey(name)] = func(c context.Context, p *Parser) (eval Evaluable, err error) {
		return p.Const(value), nil
	}
	return l
}

func PrefixExtension(r rune, ext func(context.Context, *Parser) (Evaluable, error)) Language {
	l := newLanguage()
	l.prefixes[r] = ext
	return l
}

func PrefixMetaPrefix(r rune, ext func(context.Context, *Parser) (call string, alternative func() (Evaluable, error), err error)) Language {
	l := newLanguage()
	l.prefixes[r] = func(c context.Context, p *Parser) (Evaluable, error) {
		call, alternative, err := ext(c, p)
		if err != nil {
			return nil, err
		}
		if prefix, ok := p.prefixes[l.makePrefixKey(call)]; ok {
			return prefix(c, p)
		}
		return alternative()
	}
	return l
}

func PrefixOperator(name string, e Evaluable) Language {
	l := newLanguage()
	l.prefixes[l.makePrefixKey(name)] = func(c context.Context, p *Parser) (Evaluable, error) {
		eval, err := p.ParseNextExpression(c)
		if err != nil {
			return nil, err
		}
		prefix := func(c context.Context, v interface{}) (interface{}, error) {
			a, err := eval(c, v)
			if err != nil {
				return nil, err
			}
			return e(c, a)
		}
		if eval.IsConst() {
			v, err := prefix(context.Background(), nil)
			if err != nil {
				return nil, err
			}
			prefix = p.Const(v)
		}
		return prefix, nil
	}
	return l
}

func PostfixOperator(name string, ext func(context.Context, *Parser, Evaluable) (Evaluable, error)) Language {
	l := newLanguage()
	l.operators[l.makeInfixKey(name)] = postfix{
		f: func(c context.Context, p *Parser, eval Evaluable, pre operatorPrecedence) (Evaluable, error) {
			return ext(c, p, eval)
		},
	}
	return l
}

func InfixOperator(name string, f func(a, b interface{}) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{arbitrary: f})
}

func InfixShortCircuit(name string, f func(a interface{}) (interface{}, bool)) Language {
	return newLanguageOperator(name, &infix{shortCircuit: f})
}

func InfixTextOperator(name string, f func(a, b string) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{text: f})
}

func InfixNumberOperator(name string, f func(a, b float64) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{number: f})
}

func InfixBoolOperator(name string, f func(a, b bool) (interface{}, error)) Language {
	return newLanguageOperator(name, &infix{boolean: f})
}

func Precedence(name string, operatorPrecendence uint8) Language {
	return newLanguageOperator(name, operatorPrecedence(operatorPrecendence))
}

func InfixEvalOperator(name string, f func(a, b Evaluable) (Evaluable, error)) Language {
	return newLanguageOperator(name, directInfix{infixBuilder: f})
}

func newLanguageOperator(name string, op operator) Language {
	op.initiate(name)
	l := newLanguage()
	l.operators[l.makeInfixKey(name)] = op
	return l
}

func (l *Language) makePrefixKey(key string) interface{} {
	runes := []rune(key)
	if len(runes) == 1 && !unicode.IsLetter(runes[0]) {
		return runes[0]
	}
	return key
}

func (l *Language) makeInfixKey(key string) string {
	runes := []rune(key)
	for _, r := range runes {
		l.operatorSymbols[r] = struct{}{}
	}
	return key
}

func VariableSelector(selector func(path Evaluables) Evaluable) Language {
	l := newLanguage()
	l.selector = selector
	return l
}