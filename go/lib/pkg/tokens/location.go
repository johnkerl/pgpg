package tokens

// NewTokenLocation is the normal use-case for a lexer starting at the beginning of input text.
func NewTokenLocation() *TokenLocation {
	return &TokenLocation{
		LineNumber:   1,
		ColumnNumber: 1,
		ByteOffset:   0,
	}
}

// NewNonDefaultTokenLocation is intended for unit-test scenarios.
func NewNonDefaultTokenLocation(lineNumber int, columnNumber int) *TokenLocation {
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
