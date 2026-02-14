package lexers

import (
	"github.com/johnkerl/pgpg/manual/go/pkg/tokens"
)

// LookaheadLexer wraps an AbstractLexer so that there is a LookAhead() token (one lookahead level) and
// an Advance().
type LookaheadLexer struct {
	underlying AbstractLexer
	lookToken  *tokens.Token
}

func NewLookaheadLexer(underlying AbstractLexer) *LookaheadLexer {
	lal := &LookaheadLexer{
		underlying: underlying,
	}

	// There is always at least one token, even if it's Error or EOF.
	lal.lookToken = lal.underlying.Scan()

	return lal
}

// LookAhead returns the current lookahead token without advancing the underlying scanner.  On EOF, the
// token-type will be EOF.  On error, the token-type will be Error, with Lexeme slot having the
// errortext.
func (lal *LookaheadLexer) LookAhead() (token *tokens.Token) {
	return lal.lookToken
}

// Advance scans the next token. Nominally the error will have already accepted the current token.
// It returns the current token as a convenience; the same can be gotten from calling LookAhead
// before Advance.
func (lal *LookaheadLexer) Advance() (token *tokens.Token) {
	currentToken := lal.lookToken
	lal.lookToken = lal.underlying.Scan()
	return currentToken
}
