package ir

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/johnkerl/pgpg/tmp/grammar"
)

func TestIRRoundTrip(t *testing.T) {
	g := &grammar.Grammar{
		Start: grammar.Symbol{Name: "expr", Kind: grammar.Nonterminal},
		Rules: []grammar.Rule{
			{LHS: grammar.Symbol{Name: "expr", Kind: grammar.Nonterminal}, RHS: []grammar.Symbol{{Name: "term", Kind: grammar.Nonterminal}}},
			{LHS: grammar.Symbol{Name: "term", Kind: grammar.Nonterminal}, RHS: []grammar.Symbol{{Name: "NUMBER", Kind: grammar.Terminal}}},
		},
	}

	doc := FromGrammar(g, "example.bnf")
	var buf bytes.Buffer
	if err := EncodeJSON(&buf, doc); err != nil {
		t.Fatalf("EncodeJSON error: %v", err)
	}

	decoded, err := DecodeJSON(&buf)
	if err != nil {
		t.Fatalf("DecodeJSON error: %v", err)
	}

	if !reflect.DeepEqual(doc, decoded) {
		t.Fatalf("IR round-trip mismatch\n got: %#v\nwant: %#v", decoded, doc)
	}
}
