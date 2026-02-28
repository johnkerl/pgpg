package parsers

import (
	"errors"
	"fmt"
	"io"

	"github.com/johnkerl/pgpg/apps/go/manual/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/parsers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

type VBCParser struct {
	lexer *liblexers.LookaheadLexer
}

const (
	VBCParserNodeTypeIdentifier asts.NodeType = "identifier"
	VBCParserNodeTypeOperator   asts.NodeType = "operator"
)

func NewVBCParser() parsers.AbstractParser {
	return &VBCParser{}
}

func (parser *VBCParser) Parse(r io.Reader) (*asts.AST, error) {
	parser.lexer = liblexers.NewLookaheadLexer(lexers.NewVBCLexer(r))

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
	left, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfOr(left)
}

func (parser *VBCParser) parseRestOfOr(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.VBCLexerTypeOr)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parseAnd()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VBCParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfOr(parent)
}

func (parser *VBCParser) parseAnd() (*asts.ASTNode, error) {
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfAnd(left)
}

func (parser *VBCParser) parseRestOfAnd(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.VBCLexerTypeAnd)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VBCParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfAnd(parent)
}

func (parser *VBCParser) parseUnary() (*asts.ASTNode, error) {
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
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	return false, nil, nil
}

func (parser *VBCParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		return lexerr
	}
	if !accepted {
		lookaheadToken := parser.lexer.LookAhead()
		return fmt.Errorf(
			"expect: expected %s; got %s (%q)",
			tokenType,
			lookaheadToken.Type,
			string(lookaheadToken.Lexeme),
		)
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
