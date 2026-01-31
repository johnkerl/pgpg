package grammar

import (
	"fmt"
)

type bnfTokenKind int

const (
	bnfTokEOF bnfTokenKind = iota
	bnfTokIdent
	bnfTokString
	bnfTokAssign // ::=
	bnfTokPipe   // |
	bnfTokSemi   // ;
	bnfTokEOL    // \n
)

type bnfToken struct {
	kind   bnfTokenKind
	lexeme string
	line   int
	col    int
}

type bnfLexer struct {
	input []byte
	pos   int
	line  int
	col   int
}

func newBNFLexer(input []byte) *bnfLexer {
	return &bnfLexer{
		input: input,
		pos:   0,
		line:  1,
		col:   1,
	}
}

func (l *bnfLexer) next() (bnfToken, error) {
	for {
		ch, ok := l.peek()
		if !ok {
			return bnfToken{kind: bnfTokEOF, line: l.line, col: l.col}, nil
		}
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.read()
			continue
		}
		if ch == '#' {
			l.read()
			for {
				next, ok := l.peek()
				if !ok || next == '\n' {
					break
				}
				l.read()
			}
			continue
		}
		break
	}

	ch, _ := l.peek()
	startLine, startCol := l.line, l.col

	switch ch {
	case '\n':
		l.read()
		return bnfToken{kind: bnfTokEOL, lexeme: "\n", line: startLine, col: startCol}, nil
	case '|':
		l.read()
		return bnfToken{kind: bnfTokPipe, lexeme: "|", line: startLine, col: startCol}, nil
	case ';':
		l.read()
		return bnfToken{kind: bnfTokSemi, lexeme: ";", line: startLine, col: startCol}, nil
	case ':':
		l.read()
		if next, ok := l.peek(); !ok || next != ':' {
			return bnfToken{}, l.errorf(startLine, startCol, "expected '::='")
		}
		l.read()
		if next, ok := l.peek(); !ok || next != '=' {
			return bnfToken{}, l.errorf(startLine, startCol, "expected '::='")
		}
		l.read()
		return bnfToken{kind: bnfTokAssign, lexeme: "::=", line: startLine, col: startCol}, nil
	case '\'':
		return l.lexString()
	default:
		if isIdentStart(ch) {
			return l.lexIdent()
		}
		return bnfToken{}, l.errorf(startLine, startCol, "unexpected character %q", ch)
	}
}

func (l *bnfLexer) lexIdent() (bnfToken, error) {
	startLine, startCol := l.line, l.col
	var buf []byte
	for {
		ch, ok := l.peek()
		if !ok || !isIdentContinue(ch) {
			break
		}
		buf = append(buf, ch)
		l.read()
	}
	return bnfToken{kind: bnfTokIdent, lexeme: string(buf), line: startLine, col: startCol}, nil
}

func (l *bnfLexer) lexString() (bnfToken, error) {
	startLine, startCol := l.line, l.col
	l.read() // opening quote
	var buf []byte
	for {
		ch, ok := l.peek()
		if !ok {
			return bnfToken{}, l.errorf(startLine, startCol, "unterminated string")
		}
		if ch == '\n' {
			return bnfToken{}, l.errorf(startLine, startCol, "unterminated string before newline")
		}
		if ch == '\'' {
			l.read()
			break
		}
		if ch == '\\' {
			l.read()
			esc, ok := l.peek()
			if !ok {
				return bnfToken{}, l.errorf(startLine, startCol, "unterminated escape")
			}
			l.read()
			switch esc {
			case '\\', '\'':
				buf = append(buf, esc)
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'r':
				buf = append(buf, '\r')
			default:
				return bnfToken{}, l.errorf(startLine, startCol, "unsupported escape \\%c", esc)
			}
			continue
		}
		buf = append(buf, ch)
		l.read()
	}
	return bnfToken{kind: bnfTokString, lexeme: string(buf), line: startLine, col: startCol}, nil
}

func (l *bnfLexer) peek() (byte, bool) {
	if l.pos >= len(l.input) {
		return 0, false
	}
	return l.input[l.pos], true
}

func (l *bnfLexer) read() {
	if l.pos >= len(l.input) {
		return
	}
	ch := l.input[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

func (l *bnfLexer) errorf(line, col int, format string, args ...any) error {
	return fmt.Errorf("bnf lex %d:%d: %s", line, col, fmt.Sprintf(format, args...))
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentContinue(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}
