package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type ArithWhitespaceParser struct {
	Trace *ArithWhitespaceParserTraceHooks
}

type ArithWhitespaceParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action ArithWhitespaceParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewArithWhitespaceParser() *ArithWhitespaceParser { return &ArithWhitespaceParser{} }

func (parser *ArithWhitespaceParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := ArithWhitespaceParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case ArithWhitespaceParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case ArithWhitespaceParserActionReduce:
			prod := ArithWhitespaceParserProductions[action.Target]
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
			nextState, ok := ArithWhitespaceParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case ArithWhitespaceParserActionAccept:
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

type ArithWhitespaceParserActionKind int

const (
	ArithWhitespaceParserActionShift ArithWhitespaceParserActionKind = iota
	ArithWhitespaceParserActionReduce
	ArithWhitespaceParserActionAccept
)

type ArithWhitespaceParserAction struct {
	Kind   ArithWhitespaceParserActionKind
	Target int
}

type ArithWhitespaceParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var ArithWhitespaceParserActions = map[int]map[tokens.TokenType]ArithWhitespaceParserAction{
	0: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: ArithWhitespaceParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: ArithWhitespaceParserActionShift, Target: 8},
		tokens.TokenType("plus"):  {Kind: ArithWhitespaceParserActionShift, Target: 9},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionShift, Target: 12},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: ArithWhitespaceParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionReduce, Target: 10},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: ArithWhitespaceParserActionAccept},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: ArithWhitespaceParserActionReduce, Target: 1},
	},
	7: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionReduce, Target: 11},
	},
	8: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	9: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	10: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	11: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	12: {
		tokens.TokenType("int_literal"): {Kind: ArithWhitespaceParserActionShift, Target: 7},
	},
	13: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionShift, Target: 12},
	},
	14: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionShift, Target: 12},
	},
	15: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionReduce, Target: 8},
	},
	16: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionReduce, Target: 9},
	},
	17: {
		tokens.TokenTypeEOF:        {Kind: ArithWhitespaceParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: ArithWhitespaceParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: ArithWhitespaceParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: ArithWhitespaceParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: ArithWhitespaceParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: ArithWhitespaceParserActionReduce, Target: 7},
	},
}

var ArithWhitespaceParserGotos = map[int]map[asts.NodeType]int{
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

var ArithWhitespaceParserProductions = []ArithWhitespaceParserProduction{
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
