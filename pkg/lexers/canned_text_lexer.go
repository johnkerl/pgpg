package lexers

import (
	"errors"

	"github.com/johnkerl/pgpg/pkg/tokens"
	"github.com/johnkerl/pgpg/pkg/util"
)

// CannedTextLexer is primarily for unit-test purposes
type CannedTextLexer struct {
	outputs  []string
	position int

	tokenLocation *tokens.TokenLocation
}

func NewCannedTextLexer(text string) *CannedTextLexer {
	outputs := util.SplitString(text, " ")
	return &CannedTextLexer{
		outputs:  outputs,
		position: 0,
		tokenLocation: tokens.NewTokenLocation(1, 1),
	}
}

func (lxr *CannedTextLexer) Scan() (token *tokens.Token, err error) {
	if lxr.position >= len(lxr.outputs) {
		return nil, errors.New("input exhausted")
	}
	retval := tokens.NewToken(lxr.outputs[lxr.position], lxr.tokenLocation)
	lxr.position++
	lxr.tokenLocation.ColumnNumber++
	return retval, nil
}
