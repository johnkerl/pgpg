package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type EBNFParser struct {
	lexer      *lexers.LookaheadLexer
	sourceName string
}

const (
	EBNFParserNodeTypeGrammar    asts.NodeType = "grammar"
	EBNFParserNodeTypeRule       asts.NodeType = "rule"
	EBNFParserNodeTypeAlternates asts.NodeType = "alternates"
	EBNFParserNodeTypeSequence   asts.NodeType = "sequence"
	EBNFParserNodeTypeOptional   asts.NodeType = "optional"
	EBNFParserNodeTypeRepeat     asts.NodeType = "repeat"
	EBNFParserNodeTypeIdentifier asts.NodeType = "identifier"
	EBNFParserNodeTypeLiteral    asts.NodeType = "literal"
)

func NewEBNFParser() AbstractParser {
	return &EBNFParser{}
}

func NewEBNFParserWithSourceName(sourceName string) AbstractParser {
	return &EBNFParser{sourceName: sourceName}
}

func (parser *EBNFParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewEBNFLexer(inputText))

	rootNode, err := parser.parseGrammar()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *EBNFParser) parseGrammar() (*asts.ASTNode, error) {
	var rules []*asts.ASTNode
	for {
		if parser.lexer.LookAhead().Type == tokens.TokenTypeEOF {
			break
		}
		rule, err := parser.parseRule()
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	if len(rules) == 0 {
		return nil, errors.New("syntax error: expected one or more rules")
	}
	return asts.NewASTNode(nil, EBNFParserNodeTypeGrammar, rules), nil
}

func (parser *EBNFParser) parseRule() (*asts.ASTNode, error) {
	accepted, nameToken, err := parser.accept(lexers.EBNFLexerTypeIdentifier)
	if err != nil {
		return nil, err
	}
	if !accepted {
		lookaheadToken := parser.lexer.LookAhead()
		return nil, fmt.Errorf(
			"syntax error: expected rule name at \"%s\"",
			parser.formatTokenLocation(lookaheadToken),
		)
	}

	if err := parser.expect(lexers.EBNFLexerTypeAssign); err != nil {
		return nil, err
	}

	expr, err := parser.parseExpression()
	if err != nil {
		return nil, err
	}

	// Semicolons are optional terminators.
	if _, _, err := parser.accept(lexers.EBNFLexerTypeSemicolon); err != nil {
		return nil, err
	}

	nameNode := asts.NewASTNode(nameToken, EBNFParserNodeTypeIdentifier, nil)
	return asts.NewASTNode(nil, EBNFParserNodeTypeRule, []*asts.ASTNode{nameNode, expr}), nil
}

func (parser *EBNFParser) parseExpression() (*asts.ASTNode, error) {
	// Expression : Sequence ( '|' Sequence )* ;
	left, err := parser.parseSequence()
	if err != nil {
		return nil, err
	}

	var alternates []*asts.ASTNode
	alternates = append(alternates, left)
	for {
		accepted, _, err := parser.accept(lexers.EBNFLexerTypeOr)
		if err != nil {
			return nil, err
		}
		if !accepted {
			break
		}
		right, err := parser.parseSequence()
		if err != nil {
			return nil, err
		}
		alternates = append(alternates, right)
	}

	if len(alternates) == 1 {
		return alternates[0], nil
	}
	return asts.NewASTNode(nil, EBNFParserNodeTypeAlternates, alternates), nil
}

func (parser *EBNFParser) parseSequence() (*asts.ASTNode, error) {
	// Sequence : Term+ ;
	var terms []*asts.ASTNode
	for {
		term, ok, err := parser.parseTermIfPresent()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		terms = append(terms, term)
	}
	if len(terms) == 0 {
		return nil, errors.New("syntax error: expected term")
	}
	if len(terms) == 1 {
		return terms[0], nil
	}
	return asts.NewASTNode(nil, EBNFParserNodeTypeSequence, terms), nil
}

func (parser *EBNFParser) parseTermIfPresent() (*asts.ASTNode, bool, error) {
	accepted, token, err := parser.accept(lexers.EBNFLexerTypeIdentifier)
	if err != nil {
		return nil, false, err
	}
	if accepted {
		return asts.NewASTNode(token, EBNFParserNodeTypeIdentifier, nil), true, nil
	}

	accepted, token, err = parser.accept(lexers.EBNFLexerTypeString)
	if err != nil {
		return nil, false, err
	}
	if accepted {
		return asts.NewASTNode(token, EBNFParserNodeTypeLiteral, nil), true, nil
	}

	accepted, _, err = parser.accept(lexers.EBNFLexerTypeLParen)
	if err != nil {
		return nil, false, err
	}
	if accepted {
		expr, err := parser.parseExpression()
		if err != nil {
			return nil, false, err
		}
		if err := parser.expect(lexers.EBNFLexerTypeRParen); err != nil {
			return nil, false, err
		}
		return expr, true, nil
	}

	accepted, _, err = parser.accept(lexers.EBNFLexerTypeLBracket)
	if err != nil {
		return nil, false, err
	}
	if accepted {
		expr, err := parser.parseExpression()
		if err != nil {
			return nil, false, err
		}
		if err := parser.expect(lexers.EBNFLexerTypeRBracket); err != nil {
			return nil, false, err
		}
		return asts.NewASTNode(nil, EBNFParserNodeTypeOptional, []*asts.ASTNode{expr}), true, nil
	}

	accepted, _, err = parser.accept(lexers.EBNFLexerTypeLBrace)
	if err != nil {
		return nil, false, err
	}
	if accepted {
		expr, err := parser.parseExpression()
		if err != nil {
			return nil, false, err
		}
		if err := parser.expect(lexers.EBNFLexerTypeRBrace); err != nil {
			return nil, false, err
		}
		return asts.NewASTNode(nil, EBNFParserNodeTypeRepeat, []*asts.ASTNode{expr}), true, nil
	}

	return nil, false, nil
}

func (parser *EBNFParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.Type == tokenType {
		// The current token is of the expected type, but there may or may not have been a lex error
		// getting the next token
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	// The current token is not of the expected type
	return false, nil, nil
}

func (parser *EBNFParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		// Lex error getting the next token
		return lexerr
	}
	if !accepted {
		// No lex error getting the next token, but the current
		// token isn't of the expected type
		lookaheadToken := parser.lexer.LookAhead()
		return fmt.Errorf(
			"expect: expected %s; got %s (%q) at %s",
			tokenType,
			lookaheadToken.Type,
			string(lookaheadToken.Lexeme),
			parser.formatTokenLocation(lookaheadToken),
		)
	}
	return nil
}

func (parser *EBNFParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}

func (parser *EBNFParser) formatTokenLocation(token *tokens.Token) string {
	if token == nil {
		return "unknown location"
	}
	if parser.sourceName != "" {
		return fmt.Sprintf(
			"%s, line %d, position %d",
			parser.sourceName,
			token.Location.LineNumber,
			token.Location.ColumnNumber,
		)
	}
	return fmt.Sprintf(
		"line %d, column %d",
		token.Location.LineNumber,
		token.Location.ColumnNumber,
	)
}
