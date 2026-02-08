package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type StatementsParser struct{}

func NewStatementsParser() *StatementsParser { return &StatementsParser{} }

func (parser *StatementsParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
	if lexer == nil {
		return nil, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	lookahead := lexer.Scan()
	for {
		if lookahead == nil {
			return nil, fmt.Errorf("parser: lexer returned nil token")
		}
		if lookahead.Type == tokens.TokenTypeError {
			return nil, fmt.Errorf("lexer error: %s", string(lookahead.Lexeme))
		}
		state := stateStack[len(stateStack)-1]
		action, ok := StatementsParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		switch action.kind {
		case StatementsParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.target)
			lookahead = lexer.Scan()
		case StatementsParserActionReduce:
			prod := StatementsParserProductions[action.target]
			children := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				children[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if prod.rhsCount == 0 {
				children = []*asts.ASTNode{}
			}
			node := asts.NewASTNode(nil, prod.lhs, children)
			nodeStack = append(nodeStack, node)
			state = stateStack[len(stateStack)-1]
			nextState, ok := StatementsParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
		case StatementsParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			return asts.NewAST(nodeStack[0]), nil
		default:
			return nil, fmt.Errorf("parse error: no action")
		}
	}
}

type StatementsParserActionKind int

const (
	StatementsParserActionShift StatementsParserActionKind = iota
	StatementsParserActionReduce
	StatementsParserActionAccept
)

type StatementsParserAction struct {
	kind   StatementsParserActionKind
	target int
}

type StatementsParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var StatementsParserActions = map[int]map[tokens.TokenType]StatementsParserAction{
	0: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 1},
		tokens.TokenType("id"):          {kind: StatementsParserActionShift, target: 7},
		tokens.TokenType("if"):          {kind: StatementsParserActionShift, target: 8},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 9},
		tokens.TokenType("print"):       {kind: StatementsParserActionShift, target: 10},
	},
	1: {
		tokens.TokenType("semicolon"): {kind: StatementsParserActionShift, target: 11},
	},
	2: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 5},
		tokens.TokenType("id"):          {kind: StatementsParserActionReduce, target: 5},
		tokens.TokenType("if"):          {kind: StatementsParserActionReduce, target: 5},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionReduce, target: 5},
		tokens.TokenType("print"):       {kind: StatementsParserActionReduce, target: 5},
	},
	3: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 6},
		tokens.TokenType("id"):          {kind: StatementsParserActionReduce, target: 6},
		tokens.TokenType("if"):          {kind: StatementsParserActionReduce, target: 6},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionReduce, target: 6},
		tokens.TokenType("print"):       {kind: StatementsParserActionReduce, target: 6},
	},
	4: {
		tokens.TokenTypeEOF: {kind: StatementsParserActionAccept},
	},
	5: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 1},
		tokens.TokenType("id"):          {kind: StatementsParserActionShift, target: 7},
		tokens.TokenType("if"):          {kind: StatementsParserActionShift, target: 8},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 9},
		tokens.TokenType("print"):       {kind: StatementsParserActionShift, target: 10},
	},
	6: {
		tokens.TokenTypeEOF: {kind: StatementsParserActionReduce, target: 3},
	},
	7: {
		tokens.TokenType("equals"): {kind: StatementsParserActionShift, target: 13},
	},
	8: {
		tokens.TokenType("lparen"): {kind: StatementsParserActionShift, target: 14},
	},
	9: {
		tokens.TokenType("semicolon"): {kind: StatementsParserActionReduce, target: 8},
	},
	10: {
		tokens.TokenType("lparen"): {kind: StatementsParserActionShift, target: 15},
	},
	11: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 4},
		tokens.TokenType("id"):          {kind: StatementsParserActionReduce, target: 4},
		tokens.TokenType("if"):          {kind: StatementsParserActionReduce, target: 4},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionReduce, target: 4},
		tokens.TokenType("print"):       {kind: StatementsParserActionReduce, target: 4},
	},
	12: {
		tokens.TokenTypeEOF: {kind: StatementsParserActionReduce, target: 2},
	},
	13: {
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 16},
	},
	14: {
		tokens.TokenType("id"):          {kind: StatementsParserActionShift, target: 18},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 19},
	},
	15: {
		tokens.TokenType("id"):          {kind: StatementsParserActionShift, target: 18},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 19},
	},
	16: {
		tokens.TokenType("semicolon"): {kind: StatementsParserActionReduce, target: 7},
	},
	17: {
		tokens.TokenType("rparen"): {kind: StatementsParserActionShift, target: 21},
	},
	18: {
		tokens.TokenType("equals"): {kind: StatementsParserActionShift, target: 22},
	},
	19: {
		tokens.TokenType("rparen"): {kind: StatementsParserActionReduce, target: 8},
	},
	20: {
		tokens.TokenType("rparen"): {kind: StatementsParserActionShift, target: 23},
	},
	21: {
		tokens.TokenType("id"):          {kind: StatementsParserActionShift, target: 7},
		tokens.TokenType("if"):          {kind: StatementsParserActionShift, target: 8},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 9},
		tokens.TokenType("print"):       {kind: StatementsParserActionShift, target: 10},
	},
	22: {
		tokens.TokenType("int_literal"): {kind: StatementsParserActionShift, target: 25},
	},
	23: {
		tokens.TokenType("semicolon"): {kind: StatementsParserActionShift, target: 26},
	},
	24: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 9},
		tokens.TokenType("id"):          {kind: StatementsParserActionReduce, target: 9},
		tokens.TokenType("if"):          {kind: StatementsParserActionReduce, target: 9},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionReduce, target: 9},
		tokens.TokenType("print"):       {kind: StatementsParserActionReduce, target: 9},
	},
	25: {
		tokens.TokenType("rparen"): {kind: StatementsParserActionReduce, target: 7},
	},
	26: {
		tokens.TokenTypeEOF:             {kind: StatementsParserActionReduce, target: 10},
		tokens.TokenType("id"):          {kind: StatementsParserActionReduce, target: 10},
		tokens.TokenType("if"):          {kind: StatementsParserActionReduce, target: 10},
		tokens.TokenType("int_literal"): {kind: StatementsParserActionReduce, target: 10},
		tokens.TokenType("print"):       {kind: StatementsParserActionReduce, target: 10},
	},
}

var StatementsParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Expression"):      1,
		asts.NodeType("IfStatement"):     2,
		asts.NodeType("PrintStatement"):  3,
		asts.NodeType("Program"):         4,
		asts.NodeType("Statement"):       5,
		asts.NodeType("__pgpg_repeat_1"): 6,
	},
	5: {
		asts.NodeType("Expression"):      1,
		asts.NodeType("IfStatement"):     2,
		asts.NodeType("PrintStatement"):  3,
		asts.NodeType("Statement"):       5,
		asts.NodeType("__pgpg_repeat_1"): 12,
	},
	14: {
		asts.NodeType("Expression"): 17,
	},
	15: {
		asts.NodeType("Expression"): 20,
	},
	21: {
		asts.NodeType("Expression"):     1,
		asts.NodeType("IfStatement"):    2,
		asts.NodeType("PrintStatement"): 3,
		asts.NodeType("Statement"):      24,
	},
}

var StatementsParserProductions = []StatementsParserProduction{
	{lhs: asts.NodeType("__pgpg_start_2"), rhsCount: 1},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 2},
	{lhs: asts.NodeType("Program"), rhsCount: 1},
	{lhs: asts.NodeType("Statement"), rhsCount: 2},
	{lhs: asts.NodeType("Statement"), rhsCount: 1},
	{lhs: asts.NodeType("Statement"), rhsCount: 1},
	{lhs: asts.NodeType("Expression"), rhsCount: 3},
	{lhs: asts.NodeType("Expression"), rhsCount: 1},
	{lhs: asts.NodeType("IfStatement"), rhsCount: 5},
	{lhs: asts.NodeType("PrintStatement"), rhsCount: 5},
}
