package grammar

import (
	"reflect"
	"testing"
)

func TestParseBNFArithmetic(t *testing.T) {
	input := []byte(`
# simple arithmetic grammar
expr ::= term | expr '+' term
term ::= factor | term '*' factor
factor ::= 'NUMBER' | '(' expr ')'
`)

	got, err := ParseBNF(input)
	if err != nil {
		t.Fatalf("ParseBNF error: %v", err)
	}

	want := &Grammar{
		Start: Symbol{Name: "expr", Kind: Nonterminal},
		Rules: []Rule{
			{LHS: Symbol{Name: "expr", Kind: Nonterminal}, RHS: []Symbol{{Name: "term", Kind: Nonterminal}}},
			{LHS: Symbol{Name: "expr", Kind: Nonterminal}, RHS: []Symbol{{Name: "expr", Kind: Nonterminal}, {Name: "+", Kind: Terminal}, {Name: "term", Kind: Nonterminal}}},
			{LHS: Symbol{Name: "term", Kind: Nonterminal}, RHS: []Symbol{{Name: "factor", Kind: Nonterminal}}},
			{LHS: Symbol{Name: "term", Kind: Nonterminal}, RHS: []Symbol{{Name: "term", Kind: Nonterminal}, {Name: "*", Kind: Terminal}, {Name: "factor", Kind: Nonterminal}}},
			{LHS: Symbol{Name: "factor", Kind: Nonterminal}, RHS: []Symbol{{Name: "NUMBER", Kind: Terminal}}},
			{LHS: Symbol{Name: "factor", Kind: Nonterminal}, RHS: []Symbol{{Name: "(", Kind: Terminal}, {Name: "expr", Kind: Nonterminal}, {Name: ")", Kind: Terminal}}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseBNF mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
