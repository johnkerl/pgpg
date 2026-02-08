package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type VICParser struct {
	lexer *lexers.LookaheadLexer
}

const (
	VICParserNodeTypeNumber     asts.NodeType = "number"
	VICParserNodeTypeIdentifier asts.NodeType = "identifier"
	VICParserNodeTypeOperator   asts.NodeType = "operator"
	VICParserNodeTypeAssignment asts.NodeType = "assignment"
)

func NewVICParser() AbstractParser {
	return &VICParser{}
}

func (parser *VICParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewVICLexer(inputText))

	rootNode, err := parser.parseStatement()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *VICParser) parseStatement() (*asts.ASTNode, error) {
	left, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	accepted, assignToken, err := parser.accept(lexers.VICLexerTypeAssign)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil
	}
	if left.Type != VICParserNodeTypeIdentifier {
		return nil, errors.New("syntax error: assignment requires identifier on left-hand side")
	}

	right, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	return asts.NewASTNode(assignToken, VICParserNodeTypeAssignment, []*asts.ASTNode{left, right}), nil
}

func (parser *VICParser) parseSum() (*asts.ASTNode, error) {
	// Sum : Product RestOfSum ;
	left, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfSum(left)
}

func (parser *VICParser) parseRestOfSum(left *asts.ASTNode) (*asts.ASTNode, error) {
	// RestOfSum
	//   : plus Product RestOfSum
	//   | minus Product RestOfSum
	//   | empty
	// ;
	accepted, opToken, err := parser.accept(lexers.VICLexerTypePlus)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.VICLexerTypeMinus)
		if err != nil {
			return nil, err
		}
	}
	if !accepted {
		return left, nil // empty production rule
	}

	right, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VICParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfSum(parent)
}

func (parser *VICParser) parseProduct() (*asts.ASTNode, error) {
	// Product : Power RestOfProduct ;
	left, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfProduct(left)
}

// parseRestOfProduct implements the following production rule:
//
//	RestOfProduct
//	  : times Power RestOfProduct
//	  | divide Power RestOfProduct
//	  | empty
//	;
func (parser *VICParser) parseRestOfProduct(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.VICLexerTypeTimes)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.VICLexerTypeDivide)
		if err != nil {
			return nil, err
		}
	}
	if !accepted {
		return left, nil // empty production rule
	}

	right, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VICParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parser.parseRestOfProduct(parent)
}

func (parser *VICParser) parsePower() (*asts.ASTNode, error) {
	// Power : Unary RestOfPower ;
	left, err := parser.parseUnary()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfPower(left)
}

// parseRestOfPower implements the following production rule:
//
//	RestOfPower
//	  : power Power
//	  | empty
//	;
func (parser *VICParser) parseRestOfPower(left *asts.ASTNode) (*asts.ASTNode, error) {
	accepted, opToken, err := parser.accept(lexers.VICLexerTypePower)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return left, nil // empty production rule
	}

	right, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, VICParserNodeTypeOperator, []*asts.ASTNode{left, right})
	return parent, nil
}

func (parser *VICParser) parseUnary() (*asts.ASTNode, error) {
	// Unary
	//   : plus Unary
	//   | minus Unary
	//   | Primary
	// ;
	accepted, opToken, err := parser.accept(lexers.VICLexerTypePlus)
	if err != nil {
		return nil, err
	}
	if !accepted {
		accepted, opToken, err = parser.accept(lexers.VICLexerTypeMinus)
		if err != nil {
			return nil, err
		}
	}
	if accepted {
		child, err := parser.parseUnary()
		if err != nil {
			return nil, err
		}
		return asts.NewASTNode(opToken, VICParserNodeTypeOperator, []*asts.ASTNode{child}), nil
	}

	return parser.parsePrimary()
}

func (parser *VICParser) parsePrimary() (*asts.ASTNode, error) {
	accepted, token, err := parser.accept(lexers.VICLexerTypeNumber)
	if err != nil {
		return nil, err
	}
	if accepted {
		return asts.NewASTNode(token, VICParserNodeTypeNumber, nil), nil
	}

	accepted, token, err = parser.accept(lexers.VICLexerTypeIdentifier)
	if err != nil {
		return nil, err
	}
	if accepted {
		return asts.NewASTNode(token, VICParserNodeTypeIdentifier, nil), nil
	}

	accepted, _, err = parser.accept(lexers.VICLexerTypeLParen)
	if err != nil {
		return nil, err
	}
	if accepted {
		expr, err := parser.parseSum()
		if err != nil {
			return nil, err
		}
		if err := parser.expect(lexers.VICLexerTypeRParen); err != nil {
			return nil, err
		}
		return expr, nil
	}

	return nil, errors.New("syntax error: expected int literal, identifier, or '('")
}

func (parser *VICParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
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

func (parser *VICParser) expect(tokenType tokens.TokenType) error {
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
			"expect: expected %s; got %s (%q)",
			tokenType,
			lookaheadToken.Type,
			string(lookaheadToken.Lexeme),
		)
	}
	return nil
}

func (parser *VICParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}
