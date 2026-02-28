package parsegen

import (
	"bytes"
	"testing"
)

func TestGenerateTablesFromReader(t *testing.T) {
	grammar := `!ws ::= " " ; num ::= "0" | "1" ; Root ::= num ;`
	r := bytes.NewBufferString(grammar)
	tables, err := GenerateTablesFromReader(r, &ParseTableOptions{SourceName: "test.bnf"})
	if err != nil {
		t.Fatalf("GenerateTablesFromReader: %v", err)
	}
	if tables == nil {
		t.Fatal("tables is nil")
	}
	if tables.StartSymbol != "Root" {
		t.Errorf("StartSymbol: got %q", tables.StartSymbol)
	}
	if len(tables.Productions) == 0 {
		t.Error("expected at least one production")
	}
}

// Minimal JSON-like grammar: Json -> Value, Value -> Object, Object -> lcurly rcurly.
// Multi-object fix should add reduce(Value->Object) for lcurly in states that have [Value -> Object .].
const jsonLikeBNF = `
!ws ::= " " ;
lcurly ::= "{" ;
rcurly ::= "}" ;
Json  ::= Value ;
Value ::= Object ;
Object ::= lcurly rcurly ;
`

func TestGenerateTablesMultiObjectAddsLcurlyReduce(t *testing.T) {
	r := bytes.NewBufferString(jsonLikeBNF)
	tables, err := GenerateTablesFromReader(r, &ParseTableOptions{SourceName: "test.bnf"})
	if err != nil {
		t.Fatalf("GenerateTablesFromReader: %v", err)
	}
	var found bool
	for _, stateActions := range tables.Actions {
		if act, ok := stateActions["lcurly"]; ok && act.Type == "reduce" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected some state to have lcurly: reduce (multi-object fix); got none")
	}
}

// JSON-like grammar with Value ::= Object | Array so multi-object must add reduce for both lcurly and lbracket.
// Before the fix, states after [] had only lbracket reduce (so [] {} failed); after {} only lcurly (so {} [] failed).
const jsonLikeMixedBNF = `
!ws ::= " " ;
lcurly  ::= "{" ;
rcurly  ::= "}" ;
lbracket ::= "[" ;
rbracket ::= "]" ;
Json   ::= Value ;
Value  ::= Object | Array ;
Object ::= lcurly rcurly ;
Array  ::= lbracket rbracket ;
`

func TestGenerateTablesMultiObjectMixedTypeReduceTerminals(t *testing.T) {
	r := bytes.NewBufferString(jsonLikeMixedBNF)
	tables, err := GenerateTablesFromReader(r, &ParseTableOptions{SourceName: "test.bnf"})
	if err != nil {
		t.Fatalf("GenerateTablesFromReader: %v", err)
	}
	var stateWithLcurlyReduce, stateWithLbracketReduce int
	foundLcurly := false
	foundLbracket := false
	for id, stateActions := range tables.Actions {
		if act, ok := stateActions["lcurly"]; ok && act.Type == "reduce" {
			stateWithLcurlyReduce = id
			foundLcurly = true
		}
		if act, ok := stateActions["lbracket"]; ok && act.Type == "reduce" {
			stateWithLbracketReduce = id
			foundLbracket = true
		}
	}
	if !foundLcurly {
		t.Error("expected some state to have lcurly: reduce (multi-object mixed-type); got none")
	}
	if !foundLbracket {
		t.Error("expected some state to have lbracket: reduce (multi-object mixed-type); got none")
	}
	// After fix: every state that can reduce to Value gets reduce for all of First(Value), so at least one
	// state should have both (the state we reach after reducing Value -> Object or Value -> Array).
	var hasBoth bool
	for _, stateActions := range tables.Actions {
		lc, okLcurly := stateActions["lcurly"]
		lb, okLbracket := stateActions["lbracket"]
		if okLcurly && okLbracket && lc.Type == "reduce" && lb.Type == "reduce" {
			hasBoth = true
			break
		}
	}
	if !hasBoth {
		t.Errorf("multi-object mixed-type fix: expected at least one state to have both lcurly: reduce and lbracket: reduce (stateWithLcurly=%d stateWithLbracket=%d)", stateWithLcurlyReduce, stateWithLbracketReduce)
	}
}
