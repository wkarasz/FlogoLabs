// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/expression/gocc/token"
)

const (
	NoState    = -1
	NumStates  = 75
	NumSymbols = 80
)

type Lexer struct {
	src    []byte
	pos    int
	line   int
	column int
}

func NewLexer(src []byte) *Lexer {
	lexer := &Lexer{
		src:    src,
		pos:    0,
		line:   1,
		column: 1,
	}
	return lexer
}

func NewLexerFile(fpath string) (*Lexer, error) {
	src, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return NewLexer(src), nil
}

func (l *Lexer) Scan() (tok *token.Token) {
	tok = new(token.Token)
	if l.pos >= len(l.src) {
		tok.Type = token.EOF
		tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = l.pos, l.line, l.column
		return
	}
	start, startLine, startColumn, end := l.pos, l.line, l.column, 0
	tok.Type = token.INVALID
	state, rune1, size := 0, rune(-1), 0
	for state != -1 {
		if l.pos >= len(l.src) {
			rune1 = -1
		} else {
			rune1, size = utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}

		nextState := -1
		if rune1 != -1 {
			nextState = TransTab[state](rune1)
		}
		state = nextState

		if state != -1 {

			switch rune1 {
			case '\n':
				l.line++
				l.column = 1
			case '\r':
				l.column = 1
			case '\t':
				l.column += 4
			default:
				l.column++
			}

			switch {
			case ActTab[state].Accept != -1:
				tok.Type = ActTab[state].Accept
				end = l.pos
			case ActTab[state].Ignore != "":
				start, startLine, startColumn = l.pos, l.line, l.column
				state = 0
				if start >= len(l.src) {
					tok.Type = token.EOF
				}

			}
		} else {
			if tok.Type == token.INVALID {
				end = l.pos
			}
		}
	}
	if end > start {
		l.pos = end
		tok.Lit = l.src[start:end]
	} else {
		tok.Lit = []byte{}
	}
	tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = start, startLine, startColumn

	return
}

func (l *Lexer) Reset() {
	l.pos = 0
}

/*
Lexer symbols:
0: '"'
1: '"'
2: '''
3: '''
4: '$'
5: '|'
6: '|'
7: '&'
8: '&'
9: '('
10: ')'
11: '='
12: '='
13: '!'
14: '='
15: '<'
16: '<'
17: '='
18: '>'
19: '>'
20: '='
21: '+'
22: '-'
23: '*'
24: '/'
25: '%'
26: '('
27: ')'
28: ','
29: '?'
30: ':'
31: 't'
32: 'r'
33: 'u'
34: 'e'
35: 'f'
36: 'a'
37: 'l'
38: 's'
39: 'e'
40: 'n'
41: 'i'
42: 'l'
43: 'n'
44: 'u'
45: 'l'
46: 'l'
47: 'e'
48: 'E'
49: '+'
50: '-'
51: '.'
52: '.'
53: '.'
54: '.'
55: '_'
56: '['
57: ']'
58: '.'
59: '-'
60: '['
61: ']'
62: '_'
63: ' '
64: '$'
65: '{'
66: '}'
67: '\'
68: ' '
69: '\t'
70: '\n'
71: '\r'
72: '0'-'9'
73: 'a'-'z'
74: 'A'-'Z'
75: '0'-'9'
76: 'a'-'z'
77: 'A'-'Z'
78: '0'-'9'
79: .
*/
