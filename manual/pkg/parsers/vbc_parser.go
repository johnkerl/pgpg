package parsers

import (
	"errors"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

type VBCParser struct {
	lexer *lexers.LookaheadLexer
}

const (
	VBCParserNodeTypeIdentifier asts.NodeType = "identifier"
	VBCParserNodeTypeOperator   asts.NodeType = "operator"
)

func NewVBCParser() AbstractParser {
	return &VBCParser{}
}

func (parser *VBCParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewVBCLexer(inputText))

	rootNode, err := parser.parseOr()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *VBCParser) parseOr() (*asts.ASTNode, error) {
	// Or : And RestOfOr ;
	left, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfOr(left)
}

func (parser *VBCParser) parseRestOfOr(left *asts.ASTNode) (*asts.ASTNode, error) {
	// RestOfOr
	//   : or And RestOfOr
	//   | empty
	// ;
	accepted, opToken, err := parser.accept(lexers.VBCLexerTypeOr)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil // empty production rule
	}

	right, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VBCParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfOr(parent)
}

func (parser *VBCParser) parseAnd() (*asts.ASTNode, error) {
	// And : Unary RestOfAnd ;
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfAnd(left)
}

// parseRestOfAnd implements the following production rule:
//
//	RestOfAnd
//	  : and Unary RestOfAnd
//	  | empty
//	;
func (parser *VBCParser) parseRestOfAnd(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.VBCLexerTypeAnd)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil // empty production rule
	}

	right, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VBCParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfAnd(parent)
}

func (parser *VBCParser) parseUnary() (*asts.ASTNode, error) {
	// Unary
	//   : not Unary
	//   | Primary
	// ;
	accepted, opToken, err := parser.accept(lexers.VBCLexerTypeNot)
	if err != nil {
		return nil, err
	}
	if accepted {
		child, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}
		return asts.NewASTNode(opToken, VBCParserNodeTypeOperator, []*asts.ASTNode{child}), nil
	}

	return parser.parsePrimary()
}

func (parser *VBCParser) parsePrimary() (*asts.ASTNode, error) {
	accepted, token, err := parser.accept(lexers.VBCLexerTypeIdentifier)
	if err != nil {
		return nil, err
	}
	if accepted {
		return asts.NewASTNode(token, VBCParserNodeTypeIdentifier, nil), nil
	}

	accepted, _, err = parser.accept(lexers.VBCLexerTypeLParen)
	if err != nil {
		return nil, err
	}
	if accepted {
		expr, err := parser.parseOr()
		if err != nil {
			return nil, err
		}
		if err := parser.expect(lexers.VBCLexerTypeRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}

	return nil, errors.New("syntax error: expected identifier or '('")
}

func (parser *VBCParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
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

func (parser *VBCParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		// Lex error getting the next token
		return lexerr
	}
	if !accepted {
		// No lex error getting the next token, but the current
		// token isn't of the expected type
		return errors.New("expect: unexpected symbol") // TODO: describe it: expected & actual type and lexeme
	}
	return nil
}

func (parser *VBCParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}
