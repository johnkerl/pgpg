package parsers

import (
	"errors"

	"github.com/johnkerl/pgpg/apps/go/manual/lexers"
	"github.com/johnkerl/pgpg/lib/go/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
	"github.com/johnkerl/pgpg/lib/go/pkg/parsers"
	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
)

// Root : Sum ; Sum : Product RestOfSum ; RestOfSum : plus Product RestOfSum | empty ;
// Product : int_literal RestOfProduct ; RestOfProduct : times int_literal RestOfProduct | empty ;

type AMNEParser struct {
	lexer *liblexers.LookaheadLexer
}

const (
	AMNEParserNodeTypeNumber   asts.NodeType = "number"
	AMNEParserNodeTypeOperator asts.NodeType = "operator"
)

func NewAMNEParser() parsers.AbstractParser {
	return &AMNEParser{}
}

func (parser *AMNEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = liblexers.NewLookaheadLexer(lexers.NewAMLexer(inputText))

	rootNode, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *AMNEParser) parseSum() (*asts.ASTNode, error) {
	left, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfSum(left)
}

func (parser *AMNEParser) parseRestOfSum(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.AMLexerTypePlus)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, AMNEParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfSum(parent)
}

func (parser *AMNEParser) parseProduct() (*asts.ASTNode, error) {
	left, err := parser.parseIntLiteral()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfProduct(left)
}

func (parser *AMNEParser) parseRestOfProduct(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.AMLexerTypeTimes)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}

	right, err := parser.parseIntLiteral()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, AMNEParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfProduct(parent)
}

func (parser *AMNEParser) parseIntLiteral() (*asts.ASTNode, error) {
	accepted, token, err := parser.accept(lexers.AMLexerTypeNumber)
	if accepted && err == nil {
		return asts.NewASTNode(token, AMNEParserNodeTypeNumber, nil), nil
	}
	return nil, errors.New("syntax error: expected int literal; got " + token.String())
}

func (parser *AMNEParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.Type == tokenType {
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	return false, nil, nil
}

func (parser *AMNEParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		return lexerr
	}
	if !accepted {
		return errors.New("expect: unexpected symbol")
	}
	return nil
}

func (parser *AMNEParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}
	return nil
}
