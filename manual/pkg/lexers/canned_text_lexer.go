package lexers

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
	"github.com/johnkerl/pgpg/pkg/util"
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

func NewCannedTextLexer(text string) AbstractLexer {
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
