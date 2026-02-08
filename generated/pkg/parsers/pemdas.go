package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type PEMDASParser struct {
	Trace *PEMDASParserTraceHooks
}

type PEMDASParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action PEMDASParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewPEMDASParser() *PEMDASParser { return &PEMDASParser{} }

func (parser *PEMDASParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := PEMDASParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case PEMDASParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASParserActionReduce:
			prod := PEMDASParserProductions[action.Target]
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
			nextState, ok := PEMDASParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASParserActionAccept:
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

type PEMDASParserActionKind int

const (
	PEMDASParserActionShift PEMDASParserActionKind = iota
	PEMDASParserActionReduce
	PEMDASParserActionAccept
)

type PEMDASParserAction struct {
	Kind   PEMDASParserActionKind
	Target int
}

type PEMDASParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var PEMDASParserActions = map[int]map[tokens.TokenType]PEMDASParserAction{
	0: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: PEMDASParserActionShift, Target: 8},
		tokens.TokenType("plus"):  {Kind: PEMDASParserActionShift, Target: 9},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 12},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 10},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionAccept},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionReduce, Target: 1},
	},
	7: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 11},
	},
	8: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	9: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	10: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	11: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	12: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 7},
	},
	13: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 12},
	},
	14: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 12},
	},
	15: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 8},
	},
	16: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 9},
	},
	17: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 7},
	},
}

var PEMDASParserGotos = map[int]map[asts.NodeType]int{
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

var PEMDASParserProductions = []PEMDASParserProduction{
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
