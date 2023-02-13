package parsers

import (
	"errors"
	//"fmt"

	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/lexers"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

// ----------------------------------------------------------------
// Original grammar:
// Root : Sum ;
//
// Sum
//   : Sum plus Product
//   | Product ;
//
// Product
//   : Product times int_literal
//   | int_literal ;
// ----------------------------------------------------------------

// ----------------------------------------------------------------
// Factored grammar:
//
// Root : Sum ;
//
// Sum : Product RestOfSum ;
//
// RestOfSum
//   : plus Product RestOfSum
//   | empty ;
//
// Product : int_literal RestOfProduct ;
//
// RestOfProduct
//   : times int_literal RestOfProduct
//   | empty ;
// ----------------------------------------------------------------

type AMNEParser struct {
	lexer *lexers.LookaheadLexer
}

func NewAMNEParser() AbstractParser {
	return &AMNEParser{}
}

// My goal (not the only possible goal): map input string -> tokens -> AST
func (parser *AMNEParser) Parse(inputText string) (*asts.AST, error) {
	parser.lexer = lexers.NewLookaheadLexer(lexers.NewAMLexer(inputText))

	rootNode, err := parser.parseSum()
	if err != nil {
		return nil, err
	}

	if err := parser.expect(tokens.TokenTypeEOF); err != nil {
		return nil, err
	}

	return asts.NewAST(rootNode), nil
}

// ----------------------------------------------------------------
func (parser *AMNEParser) parseSum() (*asts.ASTNode, error) {

	// TODO:
	// Sum : Product RestOfSum ;

	//    if (parser.lookaheadToken == lexers.AMLexerTypePlus) {
	//        if err := parser.getAndValidateLookaheadToken(); err != nil {
	//            return err
	//        }
	//    }
	//    if err := parser.parseProduct(); err != nil {
	//        return nil, err
	//    }
	//    while (parser.lookaheadToken == lexers.AMLexerTypePlus) {
	//        nextsym();
	//        err := parser.getAndValidateLookaheadToken()
	//        if err != nil {
	//            return err
	//        }
	//        parser.parseProduct();
	//    }
	return nil, errors.New("unimpl")
}

// ----------------------------------------------------------------
func (parser *AMNEParser) parseRestOfSum() (*asts.ASTNode, error) {
	// TODO:
	// RestOfSum
	//   : plus Product RestOfSum
	//   | empty
	// ;
	return nil, errors.New("unimpl")
}

// ----------------------------------------------------------------
func (parser *AMNEParser) parseProduct() (*asts.ASTNode, error) {
	// TODO:
	// Product
	//   : int_literal RestOfProduct
	//   | empty
	// ;
	return nil, errors.New("unimpl")
}

// ----------------------------------------------------------------
// parseRestOfProduct implements the following production rule:
//   RestOfProduct
//     : times int_literal RestOfProduct
//     | empty
//   ;
func (parser *AMNEParser) parseRestOfProduct() (*asts.ASTNode, error) {
	accepted, _, err := parser.accept(lexers.AMLexerTypeTimes)
	if err != nil {
		return nil, err
	}
	if accepted {
		return nil, errors.New("needs binary") // TODO: new binary
	} else {
		return nil, nil // empty production rule
	}
	return nil, errors.New("unimpl")
}

// ----------------------------------------------------------------
func (parser *AMNEParser) parseIntLiteral() (*asts.ASTNode, error) {
	accepted, token, err := parser.accept(lexers.AMLexerTypeNumber)
	if accepted && err == nil {
		return asts.NewASTNode(token), nil

	} else {
		return nil, errors.New("syntax error: expected int literal; got " + token.String())
	}
}

// ----------------------------------------------------------------

func (parser *AMNEParser) accept(tokenType tokens.TokenType) (bool, *tokens.Token, error) {
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

func (parser *AMNEParser) expect(tokenType tokens.TokenType) error {
	accepted, _, lexerr := parser.accept(tokenType)
	if lexerr != nil {
		// Lex error getting the next token
		return lexerr
	}
	if !accepted {
		// No lex error getting the next token, but the current
		// token isn't of the expected type
		return errors.New("expect: unexpected symbol") // XXX describe it: expected & actual type and lexeme
	}
	return nil
}

// TODO: copy to AME
func (parser *AMNEParser) getAndValidateLookaheadToken() error {
	parser.lexer.Advance()
	lookaheadToken := parser.lexer.LookAhead()
	if lookaheadToken.IsError() {
		return errors.New(string(lookaheadToken.Lexeme))
	}

	return nil
}
