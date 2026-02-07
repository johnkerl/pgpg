package ir

import (
	"encoding/json"
	"io"

	"github.com/johnkerl/pgpg/v2/grammar"
)

// IR is the language-independent grammar representation.
type IR struct {
	Version string     `json:"version"`
	Grammar GrammarIR  `json:"grammar"`
	Lexer   LexerIR    `json:"lexer"`
	Parser  ParserIR   `json:"parser"`
	Meta    MetadataIR `json:"meta,omitempty"`
}

type GrammarIR struct {
	StartSymbol string   `json:"start_symbol"`
	Rules       []RuleIR `json:"rules"`
}

type RuleIR struct {
	LHS string   `json:"lhs"`
	RHS []string `json:"rhs"`
}

type LexerIR struct {
	Rules []LexerRuleIR `json:"rules"`
}

type LexerRuleIR struct {
	Name  string `json:"name"`
	Regex string `json:"regex"`
}

type ParserIR struct {
	Algorithm       string             `json:"algorithm"`
	Productions     []ProductionIR     `json:"productions,omitempty"`
	Actions         []ActionEntryIR    `json:"actions,omitempty"`
	Gotos           []GotoEntryIR      `json:"gotos,omitempty"`
	SemanticActions []SemanticActionIR `json:"semantic_actions,omitempty"`
}

type ProductionIR struct {
	LHS      string `json:"lhs"`
	RHSCount int    `json:"rhs_count"`
	ActionID string `json:"action_id,omitempty"`
}

type ActionEntryIR struct {
	State    int    `json:"state"`
	Terminal string `json:"terminal"`
	Action   string `json:"action"`
	Value    int    `json:"value,omitempty"`
}

type GotoEntryIR struct {
	State       int    `json:"state"`
	Nonterminal string `json:"nonterminal"`
	Next        int    `json:"next"`
}

type SemanticActionIR struct {
	ID   string `json:"id"`
	Note string `json:"note,omitempty"`
}

type MetadataIR struct {
	Source string `json:"source,omitempty"`
}

const VersionV0 = "v0"

// FromGrammar converts a grammar AST into an IR skeleton.
func FromGrammar(g *grammar.Grammar, source string) *IR {
	if g == nil {
		return &IR{Version: VersionV0}
	}

	rules := make([]RuleIR, 0, len(g.Rules))
	for _, rule := range g.Rules {
		rhs := make([]string, 0, len(rule.RHS))
		for _, sym := range rule.RHS {
			rhs = append(rhs, sym.Name)
		}
		rules = append(rules, RuleIR{
			LHS: rule.LHS.Name,
			RHS: rhs,
		})
	}

	return &IR{
		Version: VersionV0,
		Grammar: GrammarIR{
			StartSymbol: g.Start.Name,
			Rules:       rules,
		},
		Lexer: LexerIR{},
		Parser: ParserIR{
			Productions: grammarProductions(g),
		},
		Meta: MetadataIR{
			Source: source,
		},
	}
}

func grammarProductions(g *grammar.Grammar) []ProductionIR {
	if g == nil {
		return nil
	}
	out := make([]ProductionIR, 0, len(g.Rules))
	for _, rule := range g.Rules {
		out = append(out, ProductionIR{
			LHS:      rule.LHS.Name,
			RHSCount: len(rule.RHS),
		})
	}
	return out
}

func EncodeJSON(w io.Writer, ir *IR) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(ir)
}

func DecodeJSON(r io.Reader) (*IR, error) {
	var out IR
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
