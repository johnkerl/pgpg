// Package ruledefs extracts rule definitions from a parsed EBNF grammar AST.
// It is the shared front end of the lexgen and parsegen table generators.
package ruledefs

import (
	"fmt"
	"unicode"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	"github.com/johnkerl/pgpg/go/lib/pkg/parsers"
)

// RuleDef is one grammar rule: its name and its defining expression subtree.
type RuleDef struct {
	Name string
	Expr *asts.ASTNode
}

// Extract walks a grammar AST into its list of rule definitions.
func Extract(ast *asts.AST) ([]RuleDef, error) {
	if ast == nil || ast.RootNode == nil {
		return nil, fmt.Errorf("nil AST")
	}
	if ast.RootNode.Type != parsers.EBNFParserNodeTypeGrammar {
		return nil, fmt.Errorf("expected grammar root, got %q", ast.RootNode.Type)
	}
	var rules []RuleDef
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
		rules = append(rules, RuleDef{Name: ruleName, Expr: exprNode})
	}
	return rules, nil
}

// SelectLexerRuleNames returns the names of lexer rules, in definition order.
func SelectLexerRuleNames(ruleDefs []RuleDef) []string {
	var names []string
	for _, rule := range ruleDefs {
		if IsLexerRuleName(rule.Name) {
			names = append(names, rule.Name)
		}
	}
	return names
}

// IsLexerRuleName reports whether name names a lexer rule: leading '!' (ignore
// rules), '_' (helper rules), or a lowercase letter. Parser rules start with
// an uppercase letter.
func IsLexerRuleName(name string) bool {
	if name == "" {
		return false
	}
	first := []rune(name)[0]
	if first == '!' {
		return true
	}
	return first == '_' || unicode.IsLower(first)
}
