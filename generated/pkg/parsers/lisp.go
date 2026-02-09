package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type LISPParser struct {
	Trace *LISPParserTraceHooks
}

type LISPParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action LISPParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewLISPParser() *LISPParser { return &LISPParser{} }

func (parser *LISPParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := LISPParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case LISPParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case LISPParserActionReduce:
			prod := LISPParserProductions[action.Target]
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
			nextState, ok := LISPParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case LISPParserActionAccept:
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

type LISPParserActionKind int

const (
	LISPParserActionShift LISPParserActionKind = iota
	LISPParserActionReduce
	LISPParserActionAccept
)

type LISPParserAction struct {
	Kind   LISPParserActionKind
	Target int
}

type LISPParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var LISPParserActions = map[int]map[tokens.TokenType]LISPParserAction{
	0: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 4},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 5},
	},
	1: {
		tokens.TokenTypeEOF: {Kind: LISPParserActionReduce, Target: 1},
	},
	2: {
		tokens.TokenTypeEOF: {Kind: LISPParserActionReduce, Target: 2},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: LISPParserActionAccept},
	},
	4: {
		tokens.TokenTypeEOF: {Kind: LISPParserActionReduce, Target: 6},
	},
	5: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 9},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 10},
	},
	6: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 1},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 1},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 1},
	},
	7: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 2},
	},
	8: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 9},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 3},
	},
	9: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 6},
	},
	10: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 9},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 10},
	},
	11: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 9},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 3},
	},
	12: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionShift, Target: 15},
	},
	13: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 9},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 3},
	},
	14: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionReduce, Target: 4},
	},
	15: {
		tokens.TokenTypeEOF: {Kind: LISPParserActionReduce, Target: 5},
	},
	16: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionShift, Target: 17},
	},
	17: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 5},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 5},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 5},
	},
}

var LISPParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Atom"):         1,
		asts.NodeType("List"):         2,
		asts.NodeType("S_expression"): 3,
	},
	5: {
		asts.NodeType("Atom"):         6,
		asts.NodeType("List"):         7,
		asts.NodeType("S_expression"): 8,
	},
	8: {
		asts.NodeType("Atom"):            6,
		asts.NodeType("List"):            7,
		asts.NodeType("S_expression"):    11,
		asts.NodeType("__pgpg_repeat_1"): 12,
	},
	10: {
		asts.NodeType("Atom"):         6,
		asts.NodeType("List"):         7,
		asts.NodeType("S_expression"): 13,
	},
	11: {
		asts.NodeType("Atom"):            6,
		asts.NodeType("List"):            7,
		asts.NodeType("S_expression"):    11,
		asts.NodeType("__pgpg_repeat_1"): 14,
	},
	13: {
		asts.NodeType("Atom"):            6,
		asts.NodeType("List"):            7,
		asts.NodeType("S_expression"):    11,
		asts.NodeType("__pgpg_repeat_1"): 16,
	},
}

var LISPParserProductions = []LISPParserProduction{
	{lhs: asts.NodeType("__pgpg_start_2"), rhsCount: 1},
	{lhs: asts.NodeType("S_expression"), rhsCount: 1},
	{lhs: asts.NodeType("S_expression"), rhsCount: 1},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 2},
	{lhs: asts.NodeType("List"), rhsCount: 4},
	{lhs: asts.NodeType("Atom"), rhsCount: 1},
}
