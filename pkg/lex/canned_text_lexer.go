package lex

import (
	"errors"

	"github.com/johnkerl/pgpg/pkg/types"
	"github.com/johnkerl/pgpg/pkg/util"
)

// CannedTextLexer is primarily for unit-test purposes
type CannedTextLexer struct {
	outputs  []string
	position int

	tokenLocation *types.TokenLocation
}

func NewCannedTextLexer(text string) *CannedTextLexer {
	outputs := util.SplitString(text, " ")
	return &CannedTextLexer{
		outputs:  outputs,
		position: 0,
		tokenLocation: types.NewTokenLocation(1, 1),
	}
}

func (lxr *CannedTextLexer) Scan() (token *types.Token, err error) {
	if lxr.position >= len(lxr.outputs) {
		return nil, errors.New("input exhausted")
	}
	retval := types.NewToken(lxr.outputs[lxr.position], lxr.tokenLocation)
	lxr.position++
	lxr.tokenLocation.ColumnNumber++
	return retval, nil
}
