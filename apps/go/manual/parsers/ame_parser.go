package parsers

import (
	"errors"
	"fmt"

	"github.com/johnkerl/pgpg/apps/go/manual/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/parsers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

// Grammar:
//
// Root
//   : int_literal
//   | int_literal plus Root
//   | int_literal times Root
// ;

type AMEParser struct {
	lexer *liblexers.LookaheadLexer
}

const (
	AMEParserNodeTypeNumber   asts.NodeType = "number"
	AMEParserNodeTypeOperator asts.NodeType = "operator"
)

func NewAMEParser() parsers.AbstractParser {
	return &AMEParser{}
}

func (parser *AMEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = liblexers.NewLookaheadLexer(lexers.NewAMLexer(inputText))
	rootNode, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}
	return asts.NewAST(rootNode), nil
}

func (parser *AMEParser) parseSumOrProduct() (*asts.ASTNode, error) {
	lookaheadToken := parser.lexer.LookAhead()

	if lookaheadToken.IsError() {
		return nil, errors.New(string(lookaheadToken.Lexeme))
	}
	if lookaheadToken.IsEOF() {
		return nil, errors.New("AMEParser: no token found in input")
	}

	if lookaheadToken.Type != lexers.AMLexerTypeNumber {
		return nil, fmt.Errorf(
			"AMEParser: initial token was of type %s; expected %s",
			lookaheadToken.Type,
			lexers.AMLexerTypeNumber,
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
		node := asts.NewASTNode(acceptedToken, AMEParserNodeTypeNumber, nil)
		return node, nil
	}

	if lookaheadToken.Type != lexers.AMLexerTypePlus && lookaheadToken.Type != lexers.AMLexerTypeTimes {
		return nil, fmt.Errorf(
			"AMEParser: expected %s or %s; got %s",
			lexers.AMLexerTypePlus,
			lexers.AMLexerTypeTimes,
			lookaheadToken.Type,
		)
	}

	opToken := lookaheadToken
	if err := parser.expect(opToken.Type); err != nil {
		return nil, err
	}

	rightChild, err := parser.parseSumOrProduct()
	if err != nil {
		return nil, err
	}
	leftChild := asts.NewASTNode(acceptedToken, AMEParserNodeTypeNumber, nil)
	parent := asts.NewASTNode(opToken, AMEParserNodeTypeOperator, []*asts.ASTNode{leftChild, rightChild})

	return parent, nil
}

func (parser *AMEParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.Type == tokenType {
		lexerr := parser.getAndValidateLookaheadToken()
		return true, lookaheadToken, lexerr
	}
	return false, nil, nil
}

func (parser *AMEParser) expect(tokenType tokens.TokenType) error {
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

func (parser *AMEParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}
	return nil
}
