package main

var scn *Scanner
var tok Token
var dc int64
var level int
var modid string

func next() {
	tok = scn.get()
}

func check(s int, msg string) {
	if tok.Sym == s {
		next()
	} else {
		scn.mark(tok.Line, tok.Col, msg)
	}
}

func identList(class int) *ObjDesc {
	if tok.Sym != symIdent {
		scn.mark(tok.Line, tok.Col, "identifier expected")

		return nil
	}

	obj, isDup := newObj(tok.Lexeme, class)

	if isDup {
		scn.mark(tok.Line, tok.Col, "mult def")
	}

	first := obj

	next()

	for tok.Sym == symComma {
		next()

		if tok.Sym == symIdent {
			obj, isDup = newObj(tok.Lexeme, class)

			if isDup {
				scn.mark(tok.Line, tok.Col, "mult def")
			}

			next()
		} else {
			scn.mark(tok.Line, tok.Col, "identifier expected")
		}
	}

	check(symColon, "':' expected")

	return first
}

func type_() *TypeDesc {
	switch tok.Sym {
		case symKwInteger:
			next()

			return intType

		case symKwStringType:
			next()

			return strType

		default:
			scn.mark(tok.Line, tok.Col, "type expected ('integer' or 'string')")

			return noType
	}
}

func declarations() {
	if tok.Sym != symKwVar && tok.Sym != symKwBegin {
		scn.mark(tok.Line, tok.Col, "'var' or 'begin' expected")

		for tok.Sym != symKwVar && tok.Sym != symKwBegin && tok.Sym != symEOF {
			next()
		}
	}

	if tok.Sym == symKwVar {
		next()

		for tok.Sym == symIdent {
			first := identList(Var)
			tp := type_()
			obj := first

			for obj != nil {
				if obj.Name == modid {
					scn.mark(tok.Line, tok.Col, "variable name same as program name")
				}

				obj.Type = tp
				obj.Lev = level

				if tp.Size > 1 {
					dc = (dc + 7) / 8 * 8
				}

				obj.Val = dc
				dc += tp.Size
				obj = obj.Next
			}

			check(symSemicolon, "';' expected")
		}

		dc = (dc + 7) / 8 * 8
	}
}

func factor(x *Item) {
	if tok.Sym == symIdent {
		name := tok.Lexeme

		next()

		obj := thisObj(name)

		if obj == nil {
			scn.mark(tok.Line, tok.Col, "undef: "+name)

			makeConstItem(x, intType, 0)
		} else {
			makeItem(x, obj)
		}
	} else if tok.Sym == symNumber {
		makeConstItem(x, intType, scn.ival)

		next()
	} else {
		scn.mark(tok.Line, tok.Col, "identifier or number expected")

		makeConstItem(x, intType, 0)

		next()
	}
}

func statSequence() {
	for tok.Sym != symKwEnd && tok.Sym != symEOF {
		if tok.Sym == symIdent {
			name := tok.Lexeme

			next()

			obj := thisObj(name)

			if obj == nil {
				scn.mark(tok.Line, tok.Col, "undef: "+name)
			}

			var x Item

			if obj != nil {
				makeItem(&x, obj)
			} else {
				makeConstItem(&x, noType, 0)
			}

			check(symAssign, "':=' expected")

			var y Item

			factor(&y)

			if tok.Sym == symPlus || tok.Sym == symMinus {
				op := tok.Sym

				next()

				var z Item

				factor(&z)

				addOp(op, &y, &z)
			}

			if obj != nil {
				store(&x, &y)
			}

		} else if tok.Sym == symKwWrite {
			next()

			check(symLParen, "'(' expected")

			if tok.Sym == symIdent {
				name := tok.Lexeme

				next()

				obj := thisObj(name)

				if obj == nil {
					scn.mark(tok.Line, tok.Col, "undef: "+name)
				} else {
					var x Item

					makeItem(&x, obj)

					writeCall(&x)
				}
			} else {
				scn.mark(tok.Line, tok.Col, "identifier expected")
			}

			check(symRParen, "')' expected")

		} else {
			scn.mark(tok.Line, tok.Col, "statement expected")

			next()
		}

		checkRegs()

		check(symSemicolon, "';' expected")
	}
}

func module() {
	check(symKwProgram, "'program' expected")

	initScope()
	openScope()

	if tok.Sym == symIdent {
		modid = tok.Lexeme

		next()
	} else {
		scn.mark(tok.Line, tok.Col, "program name expected")
	}

	check(symSemicolon, "';' expected")

	level = 0
	dc = 0

	declarations()

	check(symKwBegin, "'begin' expected")

	statSequence()

	check(symKwEnd, "'end' expected")

	if tok.Sym != symPeriod {
		scn.mark(tok.Line, tok.Col, "'.' expected")
	}

	closeScope()
}

// https://people.inf.ethz.ch/wirth/ProjectOberon/Sources/ORP.Mod.txt