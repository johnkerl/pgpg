package tokens

// TokenLocation contains information for lexer state, namely the ByteOffset, as well
// as user-facing information in the form of LineNumber and ColumnNumber.
// A string FileName is not included -- I feel like this is too bulky to keep this for every single
// token from a given input file. The filename information should be tracked up a level.
type TokenLocation struct {
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

// LocateRune updates line/column number information for an accepted rune.
// This is something all lexers need to do, so it's exposed here for re-use.
func (loc *TokenLocation) LocateRune(r rune, runeWidth int) {
	if r == '\n' {
		loc.LineNumber++
		loc.ColumnNumber = 1
	} else {
		loc.ColumnNumber += runeWidth
	}
	loc.ByteOffset += runeWidth
}
