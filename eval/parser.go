package eval

import (
	"bytes"
	"fmt"
	"strings"
	"text/scanner"
	"unicode"
)

type Parser struct {
	scanner scanner.Scanner
	Language
	lastScan   rune
	camouflage error
}

func newParser(expression string, l Language) *Parser {
	sc := scanner.Scanner{}
	sc.Init(strings.NewReader(expression))
	sc.Error = func(*scanner.Scanner, string) { return }
	sc.IsIdentRune = func(r rune, pos int) bool { return unicode.IsLetter(r) || r == '_' || (pos > 0 && unicode.IsDigit(r)) }
	sc.Filename = expression + "\t"
	return &Parser{scanner: sc, Language: l}
}

func (p *Parser) Scan() rune {
	if p.isCamouflaged() {
		p.camouflage = nil
		return p.lastScan
	}
	p.camouflage = nil
	p.lastScan = p.scanner.Scan()
	return p.lastScan
}

func (p *Parser) isCamouflaged() bool {
	return p.camouflage != nil && p.camouflage != errCamouflageAfterNext
}

func (p *Parser) Camouflage(unit string, expected ...rune) {
	if p.isCamouflaged() {
		panic(fmt.Errorf("can only Camouflage() after Scan(): %v", p.camouflage))
	}
	p.camouflage = p.Expected(unit, expected...)
	return
}

func (p *Parser) Peek() rune {
	if p.isCamouflaged() {
		panic("can not Peek() on camouflaged Parser")
	}
	return p.scanner.Peek()
}

var errCamouflageAfterNext = fmt.Errorf("Camouflage() after Next()")

func (p *Parser) Next() rune {
	if p.isCamouflaged() {
		panic("can not Next() on camouflaged Parser")
	}
	p.camouflage = errCamouflageAfterNext
	return p.scanner.Next()
}

func (p *Parser) TokenText() string {
	return p.scanner.TokenText()
}

func (p *Parser) Expected(unit string, expected ...rune) error {
	return unexpectedRune{unit, expected, p.lastScan}
}

type unexpectedRune struct {
	unit     string
	expected []rune
	got      rune
}

func (err unexpectedRune) Error() string {
	exp := bytes.Buffer{}
	runes := err.expected
	switch len(runes) {
	default:
		for _, r := range runes[:len(runes)-2] {
			exp.WriteString(scanner.TokenString(r))
			exp.WriteString(", ")
		}
		fallthrough
	case 2:
		exp.WriteString(scanner.TokenString(runes[len(runes)-2]))
		exp.WriteString(" or ")
		fallthrough
	case 1:
		exp.WriteString(scanner.TokenString(runes[len(runes)-1]))
	case 0:
		return fmt.Sprintf("unexpected %s while scanning %s", scanner.TokenString(err.got), err.unit)
	}
	return fmt.Sprintf("unexpected %s while scanning %s expected %s", scanner.TokenString(err.got), err.unit, exp.String())
}
