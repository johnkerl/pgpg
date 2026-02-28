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
