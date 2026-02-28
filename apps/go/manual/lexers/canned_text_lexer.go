package lexers

import (
	"io"

	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
	"github.com/johnkerl/pgpg/go/lib/pkg/util"
)

const (
	CannedTextLexerTypeWord tokens.TokenType = "word"
)

// CannedTextLexer is primarily for unit-test purposes.
// Unlike in WordLexer, here multiple spaces aren't the same as one space.
type CannedTextLexer struct {
	outputs  []string
	position int

	tokenLocation *tokens.TokenLocation
}

func NewCannedTextLexer(r io.Reader) AbstractLexer {
	b, _ := io.ReadAll(r)
	return NewCannedTextLexerFromString(string(b))
}

func NewCannedTextLexerFromString(text string) AbstractLexer {
	outputs := util.SplitString(text, " ")
	return &CannedTextLexer{
		outputs:       outputs,
		position:      0,
		tokenLocation: tokens.NewTokenLocation(),
	}
}

func (lexer *CannedTextLexer) Scan() (token *tokens.Token) {
	if lexer.position >= len(lexer.outputs) {
		return tokens.NewEOFToken(lexer.tokenLocation)
	}
	retval := tokens.NewToken(
		[]rune(lexer.outputs[lexer.position]),
		CannedTextLexerTypeWord,
		lexer.tokenLocation,
	)
	lexer.position++
	lexer.tokenLocation.ColumnNumber++
	return retval
}
