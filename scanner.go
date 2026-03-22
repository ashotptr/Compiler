package main

import (
    "fmt"
    "os"
    "strings"
)

const (
    symNull = 0

    symPlus = 1
    symMinus = 2
    symTimes = 3
    symDiv = 4
    symAssign = 5
    symColon = 6
    symSemicolon = 7
    symLParen = 8
    symRParen = 9

    symString = 10
    symIdent = 11
    symNumber = 12
    symComma = 13
    symPeriod = 14
    symKwIf = 20
    symKwThen = 21
    symKwBegin = 22
    symKwEnd = 23
    symKwProgram = 24
    symKwProcedure = 25
    symKwVar = 26
    symKwWrite = 27
    symKwInteger = 28
    symKwStringType = 29

    symEOF = 99
)

var tokenNames = map[int]string{
    symPlus: "PLUS",
    symMinus: "MINUS",
    symTimes: "TIMES",
    symDiv: "DIV",
    symAssign: "ASSIGN",
    symColon: "COLON",
    symSemicolon: "SEMICOLON",
    symLParen: "LPAREN",
    symRParen: "RPAREN",
    symComma: "COMMA",
    symPeriod: "PERIOD",
    symString: "STRING",
    symNumber: "NUMBER",
    symIdent: "IDENT",
    symKwIf: "KW_IF",
    symKwThen: "KW_THEN",
    symKwBegin: "KW_BEGIN",
    symKwEnd: "KW_END",
    symKwProgram: "KW_PROGRAM",
    symKwProcedure: "KW_PROCEDURE",
    symKwVar: "KW_VAR",
    symKwWrite: "KW_WRITE",
    symKwInteger: "KW_INTEGER",
    symKwStringType: "KW_STRING",
    symEOF: "EOF",
}

const NKW = 10

type kwEntry struct {
    sym int
    id string
}

var keyTab [NKW]kwEntry
var KWX [10]int

func enterKW(sym int, name string, k *int) {
    keyTab[*k] = kwEntry{sym: sym, id: name}

    (*k)++
}

func init() {
    k := 0
    KWX[0], KWX[1] = 0, 0

    enterKW(symKwIf, "if", &k)
    KWX[2] = k

    enterKW(symKwEnd, "end", &k)
    enterKW(symKwVar, "var", &k)
    KWX[3] = k

    enterKW(symKwThen, "then", &k)
    KWX[4] = k

    enterKW(symKwBegin, "begin", &k)
    enterKW(symKwWrite, "write", &k)
    KWX[5] = k

    enterKW(symKwStringType, "string", &k)
    KWX[6] = k

    enterKW(symKwInteger, "integer", &k)
    enterKW(symKwProgram, "program", &k)
    KWX[7] = k

    KWX[8] = k

    enterKW(symKwProcedure, "procedure", &k)
    KWX[9] = k
}

type Token struct {
    Sym int
    Lexeme string
    Line int
    Col int
}

type Scanner struct {
    src []byte
    pos int
    ch byte
    line int
    col int
    id string
    str string
    ival int64
    errcnt int
}

func initScanner(src []byte) *Scanner {
    s := &Scanner{src: src, line: 1, col: 0}

    s.readChar()

    return s
}

func (s *Scanner) readChar() {
	if s.pos < len(s.src) {
		s.ch = s.src[s.pos]
		s.pos++

		if s.ch == '\n' {
			s.line++
			s.col = 0
		} else {
			s.col++
		}
	} else {
		if s.ch != 0 {
			s.col++
		}

		s.ch = 0
	}
}

func (s *Scanner) eot() bool {
	return s.ch == 0
}

func (s *Scanner) mark(line, col int, msg string) {
	fmt.Fprintf(os.Stderr, "Error (line %d, col %d): %s\n", line, col, msg)

    s.errcnt++
}

func (s *Scanner) identifier() int {
	var buf strings.Builder

	for isAlphaNum(s.ch) {
		buf.WriteByte(s.ch)

		s.readChar()
	}

	s.id = buf.String()

	n := len(s.id)

	if n >= 2 && n <= 9 {
		lo, hi := KWX[n - 1], KWX[n]

		for k := lo; k < hi; k++ {
			if s.id == keyTab[k].id {
				return keyTab[k].sym
			}
		}
	}

	return symIdent
}

func (s *Scanner) scanNumber() int {
	s.ival = 0

	for s.ch >= '0' && s.ch <= '9' {
		s.ival = s.ival * 10 + int64(s.ch - '0')

		s.readChar()
	}

	return symNumber
}

