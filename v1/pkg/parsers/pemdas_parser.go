package parsers

import (
	"errors"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

type PEMDASParser struct {
	lexer *lexers.LookaheadLexer
}

func NewPEMDASParser() AbstractParser[tokens.Token] {
	return &PEMDASParser{}
}

func (parser *PEMDASParser) Parse(inputText string) (*asts.AST[tokens.Token], error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewPEMDASLexer(inputText))

	rootNode, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

func (parser *PEMDASParser) parseSum() (*asts.ASTNode[tokens.Token], error) {
	// Sum : Product RestOfSum ;
	left, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	return parser.parseRestOfSum(left)
}

func (parser *PEMDASParser) parseRestOfSum(left *asts.ASTNode[tokens.Token]) (*asts.ASTNode[tokens.Token], error) {
	// RestOfSum
	//   : plus Product RestOfSum
	//   | minus Product RestOfSum
	//   | empty
	// ;
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
		return left, nil // empty production rule
	}

	right, err := parser.parseProduct()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, []*asts.ASTNode[tokens.Token]{left, right})
	return parser.parseRestOfSum(parent)
}

func (parser *PEMDASParser) parseProduct() (*asts.ASTNode[tokens.Token], error) {
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
func (parser *PEMDASParser) parseRestOfProduct(left *asts.ASTNode[tokens.Token]) (*asts.ASTNode[tokens.Token], error) {
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
		return left, nil // empty production rule
	}

	right, err := parser.parsePower()
	if err != nil {
		return nil, err
	}
	parent := asts.NewASTNode(opToken, []*asts.ASTNode[tokens.Token]{left, right})
	return parser.parseRestOfProduct(parent)
}

func (parser *PEMDASParser) parsePower() (*asts.ASTNode[tokens.Token], error) {
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
func (parser *PEMDASParser) parseRestOfPower(left *asts.ASTNode[tokens.Token]) (*asts.ASTNode[tokens.Token], error) {
	accepted, opToken, err := parser.accept(lexers.PEMDASLexerTypePower)
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
	parent := asts.NewASTNode(opToken, []*asts.ASTNode[tokens.Token]{left, right})
	return parent, nil
}

func (parser *PEMDASParser) parseUnary() (*asts.ASTNode[tokens.Token], error) {
	// Unary
	//   : plus Unary
	//   | minus Unary
	//   | Primary
	// ;
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
		return asts.NewASTNode(opToken, []*asts.ASTNode[tokens.Token]{child}), nil
	}

	return parser.parsePrimary()
}

func (parser *PEMDASParser) parsePrimary() (*asts.ASTNode[tokens.Token], error) {
	accepted, token, err := parser.accept(lexers.PEMDASLexerTypeNumber)
	if err != nil {
		return nil, err
	}
	if accepted {
		return asts.NewASTNode(token, nil), nil
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
		// The current token is of the expected type, but there may or may not have been a lex error
		// getting the next token
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	// The current token is not of the expected type
	return false, nil, nil
}

func (parser *PEMDASParser) expect(tokenType tokens.TokenType) error {
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

func (parser *PEMDASParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}
