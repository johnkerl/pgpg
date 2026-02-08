package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type ArithWhitespaceParser struct{}

func NewArithWhitespaceParser() *ArithWhitespaceParser { return &ArithWhitespaceParser{} }

func (parser *ArithWhitespaceParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := ArithWhitespaceParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		switch action.kind {
		case ArithWhitespaceParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.target)
			lookahead = lexer.Scan()
		case ArithWhitespaceParserActionReduce:
			prod := ArithWhitespaceParserProductions[action.target]
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
		case ArithWhitespaceParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
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
	kind   ArithWhitespaceParserActionKind
	target int
}

type ArithWhitespaceParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var ArithWhitespaceParserActions = map[int]map[tokens.TokenType]ArithWhitespaceParserAction{
	0: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	1: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 6},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionShift, target: 8},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 6},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 6},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionShift, target: 9},
	},
	2: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 10},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionReduce, target: 10},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 10},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionReduce, target: 10},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 10},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionReduce, target: 10},
	},
	3: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 11},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionReduce, target: 11},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 11},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionReduce, target: 11},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 11},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionReduce, target: 11},
	},
	4: {
		tokens.TokenTypeEOF: {kind: ArithWhitespaceParserActionAccept},
	},
	5: {
		tokens.TokenTypeEOF: {kind: ArithWhitespaceParserActionReduce, target: 1},
	},
	6: {
		tokens.TokenTypeEOF: {kind: ArithWhitespaceParserActionReduce, target: 2},
	},
	7: {
		tokens.TokenTypeEOF:       {kind: ArithWhitespaceParserActionReduce, target: 3},
		tokens.TokenType("minus"): {kind: ArithWhitespaceParserActionShift, target: 11},
		tokens.TokenType("plus"):  {kind: ArithWhitespaceParserActionShift, target: 12},
	},
	8: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	9: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	10: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	11: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	12: {
		tokens.TokenType("int_literal"): {kind: ArithWhitespaceParserActionShift, target: 3},
	},
	13: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 8},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionReduce, target: 8},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 8},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionReduce, target: 8},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 8},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionReduce, target: 8},
	},
	14: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 7},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionReduce, target: 7},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 7},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionReduce, target: 7},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 7},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionReduce, target: 7},
	},
	15: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 9},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionReduce, target: 9},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 9},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionReduce, target: 9},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 9},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionReduce, target: 9},
	},
	16: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 5},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionShift, target: 8},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 5},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 5},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionShift, target: 9},
	},
	17: {
		tokens.TokenTypeEOF:        {kind: ArithWhitespaceParserActionReduce, target: 4},
		tokens.TokenType("divide"): {kind: ArithWhitespaceParserActionShift, target: 8},
		tokens.TokenType("minus"):  {kind: ArithWhitespaceParserActionReduce, target: 4},
		tokens.TokenType("modulo"): {kind: ArithWhitespaceParserActionShift, target: 10},
		tokens.TokenType("plus"):   {kind: ArithWhitespaceParserActionReduce, target: 4},
		tokens.TokenType("times"):  {kind: ArithWhitespaceParserActionShift, target: 9},
	},
}

var ArithWhitespaceParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("AddSubTerm"):           7,
		asts.NodeType("MulDivTerm"):           1,
		asts.NodeType("PrecedenceChainStart"): 6,
		asts.NodeType("RHSValue"):             2,
		asts.NodeType("Root"):                 4,
		asts.NodeType("Rvalue"):               5,
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
		asts.NodeType("RHSValue"):   2,
	},
	12: {
		asts.NodeType("MulDivTerm"): 17,
		asts.NodeType("RHSValue"):   2,
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
