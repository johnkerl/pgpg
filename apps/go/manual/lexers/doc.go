// Package lexers provides sample (manual) lexers for trylex and tryparse.
// It depends on github.com/johnkerl/pgpg/lib for AbstractLexer, tokens, and util.
package lexers

import (
	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
)

// AbstractLexer is the interface that all lexers in this package implement.
// It is the same as lib's AbstractLexer so that trylex/tryparse can use either lib or manual lexers.
type AbstractLexer = liblexers.AbstractLexer

// RunePredicateFunc is used by some lexers for predicates like unicode.IsSpace.
type RunePredicateFunc = liblexers.RunePredicateFunc
