package main

import "fmt"

var scn *Scanner
var tok Token
var dc int64
var level int
var modid string
var currentFuncRetName string
var returnAssigned bool

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

func typeMatch(a, b *TypeDesc) bool {
    if a == nil || b == nil {
        return true
    }

    if a.Form == NoTyp || b.Form == NoTyp {
        return true
    }
    
	return a.Form == b.Form
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
	if tok.Sym != symKwVar && tok.Sym != symKwBegin && tok.Sym != symKwProcedure && tok.Sym != symKwFunction {
		scn.mark(tok.Line, tok.Col, "'var', 'begin', 'procedure', or 'function' expected")

		for tok.Sym != symKwVar && tok.Sym != symKwBegin &&	tok.Sym != symKwProcedure && tok.Sym != symKwFunction && tok.Sym != symEOF {
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

func formalParams() []*ObjDesc {
	var params []*ObjDesc

	check(symLParen, "'(' expected")

	for tok.Sym == symIdent {
		var names []string

		for {
			if tok.Sym != symIdent {
				scn.mark(tok.Line, tok.Col, "identifier expected")

				break
			}

			names = append(names, tok.Lexeme)

			next()

			if tok.Sym != symComma {
				break
			}

			next()
		}

		check(symColon, "':' expected")

		tp := type_()

		for _, name := range names {
			obj, isDup := newObj(name, Par)

			if isDup {
				scn.mark(tok.Line, tok.Col, "duplicate parameter: "+name)
			}

			obj.Type = tp
			obj.Lev = level
			obj.Val = int64(-8 * (len(params) + 1))
			params = append(params, obj)
		}

		if tok.Sym != symSemicolon {
			break
		}

		next()
	}

	check(symRParen, "')' expected")

	return params
}

func parseLocalVarDefs(startOffset int64) int64 {
	offset := startOffset

	if tok.Sym != symKwVar {
		return offset
	}

	next()

	for tok.Sym == symIdent {
		var localObjs []*ObjDesc

		for {
			if tok.Sym != symIdent {
				scn.mark(tok.Line, tok.Col, "identifier expected")

				break
			}

			name := tok.Lexeme
			obj, isDup := newObj(name, Var)

			if isDup {
				scn.mark(tok.Line, tok.Col, "duplicate local variable: "+name)
			}

			obj.Lev = level
			localObjs = append(localObjs, obj)

			next()

			if tok.Sym != symComma {
				break
			}

			next()
		}

		check(symColon, "':' expected")

		tp := type_()

		for _, obj := range localObjs {
			obj.Type = tp
			offset -= 8
			obj.Val = offset
		}

		check(symSemicolon, "';' expected")
	}

	return offset
}

func evalArgsIntoRegs() []*Item {
    var args []*Item

    check(symLParen, "'(' expected")

    if tok.Sym != symRParen {
        var x Item

        factor(&x)

        load(&x)

        if len(args) < len(argRegs) {
            emit("    movq %rax, " + argRegs[len(args)])
        }

        args = append(args, &x)

        for tok.Sym == symComma {
            next()

            var y Item

            factor(&y)

            load(&y)

            if len(args) < len(argRegs) {
                emit("    movq %rax, " + argRegs[len(args)])
            }

            args = append(args, &y)
        }
    }

    check(symRParen, "')' expected")

    return args
}

func countParams(subObj *ObjDesc) int {
	if subObj.Dsc == nil {
		return 0
	}

	n := 0
	x := subObj.Dsc.Next

	for x != nil && x.Class == Par {
		n++
		x = x.Next
	}

	return n
}
func typeName(t *TypeDesc) string {
    if t == nil {
        return "unknown"
    }

	switch t.Form {
		case Int:
			return "integer"
		case Str:
			return "string"
		default:
			return "unknown"
    }
}

func checkArgTypes(subObj *ObjDesc, args []*Item, line, col int) {
    if subObj.Dsc == nil {
        return
    }

    param := subObj.Dsc.Next
    
	for i, arg := range args {
        if param == nil || param.Class != Par {
            break
        }

        if !typeMatch(param.Type, arg.Type) {
			scn.mark(line, col, fmt.Sprintf("arg %d of %s: expected %s, got %s", i + 1, subObj.Name, typeName(param.Type), typeName(arg.Type)))
        }

        param = param.Next
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
		} else if obj.Class == Func {
			var args []*Item

			if tok.Sym == symLParen {
				args = evalArgsIntoRegs()
			}

			expected := countParams(obj)

			if len(args) != expected {
				scn.mark(tok.Line, tok.Col, fmt.Sprintf("wrong arg count for %s: want %d, got %d", name, expected, len(args)))
			}

			emit("    call " + name)

			x.Mode = Reg
			x.Type = obj.Type
		} else if obj.Class == Proc {
			scn.mark(tok.Line, tok.Col, name+" is a procedure, not a value")

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

			if tok.Sym == symAssign {
				next()

				var x Item

				if obj != nil {
					makeItem(&x, obj)
				} else {
					makeConstItem(&x, noType, 0)
				}

				var y Item

				factor(&y)

				if tok.Sym == symPlus || tok.Sym == symMinus {
					op := tok.Sym

					next()

					if y.Mode == Reg {
						emit("    movq %rax, %rbx")

						y.Mode = SavedRBX
					}

					var z Item

					factor(&z)

					addOp(op, &y, &z)
				}

				if obj != nil {
					if !typeMatch(obj.Type, y.Type) {
						scn.mark(tok.Line, tok.Col,	fmt.Sprintf("type mismatch: cannot assign %s to %s", typeName(y.Type), typeName(obj.Type)))
					}

					store(&x, &y)

					if currentFuncRetName != "" && name == currentFuncRetName {
						returnAssigned = true
					}
				}
			} else {
				if obj != nil && (obj.Class == Proc || obj.Class == Func) {
					var args []*Item

					if tok.Sym == symLParen {
						args = evalArgsIntoRegs()
					}

					expected := countParams(obj)

					if len(args) != expected {
						scn.mark(tok.Line, tok.Col,	fmt.Sprintf("wrong arg count for %s: want %d, got %d", name, expected, len(args)))
					}

					emit("    call " + name)
				} else if obj == nil {
					if tok.Sym == symLParen {
						evalArgsIntoRegs()
					}
				} else {
					scn.mark(tok.Line, tok.Col, "':=' expected")
				}
			}

		} else if tok.Sym == symKwWrite {
			next()

			check(symLParen, "'(' expected")

			var wx Item

			factor(&wx)

			if tok.Sym == symPlus || tok.Sym == symMinus {
				op := tok.Sym

				next()

				if wx.Mode == Reg {
					emit("    movq %rax, %rbx")

					wx.Mode = SavedRBX
				}

				var wz Item

				factor(&wz)

				addOp(op, &wx, &wz)
			}

			writeCall(&wx)

			check(symRParen, "')' expected")
		} else {
			scn.mark(tok.Line, tok.Col, "statement expected")

			next()
		}

		checkRegs()

		check(symSemicolon, "';' expected")
	}
}

func procedureDecl() {
	check(symKwProcedure, "'procedure' expected")

	if tok.Sym != symIdent {
		scn.mark(tok.Line, tok.Col, "procedure name expected")
		return
	}
	
	procName := tok.Lexeme

	next()

	procObj, isDup := newObj(procName, Proc)

	if isDup {
		scn.mark(tok.Line, tok.Col, "duplicate procedure: "+procName)
	}

	openScope()

	level = 1

	var params []*ObjDesc

	if tok.Sym == symLParen {
		params = formalParams()
	}

	check(symSemicolon, "';' expected")

	startOffset := int64(-8 * len(params))
	endOffset := parseLocalVarDefs(startOffset)

	frameSize := (-endOffset + 15) &^ 15
	procObj.Val = frameSize

	inSubroutine = true

	emitProcEntry(procName, frameSize, params)

	check(symKwBegin, "'begin' expected")

	statSequence()

	check(symKwEnd, "'end' expected")

	if tok.Sym == symIdent {
		if tok.Lexeme != procName {
			scn.mark(tok.Line, tok.Col, "expected 'end "+procName+"', got 'end "+tok.Lexeme+"'")
		}

		next()
	} else {
		scn.mark(tok.Line, tok.Col, "procedure name expected after 'end'")
	}

	check(symSemicolon, "';' expected")

	emitProcExit()

	inSubroutine = false

	procObj.Dsc = topScope
	
	closeScope()

	level = 0
}

func functionDecl() {
	check(symKwFunction, "'function' expected")

	if tok.Sym != symIdent {
		scn.mark(tok.Line, tok.Col, "function name expected")
		return
	}

	funcName := tok.Lexeme

	next()

	funcObj, isDup := newObj(funcName, Func)

	if isDup {
		scn.mark(tok.Line, tok.Col, "duplicate function: "+funcName)
	}

	openScope()

	level = 1

	var params []*ObjDesc
	if tok.Sym == symLParen {
		params = formalParams()
	}

	check(symColon, "':' expected (return type)")

	retType := type_()
	funcObj.Type = retType

	check(symSemicolon, "';' expected")

	startOffset := int64(-8 * len(params))
	endOffset := parseLocalVarDefs(startOffset)

	retOffset := endOffset - 8
	retObj, _ := newObj(funcName, Var)
	retObj.Type = retType
	retObj.Lev = level
	retObj.Val = retOffset

	frameSize := (-retOffset + 15) &^ 15
	funcObj.Val = frameSize

	inSubroutine = true

	emitProcEntry(funcName, frameSize, params)

	prevFuncRetName := currentFuncRetName
	prevReturnAssigned := returnAssigned
	currentFuncRetName = funcName
	returnAssigned = false

	check(symKwBegin, "'begin' expected")

	statSequence()

	check(symKwEnd, "'end' expected")

	if !returnAssigned {
		scn.mark(tok.Line, tok.Col, "function '"+funcName+"' has no return value assignment")
	}

	currentFuncRetName = prevFuncRetName
	returnAssigned = prevReturnAssigned

	if tok.Sym == symIdent {
		if tok.Lexeme != funcName {
			scn.mark(tok.Line, tok.Col, "expected 'end "+funcName+"', got 'end "+tok.Lexeme+"'")
		}

		next()
	} else {
		scn.mark(tok.Line, tok.Col, "function name expected after 'end'")
	}

	check(symSemicolon, "';' expected")

	emitFuncExit(retOffset)

	inSubroutine = false

	funcObj.Dsc = topScope

	closeScope()

	level = 0
}

func subroutineDecl() {
	if tok.Sym == symKwProcedure {
		procedureDecl()
	} else if tok.Sym == symKwFunction {
		functionDecl()
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

	for tok.Sym == symKwProcedure || tok.Sym == symKwFunction {
		subroutineDecl()
	}

	check(symKwBegin, "'begin' expected")

	statSequence()

    check(symKwEnd, "'end' expected")

    if tok.Sym != symPeriod {
        scn.mark(tok.Line, tok.Col, "'.' expected")
    }

    progScope = topScope

    closeScope()
}

// https://people.inf.ethz.ch/wirth/ProjectOberon/Sources/ORP.Mod.txt