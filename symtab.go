package main

const (
	Head  = 0
	Const = 1
	Var = 2
)

const (
	NoTyp = 0
	Int = 1
	Str = 2
)

type ObjDesc struct {
	Class int
	Lev int
	Next *ObjDesc
	Dsc *ObjDesc
	Type *TypeDesc
	Name string
	Val int64
}

type TypeDesc struct {
	Form int
	Size int64
}

var topScope *ObjDesc
var universe *ObjDesc

var progScope *ObjDesc
var intType *TypeDesc
var strType *TypeDesc
var noType *TypeDesc

func newObj(name string, class int) (*ObjDesc, bool) {
	x := topScope

	for x.Next != nil && x.Next.Name != name {
		x = x.Next
	}

	if x.Next == nil {
		n := &ObjDesc{Name: name, Class: class}

		x.Next = n

		return n, false
	}

	return x.Next, true
}

func thisObj(name string) *ObjDesc {
	s := topScope

	for s != nil {
		x := s.Next

		for x != nil && x.Name != name {
			x = x.Next
		}

		if x != nil {
			return x
		}

		s = s.Dsc
	}

	return nil
}

func openScope() {
	topScope = &ObjDesc{Class: Head, Dsc: topScope}
}

func closeScope() {
    progScope = topScope
    topScope = topScope.Dsc
}

func initScope() {
	topScope = universe
}

func init() {
	intType = &TypeDesc{Form: Int, Size: 8}
	strType = &TypeDesc{Form: Str, Size: 8}
	noType = &TypeDesc{Form: NoTyp, Size: 0}

	openScope()

	universe = topScope
}

// https://people.inf.ethz.ch/wirth/ProjectOberon/Sources/ORB.Mod.txt