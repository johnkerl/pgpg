package tokens

// Token tracks a single lexeme as well as where it was found within the source text.
type Token struct {
	Lexeme string
	// TODO: type-inference -- need to figure out the API
	Location TokenLocation
}

func NewToken(lexeme string, location *TokenLocation) *Token {
	return &Token{
		Lexeme:   lexeme,
		Location: *location, // does a copy
	}
}

type TokenLocation struct {
	// FileName string -- too bulky to keep this for every single token -- store that up a level -- ?
	LineNumber   int
	ColumnNumber int
}

func NewTokenLocation(lineNumber int, columnNumber int) *TokenLocation {
	return &TokenLocation{
		LineNumber:   lineNumber,
		ColumnNumber: columnNumber,
	}
}
