package lexers

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"github.com/johnkerl/pgpg/pkg/util"
)

const (
	CannedTextLexerTypeWord = 1
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

func (lexer *CannedTextLexer) DecodeType(tokenType tokens.TokenType) (string, error) {
	switch tokenType {
	case tokens.TokenTypeEOF:
		return "EOF", nil
	case tokens.TokenTypeError:
		return "error", nil
	case CannedTextLexerTypeWord:
		return "word", nil
	default:
		return "", fmt.Errorf("unrecognized token type %d", int(tokenType))
	}
}
