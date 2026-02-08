package lexgen

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/parsers"
)

// Tables captures DFA transitions and accepting actions for a lexer.
type Tables struct {
	StartState  int                    `json:"start_state"`
	Transitions map[int]map[string]int `json:"transitions"`
	Actions     map[int]string         `json:"actions"`
	Rules       map[string][]string    `json:"rules"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// GenerateTablesFromEBNF parses an EBNF grammar and produces lexer tables.
// Rules must expand to one or more literal strings; repeats and references are unsupported.
func GenerateTablesFromEBNF(inputText string) (*Tables, error) {
	return GenerateTablesFromEBNFWithSourceName(inputText, "")
}

func GenerateTablesFromEBNFWithSourceName(inputText string, sourceName string) (*Tables, error) {
	parser := parsers.NewEBNFParserWithSourceName(sourceName)
	ast, err := parser.Parse(inputText)
	if err != nil {
		return nil, err
	}

	rules, err := extractLiteralRules(ast)
	if err != nil {
		return nil, err
	}

	transitions := map[int]map[string]int{}
	actions := map[int]string{}
	nextState := 1

	for tokenType, literals := range rules {
		for _, literal := range literals {
			if literal == "" {
				return nil, fmt.Errorf("rule %q expands to empty literal", tokenType)
			}
			state := 0
			for _, r := range []rune(literal) {
				if transitions[state] == nil {
					transitions[state] = map[string]int{}
				}
				key := string(r)
				if next, ok := transitions[state][key]; ok {
					state = next
					continue
				}
				transitions[state][key] = nextState
				state = nextState
				nextState++
			}
			if existing, ok := actions[state]; ok && existing != tokenType {
				return nil, fmt.Errorf("conflicting actions for state %d: %q vs %q", state, existing, tokenType)
			}
			actions[state] = tokenType
		}
	}

	return &Tables{
		StartState:  0,
		Transitions: transitions,
		Actions:     actions,
		Rules:       rules,
	}, nil
}

func extractLiteralRules(ast *asts.AST) (map[string][]string, error) {
	if ast == nil || ast.RootNode == nil {
		return nil, fmt.Errorf("nil AST")
	}
	if ast.RootNode.Type != parsers.EBNFParserNodeTypeGrammar {
		return nil, fmt.Errorf("expected grammar root, got %q", ast.RootNode.Type)
	}
	rules := map[string][]string{}
	for _, ruleNode := range ast.RootNode.Children {
		if ruleNode.Type != parsers.EBNFParserNodeTypeRule {
			return nil, fmt.Errorf("expected rule node, got %q", ruleNode.Type)
		}
		if err := ruleNode.CheckArity(2); err != nil {
			return nil, err
		}
		nameNode := ruleNode.Children[0]
		exprNode := ruleNode.Children[1]
		if nameNode.Type != parsers.EBNFParserNodeTypeIdentifier || nameNode.Token == nil {
			return nil, fmt.Errorf("rule name must be identifier")
		}
		ruleName := string(nameNode.Token.Lexeme)
		literals, err := expandLiterals(exprNode)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", ruleName, err)
		}
		if len(literals) == 0 {
			return nil, fmt.Errorf("rule %q has no literals", ruleName)
		}
		rules[ruleName] = literals
	}
	return rules, nil
}

func expandLiterals(node *asts.ASTNode) ([]string, error) {
	switch node.Type {
	case parsers.EBNFParserNodeTypeLiteral:
		if node.Token == nil {
			return nil, fmt.Errorf("literal node missing token")
		}
		text := string(node.Token.Lexeme)
		unquoted, err := strconv.Unquote(text)
		if err != nil {
			return nil, fmt.Errorf("invalid literal %q: %w", text, err)
		}
		return []string{unquoted}, nil
	case parsers.EBNFParserNodeTypeSequence:
		if len(node.Children) == 0 {
			return []string{""}, nil
		}
		acc := []string{""}
		for _, child := range node.Children {
			part, err := expandLiterals(child)
			if err != nil {
				return nil, err
			}
			acc = combineLiterals(acc, part)
		}
		return acc, nil
	case parsers.EBNFParserNodeTypeAlternates:
		var out []string
		for _, child := range node.Children {
			part, err := expandLiterals(child)
			if err != nil {
				return nil, err
			}
			out = append(out, part...)
		}
		return out, nil
	case parsers.EBNFParserNodeTypeOptional:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		part, err := expandLiterals(node.Children[0])
		if err != nil {
			return nil, err
		}
		return append(part, ""), nil
	case parsers.EBNFParserNodeTypeRepeat:
		return nil, fmt.Errorf("repeat expressions are not supported for lexer literals")
	case parsers.EBNFParserNodeTypeIdentifier:
		return nil, fmt.Errorf("identifier references are not supported in lexer literals")
	default:
		return nil, fmt.Errorf("unsupported node type %q", node.Type)
	}
}

func combineLiterals(left, right []string) []string {
	out := make([]string, 0, len(left)*len(right))
	for _, l := range left {
		for _, r := range right {
			out = append(out, l+r)
		}
	}
	return out
}

// EncodeTables returns pretty-printed JSON for tables.
func EncodeTables(tables *Tables) ([]byte, error) {
	return json.MarshalIndent(tables, "", "  ")
}