func (s *Scanner) scanString(startLine, startCol int) int {
	var buf strings.Builder

	buf.WriteByte('"')

	s.readChar()

	for !s.eot() && s.ch != '"' && s.ch != '\n' {
		buf.WriteByte(s.ch)

		s.readChar()
	}

	if s.eot() || s.ch == '\n' {
		s.mark(startLine, startCol, "unterminated string")
	} else {
		buf.WriteByte('"')

		s.readChar()
	}

	s.str = buf.String()

	return symString
}

func (s *Scanner) blockComment(startLine, startCol int) {
	s.readChar()

	for !s.eot() && s.ch != '}' {
		s.readChar()
	}

	if s.eot() {
		s.mark(startLine, startCol, "unterminated comment")
	} else {
		s.readChar()
	}
}

func (s *Scanner) lineComment() {
	for !s.eot() && s.ch != '\n' {
		s.readChar()
	}
}

func (s *Scanner) get() Token {
	var sym int
	var lexeme string
	var tokLine, tokCol int

	for {
		for !s.eot() && s.ch <= ' ' {
			s.readChar()
		}

		tokLine = s.line
		tokCol = s.col
		lexeme = ""

		if s.eot() {
			return Token{symEOF, "", tokLine, tokCol}
		}

		ch := s.ch

		switch {
            case isLetter(ch) || ch == '_':
                sym = s.identifier()

                lexeme = s.id

            case ch >= '0' && ch <= '9':
                sym = s.scanNumber()
                lexeme = fmt.Sprintf("%d", s.ival)

            case ch == '"':
                sym = s.scanString(tokLine, tokCol)

                lexeme = s.str

            case ch == '{':
                s.blockComment(tokLine, tokCol)

                sym = symNull

            default:
                switch ch {
                    case ':':
                        s.readChar()

                        if s.ch == '=' {
                            s.readChar()

                            sym, lexeme = symAssign, ":="
                        } else {
                            sym, lexeme = symColon, ":"
                        }

                    case '/':
                        s.readChar()

                        if s.ch == '/' {
                            s.lineComment()

                            sym = symNull
                        } else {
                            sym, lexeme = symDiv, "/"
                        }

                    case ';':
                        s.readChar()

                        sym, lexeme = symSemicolon, ";"

                    case ',':
                        s.readChar()
                        sym, lexeme = symComma, ","

                    case '.':
                        s.readChar()
                        sym, lexeme = symPeriod, "."

                    case '(':
                        s.readChar()

                        sym, lexeme = symLParen, "("

                    case ')':
                        s.readChar()

                        sym, lexeme = symRParen, ")"

                    case '+':
                        s.readChar()

                        sym, lexeme = symPlus, "+"

                    case '-':
                        s.readChar()

                        sym, lexeme = symMinus, "-"

                    case '*':
                        s.readChar()

                        sym, lexeme = symTimes, "*"

                    default:
                        s.readChar()

                        if ch >= ' ' && ch <= '~' {
                            s.mark(tokLine, tokCol, fmt.Sprintf("invalid token '%c'", ch))
                        } else {
                            s.mark(tokLine, tokCol, fmt.Sprintf("invalid token (0x%02x)", ch))
                        }

                        sym = symNull
                    }
		}

		if sym != symNull {
			return Token{sym, lexeme, tokLine, tokCol}
		}
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isAlphaNum(ch byte) bool {
	return isLetter(ch) || (ch >= '0' && ch <= '9') || ch == '_'
}

// func main() {
//     if len(os.Args) < 2 {
//         fmt.Fprintln(os.Stderr, "Usage: ./scanner <source_file.pas>")

//         os.Exit(1)
//     }

//     data, err := os.ReadFile(os.Args[1])

//     if err != nil {
//         fmt.Fprintf(os.Stderr, "Error: cannot open file '%s': %v\n", os.Args[1], err)
        
//         os.Exit(1)
//     }

//     s := initScanner(data)

//     for {
//         tok := s.get()

//         name := tokenNames[tok.Sym]

//         if tok.Sym == symEOF {
//             fmt.Printf("%-14s (%d,%d)\n", name, tok.Line, tok.Col)

//             break
//         }

//         fmt.Printf("%-14s %-20s (%d,%d)\n", name, tok.Lexeme, tok.Line, tok.Col)
//     }
// }

// https://people.inf.ethz.ch/wirth/ProjectOberon/Sources/ORS.Mod.txt