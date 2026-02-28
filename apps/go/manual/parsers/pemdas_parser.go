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

type PEMDASParser struct {
	lexer *liblexers.LookaheadLexer
}

const (
	PEMDASParserNodeTypeNumber   asts.NodeType = "number"
	PEMDASParserNodeTypeOperator asts.NodeType = "operator"
)

func NewPEMDASParser() parsers.AbstractParser {
	return &PEMDASParser{}
}

func (parser *PEMDASParser) Parse(r io.Reader) (*asts.AST, error) {
	parser.lexer = liblexers.NewLookaheadLexer(lexers.NewPEMDASLexer(r))

	rootNode, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *PEMDASParser) parseSum() (*asts.ASTNode, error) {
	left, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfSum(left)
}

func (parser *PEMDASParser) parseRestOfSum(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.PEMDASLexerTypePlus)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.PEMDASLexerTypeMinus)
		if err != nil {
			return nil, err
		}
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, PEMDASParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfSum(parent)
}

func (parser *PEMDASParser) parseProduct() (*asts.ASTNode, error) {
	left, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfProduct(left)
}

func (parser *PEMDASParser) parseRestOfProduct(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.PEMDASLexerTypeTimes)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.PEMDASLexerTypeDivide)
		if err != nil {
			return nil, err
		}
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, PEMDASParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfProduct(parent)
}

func (parser *PEMDASParser) parsePower() (*asts.ASTNode, error) {
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfPower(left)
}

func (parser *PEMDASParser) parseRestOfPower(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.PEMDASLexerTypePower)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, PEMDASParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parent, nil
}

func (parser *PEMDASParser) parseUnary() (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.PEMDASLexerTypePlus)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.PEMDASLexerTypeMinus)
		if err != nil {
			return nil, err
		}
	}
	if accepted {
		child, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}
		return asts.NewASTNode(opToken, PEMDASParserNodeTypeOperator, []*asts.ASTNode{child}), nil
	}

	return parser.parsePrimary()
}

func (parser *PEMDASParser) parsePrimary() (*asts.ASTNode, error) {
	accepted, token, err := parser.accept(lexers.PEMDASLexerTypeNumber)
	if err != nil {
		return nil, err
	}
	if accepted {
		return asts.NewASTNode(token, PEMDASParserNodeTypeNumber, nil), nil
	}

	accepted, _, err = parser.accept(lexers.PEMDASLexerTypeLParen)
	if err != nil {
		return nil, err
	}
	if accepted {
		expr, err := parser.parseSum()
		if err != nil {
			return nil, err
		}
		if err := parser.expect(lexers.PEMDASLexerTypeRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}

	return nil, errors.New("syntax error: expected int literal or '('")
}

func (parser *PEMDASParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.Type == tokenType {
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	return false, nil, nil
}

func (parser *PEMDASParser) expect(tokenType tokens.TokenType) error {
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

func (parser *PEMDASParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}
	return nil
}
