package grammar

// SymbolKind distinguishes terminals and nonterminals.
type SymbolKind int

const (
	Terminal SymbolKind = iota
	Nonterminal
)

// Symbol represents a grammar terminal or nonterminal.
type Symbol struct {
	Name string
	Kind SymbolKind
}

// Rule is a single production rule.
type Rule struct {
	LHS Symbol
	RHS []Symbol
}

// Grammar is a collection of production rules with a start symbol.
type Grammar struct {
	Start Symbol
	Rules []Rule
}

// ParseBNF parses a BNF file into a Grammar AST.
func ParseBNF(input []byte) (*Grammar, error) {
	p, err := newBNFParser(input)
	if err != nil {
		return nil, err
	}
	return p.parseGrammar()
}
