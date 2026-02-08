package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type ArithParser struct{}

func NewArithParser() *ArithParser { return &ArithParser{} }

func (parser *ArithParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := ArithParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		switch action.kind {
		case ArithParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.target)
			lookahead = lexer.Scan()
		case ArithParserActionReduce:
			prod := ArithParserProductions[action.target]
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
		case ArithParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
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
	kind   ArithParserActionKind
	target int
}

type ArithParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var ArithParserActions = map[int]map[tokens.TokenType]ArithParserAction{
	0: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	1: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 10},
		tokens.TokenType("divide"): {kind: ArithParserActionReduce, target: 10},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 10},
		tokens.TokenType("modulo"): {kind: ArithParserActionReduce, target: 10},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 10},
		tokens.TokenType("times"):  {kind: ArithParserActionReduce, target: 10},
	},
	2: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 6},
		tokens.TokenType("divide"): {kind: ArithParserActionShift, target: 9},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 6},
		tokens.TokenType("modulo"): {kind: ArithParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 6},
		tokens.TokenType("times"):  {kind: ArithParserActionShift, target: 8},
	},
	3: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 11},
		tokens.TokenType("divide"): {kind: ArithParserActionReduce, target: 11},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 11},
		tokens.TokenType("modulo"): {kind: ArithParserActionReduce, target: 11},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 11},
		tokens.TokenType("times"):  {kind: ArithParserActionReduce, target: 11},
	},
	4: {
		tokens.TokenTypeEOF:       {kind: ArithParserActionReduce, target: 3},
		tokens.TokenType("minus"): {kind: ArithParserActionShift, target: 11},
		tokens.TokenType("plus"):  {kind: ArithParserActionShift, target: 12},
	},
	5: {
		tokens.TokenTypeEOF: {kind: ArithParserActionAccept},
	},
	6: {
		tokens.TokenTypeEOF: {kind: ArithParserActionReduce, target: 1},
	},
	7: {
		tokens.TokenTypeEOF: {kind: ArithParserActionReduce, target: 2},
	},
	8: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	9: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	10: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	11: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	12: {
		tokens.TokenType("int_literal"): {kind: ArithParserActionShift, target: 3},
	},
	13: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 7},
		tokens.TokenType("divide"): {kind: ArithParserActionReduce, target: 7},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 7},
		tokens.TokenType("modulo"): {kind: ArithParserActionReduce, target: 7},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 7},
		tokens.TokenType("times"):  {kind: ArithParserActionReduce, target: 7},
	},
	14: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 8},
		tokens.TokenType("divide"): {kind: ArithParserActionReduce, target: 8},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 8},
		tokens.TokenType("modulo"): {kind: ArithParserActionReduce, target: 8},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 8},
		tokens.TokenType("times"):  {kind: ArithParserActionReduce, target: 8},
	},
	15: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 9},
		tokens.TokenType("divide"): {kind: ArithParserActionReduce, target: 9},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 9},
		tokens.TokenType("modulo"): {kind: ArithParserActionReduce, target: 9},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 9},
		tokens.TokenType("times"):  {kind: ArithParserActionReduce, target: 9},
	},
	16: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 5},
		tokens.TokenType("divide"): {kind: ArithParserActionShift, target: 9},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 5},
		tokens.TokenType("modulo"): {kind: ArithParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 5},
		tokens.TokenType("times"):  {kind: ArithParserActionShift, target: 8},
	},
	17: {
		tokens.TokenTypeEOF:        {kind: ArithParserActionReduce, target: 4},
		tokens.TokenType("divide"): {kind: ArithParserActionShift, target: 9},
		tokens.TokenType("minus"):  {kind: ArithParserActionReduce, target: 4},
		tokens.TokenType("modulo"): {kind: ArithParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithParserActionReduce, target: 4},
		tokens.TokenType("times"):  {kind: ArithParserActionShift, target: 8},
	},
}

var ArithParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("AddSubTerm"):           4,
		asts.NodeType("MulDivTerm"):           2,
		asts.NodeType("PrecedenceChainStart"): 7,
		asts.NodeType("RHSValue"):             1,
		asts.NodeType("Root"):                 5,
		asts.NodeType("Rvalue"):               6,
	},
	8: {
		asts.NodeType("RHSValue"): 13,
	},
	9: {
		asts.NodeType("RHSValue"): 14,
	},
	10: {
		asts.NodeType("RHSValue"): 15,
	},
	11: {
		asts.NodeType("MulDivTerm"): 16,
		asts.NodeType("RHSValue"):   1,
	},
	12: {
		asts.NodeType("MulDivTerm"): 17,
		asts.NodeType("RHSValue"):   1,
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
