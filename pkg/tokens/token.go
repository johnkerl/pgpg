package tokens

// Token tracks a single lexeme as well as where it was found within the source text.
type Token struct {
	Lexeme []rune
	// TODO: type-inference -- need to figure out the API

	// TODO: we want internal types like TokenTypeError and TokenTypeEOF, but also
	// external types like whatever the caller has from their grammar.

	// Maybe have the internal types have codes -1 and -2, say, and user-defined types
	// all be non-negative -- ?
	Location TokenLocation
}

func NewToken(lexeme []rune, location *TokenLocation) *Token {
	return &Token{
		Lexeme:   lexeme,
		Location: *location, // does a copy
	}
}

type TokenLocation struct {
	// FileName string -- too bulky to keep this for every single token -- store that up a level -- ?
	LineNumber   int
	ColumnNumber int
	ByteOffset   int
}

// NewDefaultTokenLocation is the normal use-case for a lexer starting at the beginning of input text.
func NewDefaultTokenLocation() *TokenLocation {
	return &TokenLocation{
		LineNumber:   1,
		ColumnNumber: 1,
		ByteOffset:   0,
	}
}

// NewTokenLocation is intended for unit-test scenarios.
func NewTokenLocation(lineNumber int, columnNumber int) *TokenLocation {
	return &TokenLocation{
		LineNumber:   lineNumber,
		ColumnNumber: columnNumber,
		ByteOffset:   0,
	}
}

// locateRune updates line/column number information for an accepted rune.
func (loc *TokenLocation) LocateRune(r rune, runeWidth int) {
	if r == '\n' {
		loc.LineNumber++
		loc.ColumnNumber = 1
	} else {
		loc.ColumnNumber += runeWidth
	}
	loc.ByteOffset += runeWidth
}
