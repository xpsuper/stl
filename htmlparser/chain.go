package htmlparser

import (
	"regexp"

	. "github.com/xpsuper/stl/htmlparser/expr"
)

func (n *Node) Satisfy(cs ...Checker) bool {
	return And(cs...)(n.InternalNode()) != nil
}

func (n *Node) Find(cs ...Checker) *Node {
	if n == nil {
		return nil
	}
	return NewNode(Find(cs...)(&n.n))
}

func (n *Node) FindNext(cs ...Checker) *Node {
	if n == nil {
		return nil
	}
	return NewNode(FindSibling(cs...)(&n.n))
}

func (n *Node) FindChild(cs ...Checker) *Node {
	return NewNode(FindChild(cs...)(&n.n))
}

func (n *Node) find(c Checker, cs []Checker) *Node {
	if n == nil {
		return nil
	}
	return n.Find(append([]Checker{c}, cs...)...)
}

func (n *Node) NextSibling() *Node {
	if n == nil {
		return nil
	}
	return NewNode(NextSibling(&n.n))
}

func (n *Node) PrevSibling() *Node {
	if n == nil {
		return nil
	}
	return NewNode(PrevSibling(&n.n))
}

func (n *Node) Parent() *Node {
	if n == nil {
		return nil
	}
	return NewNode(Parent(&n.n))
}

func (n *Node) Children(cs ...Checker) NodeIter {
	if n == nil {
		return NodeIter{nil}
	}
	return NodeIter{Children(&n.n, cs...)}
}

func (n *Node) Descendants(cs ...Checker) NodeIter {
	if n == nil {
		return NodeIter{nil}
	}
	return NodeIter{Descendants(&n.n, cs...)}
}

func (n *Node) Ahref(cs ...Checker) *Node {
	if n == nil {
		return nil
	}
	return n.find(Ahref, cs)
}

func (n *Node) TextNode(pat string) *TextNodeNode {
	if n == nil {
		return nil
	}
	rx := regexp.MustCompile(pat)
	cs := []Checker{Text_(rx)}
	return NewTextNodeNode(n.find(TextNode, cs), rx)
}

func also(c Checker, cs []Checker) []Checker {
	return append([]Checker{c}, cs...)
}
