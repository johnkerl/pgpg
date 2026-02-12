package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type StatementsParser struct {
	Trace *StatementsParserTraceHooks
}

type StatementsParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action StatementsParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewStatementsParser() *StatementsParser { return &StatementsParser{} }

func (parser *StatementsParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
	if lexer == nil {
		return nil, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	lookahead := lexer.Scan()
	if parser.Trace != nil && parser.Trace.OnToken != nil {
		parser.Trace.OnToken(lookahead)
	}
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
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case StatementsParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case StatementsParserActionReduce:
			prod := StatementsParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if prod.rhsCount == 0 {
				rhsNodes = []*asts.ASTNode{}
			}
			node := asts.NewASTNode(nil, prod.lhs, rhsNodes)
			nodeStack = append(nodeStack, node)
			state = stateStack[len(stateStack)-1]
			nextState, ok := StatementsParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case StatementsParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
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
	Kind   StatementsParserActionKind
	Target int
}

type StatementsParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var StatementsParserActions = map[int]map[tokens.TokenType]StatementsParserAction{
	0: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 1},
		tokens.TokenType("id"):          {Kind: StatementsParserActionShift, Target: 7},
		tokens.TokenType("if"):          {Kind: StatementsParserActionShift, Target: 8},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 9},
		tokens.TokenType("print"):       {Kind: StatementsParserActionShift, Target: 10},
	},
	1: {
		tokens.TokenType("semicolon"): {Kind: StatementsParserActionShift, Target: 11},
	},
	2: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 5},
		tokens.TokenType("id"):          {Kind: StatementsParserActionReduce, Target: 5},
		tokens.TokenType("if"):          {Kind: StatementsParserActionReduce, Target: 5},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionReduce, Target: 5},
		tokens.TokenType("print"):       {Kind: StatementsParserActionReduce, Target: 5},
	},
	3: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 6},
		tokens.TokenType("id"):          {Kind: StatementsParserActionReduce, Target: 6},
		tokens.TokenType("if"):          {Kind: StatementsParserActionReduce, Target: 6},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionReduce, Target: 6},
		tokens.TokenType("print"):       {Kind: StatementsParserActionReduce, Target: 6},
	},
	4: {
		tokens.TokenTypeEOF: {Kind: StatementsParserActionAccept},
	},
	5: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 1},
		tokens.TokenType("id"):          {Kind: StatementsParserActionShift, Target: 7},
		tokens.TokenType("if"):          {Kind: StatementsParserActionShift, Target: 8},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 9},
		tokens.TokenType("print"):       {Kind: StatementsParserActionShift, Target: 10},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: StatementsParserActionReduce, Target: 3},
	},
	7: {
		tokens.TokenType("equals"): {Kind: StatementsParserActionShift, Target: 13},
	},
	8: {
		tokens.TokenType("lparen"): {Kind: StatementsParserActionShift, Target: 14},
	},
	9: {
		tokens.TokenType("semicolon"): {Kind: StatementsParserActionReduce, Target: 8},
	},
	10: {
		tokens.TokenType("lparen"): {Kind: StatementsParserActionShift, Target: 15},
	},
	11: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 4},
		tokens.TokenType("id"):          {Kind: StatementsParserActionReduce, Target: 4},
		tokens.TokenType("if"):          {Kind: StatementsParserActionReduce, Target: 4},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionReduce, Target: 4},
		tokens.TokenType("print"):       {Kind: StatementsParserActionReduce, Target: 4},
	},
	12: {
		tokens.TokenTypeEOF: {Kind: StatementsParserActionReduce, Target: 2},
	},
	13: {
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 16},
	},
	14: {
		tokens.TokenType("id"):          {Kind: StatementsParserActionShift, Target: 18},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 19},
	},
	15: {
		tokens.TokenType("id"):          {Kind: StatementsParserActionShift, Target: 18},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 19},
	},
	16: {
		tokens.TokenType("semicolon"): {Kind: StatementsParserActionReduce, Target: 7},
	},
	17: {
		tokens.TokenType("rparen"): {Kind: StatementsParserActionShift, Target: 21},
	},
	18: {
		tokens.TokenType("equals"): {Kind: StatementsParserActionShift, Target: 22},
	},
	19: {
		tokens.TokenType("rparen"): {Kind: StatementsParserActionReduce, Target: 8},
	},
	20: {
		tokens.TokenType("rparen"): {Kind: StatementsParserActionShift, Target: 23},
	},
	21: {
		tokens.TokenType("id"):          {Kind: StatementsParserActionShift, Target: 7},
		tokens.TokenType("if"):          {Kind: StatementsParserActionShift, Target: 8},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 9},
		tokens.TokenType("print"):       {Kind: StatementsParserActionShift, Target: 10},
	},
	22: {
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionShift, Target: 25},
	},
	23: {
		tokens.TokenType("semicolon"): {Kind: StatementsParserActionShift, Target: 26},
	},
	24: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 9},
		tokens.TokenType("id"):          {Kind: StatementsParserActionReduce, Target: 9},
		tokens.TokenType("if"):          {Kind: StatementsParserActionReduce, Target: 9},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionReduce, Target: 9},
		tokens.TokenType("print"):       {Kind: StatementsParserActionReduce, Target: 9},
	},
	25: {
		tokens.TokenType("rparen"): {Kind: StatementsParserActionReduce, Target: 7},
	},
	26: {
		tokens.TokenTypeEOF:             {Kind: StatementsParserActionReduce, Target: 10},
		tokens.TokenType("id"):          {Kind: StatementsParserActionReduce, Target: 10},
		tokens.TokenType("if"):          {Kind: StatementsParserActionReduce, Target: 10},
		tokens.TokenType("int_literal"): {Kind: StatementsParserActionReduce, Target: 10},
		tokens.TokenType("print"):       {Kind: StatementsParserActionReduce, Target: 10},
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
