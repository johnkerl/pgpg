package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

// Grammar:
//
// Root
//   : int_literal
//   | int_literal plus Root
//   | int_literal times Root
// ;

type AMEParser struct {
	lexer *lexers.LookaheadLexer
}

func NewAMEParser() AbstractParser[tokens.Token] {
	return &AMEParser{}
}

func (parser *AMEParser) Parse(inputText string) (*asts.AST[tokens.Token], error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewAMLexer(inputText))
	rootNode, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}
	return asts.NewAST(rootNode), nil
}

func (parser *AMEParser) parseSumOrProduct() (*asts.ASTNode[tokens.Token], error) {
	lookaheadToken := parser.lexer.LookAhead()

	if lookaheadToken.IsError() {
		return nil, errors.New(string(lookaheadToken.Lexeme))
	}
	if lookaheadToken.IsEOF() {
		return nil, errors.New("AMEParser: no token found in input")
	}

	if lookaheadToken.Type != lexers.AMLexerTypeNumber {
		actualDesc, err := parser.lexer.DecodeType(lookaheadToken.Type)
		if err != nil {
			return nil, err
		}
		expectedDesc, err := parser.lexer.DecodeType(lexers.AMLexerTypeNumber)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(
			"AMEParser: initial token was of type %s; expected %s",
			actualDesc,
			expectedDesc,
		)
	}

	accepted, acceptedToken, err := parser.accept(lexers.AMLexerTypeNumber)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, errors.New("AMEParser: expected int literal")
	}

	lookaheadToken = parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return nil, errors.New(string(lookaheadToken.Lexeme))
	}

	if lookaheadToken.IsEOF() {
		// The entire expression is a single number
		node := asts.NewASTNode(acceptedToken, nil) // TODO: type
		return node, nil
	}

	if lookaheadToken.Type != lexers.AMLexerTypePlus && lookaheadToken.Type != lexers.AMLexerTypeTimes {
		actualDesc, err := parser.lexer.DecodeType(lookaheadToken.Type)
		if err != nil {
			return nil, err
		}
		expectedPlus, err := parser.lexer.DecodeType(lexers.AMLexerTypePlus)
		if err != nil {
			return nil, err
		}
		expectedTimes, err := parser.lexer.DecodeType(lexers.AMLexerTypeTimes)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(
			"AMEParser: expected %s or %s; got %s",
			expectedPlus,
			expectedTimes,
			actualDesc,
		)
	}

	opToken := lookaheadToken
	if err := parser.expect(opToken.Type); err != nil {
		return nil, err
	}

	// Make a binary node with parent being the plus or times operator, the left child being the
	// previous token, and the right child being TBD.
	rightChild, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	// TODO: type
	leftChild := asts.NewASTNode(acceptedToken, nil)
	parent := asts.NewASTNode(opToken, []*asts.ASTNode[tokens.Token]{leftChild, rightChild})

	return parent, nil
}

func (parser *AMEParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
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

func (parser *AMEParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		// Lex error getting the next token
		return lexerr
	}
	if !accepted {
		// No lex error getting the next token, but the current
		// token isn't of the expected type
		// TODO: describe it: expected & actual type and lexeme
		return errors.New("expect: unexpected symbol")
	}
	return nil
}

func (parser *AMEParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}
