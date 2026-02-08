package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type ArithParser struct {
	Trace *ArithParserTraceHooks
}

type ArithParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action ArithParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewArithParser() *ArithParser { return &ArithParser{} }

func (parser *ArithParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := ArithParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case ArithParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case ArithParserActionReduce:
			prod := ArithParserProductions[action.Target]
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
			nextState, ok := ArithParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case ArithParserActionAccept:
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

type ArithParserActionKind int

const (
	ArithParserActionShift ArithParserActionKind = iota
	ArithParserActionReduce
	ArithParserActionAccept
)

type ArithParserAction struct {
	Kind   ArithParserActionKind
	Target int
}

type ArithParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var ArithParserActions = map[int]map[tokens.TokenType]ArithParserAction{
	0: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: ArithParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: ArithParserActionShift, Target: 8},
		tokens.TokenType("plus"):  {Kind: ArithParserActionShift, Target: 9},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: ArithParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: ArithParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: ArithParserActionShift, Target: 12},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: ArithParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: ArithParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: ArithParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: ArithParserActionReduce, Target: 10},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: ArithParserActionAccept},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: ArithParserActionReduce, Target: 1},
	},
	7: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: ArithParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: ArithParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: ArithParserActionReduce, Target: 11},
	},
	8: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	9: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	10: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	11: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	12: {
		tokens.TokenType("int_literal"): {Kind: ArithParserActionShift, Target: 7},
	},
	13: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: ArithParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: ArithParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: ArithParserActionShift, Target: 12},
	},
	14: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: ArithParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: ArithParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: ArithParserActionShift, Target: 12},
	},
	15: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: ArithParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: ArithParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: ArithParserActionReduce, Target: 8},
	},
	16: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: ArithParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: ArithParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: ArithParserActionReduce, Target: 9},
	},
	17: {
		tokens.TokenTypeEOF:        {Kind: ArithParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: ArithParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: ArithParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: ArithParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: ArithParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: ArithParserActionReduce, Target: 7},
	},
}

var ArithParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("AddSubTerm"):           1,
		asts.NodeType("MulDivTerm"):           2,
		asts.NodeType("PrecedenceChainStart"): 3,
		asts.NodeType("RHSValue"):             4,
		asts.NodeType("Root"):                 5,
		asts.NodeType("Rvalue"):               6,
	},
	8: {
		asts.NodeType("MulDivTerm"): 13,
		asts.NodeType("RHSValue"):   4,
	},
	9: {
		asts.NodeType("MulDivTerm"): 14,
		asts.NodeType("RHSValue"):   4,
	},
	10: {
		asts.NodeType("RHSValue"): 15,
	},
	11: {
		asts.NodeType("RHSValue"): 16,
	},
	12: {
		asts.NodeType("RHSValue"): 17,
	},
}

var ArithParserProductions = []ArithParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1},
	{lhs: asts.NodeType("Root"), rhsCount: 1},
	{lhs: asts.NodeType("Rvalue"), rhsCount: 1},
	{lhs: asts.NodeType("PrecedenceChainStart"), rhsCount: 1},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 1},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 1},
	{lhs: asts.NodeType("RHSValue"), rhsCount: 1},
}
