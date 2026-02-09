package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type JSONParser struct {
	Trace *JSONParserTraceHooks
}

type JSONParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action JSONParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewJSONParser() *JSONParser { return &JSONParser{} }

func (parser *JSONParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := JSONParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case JSONParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case JSONParserActionReduce:
			prod := JSONParserProductions[action.Target]
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
			nextState, ok := JSONParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case JSONParserActionAccept:
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

type JSONParserActionKind int

const (
	JSONParserActionShift JSONParserActionKind = iota
	JSONParserActionReduce
	JSONParserActionAccept
)

type JSONParserAction struct {
	Kind   JSONParserActionKind
	Target int
}

type JSONParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var JSONParserActions = map[int]map[tokens.TokenType]JSONParserAction{
	0: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 6},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 7},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 8},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 9},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 10},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 11},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 12},
	},
	1: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 3},
	},
	2: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionAccept},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 4},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 1},
	},
	6: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("]"):      {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 23},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 24},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 25},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 26},
	},
	7: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 7},
	},
	8: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 30},
	},
	9: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 8},
	},
	10: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 5},
	},
	11: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 9},
	},
	12: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 6},
	},
	13: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 3},
	},
	14: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 31},
	},
	15: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("]"):      {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 23},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 24},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 25},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 26},
	},
	16: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 2},
	},
	17: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 4},
	},
	18: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 19},
	},
	19: {
		tokens.TokenType("]"): {Kind: JSONParserActionShift, Target: 33},
	},
	20: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("]"):      {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 23},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 24},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 25},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 26},
	},
	21: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 7},
	},
	22: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 30},
	},
	23: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 8},
	},
	24: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 5},
	},
	25: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 9},
	},
	26: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 6},
	},
	27: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 36},
	},
	28: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 30},
	},
	29: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 38},
	},
	30: {
		tokens.TokenType("colon"): {Kind: JSONParserActionShift, Target: 39},
	},
	31: {
		tokens.TokenType("comma"): {Kind: JSONParserActionShift, Target: 40},
	},
	32: {
		tokens.TokenType("]"): {Kind: JSONParserActionReduce, Target: 16},
	},
	33: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 17},
	},
	34: {
		tokens.TokenType("]"): {Kind: JSONParserActionShift, Target: 41},
	},
	35: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 42},
	},
	36: {
		tokens.TokenType("comma"): {Kind: JSONParserActionShift, Target: 43},
	},
	37: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 11},
	},
	38: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 12},
	},
	39: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 23},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 24},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 25},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 26},
	},
	40: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 50},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 51},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 52},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 53},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 54},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 55},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 56},
	},
	41: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 17},
	},
	42: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 12},
	},
	43: {
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 58},
	},
	44: {
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 14},
	},
	45: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 3},
	},
	46: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 59},
	},
	47: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 2},
	},
	48: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 4},
	},
	49: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 19},
	},
	50: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("]"):      {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 23},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 24},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 25},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 26},
	},
	51: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 7},
	},
	52: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 30},
	},
	53: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 8},
	},
	54: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 5},
	},
	55: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 9},
	},
	56: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 6},
	},
	57: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 62},
	},
	58: {
		tokens.TokenType("colon"): {Kind: JSONParserActionShift, Target: 63},
	},
	59: {
		tokens.TokenType("["):      {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("]"):      {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("false"):  {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("null"):   {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("number"): {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("string"): {Kind: JSONParserActionReduce, Target: 18},
		tokens.TokenType("true"):   {Kind: JSONParserActionReduce, Target: 18},
	},
	60: {
		tokens.TokenType("]"): {Kind: JSONParserActionShift, Target: 64},
	},
	61: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 65},
	},
	62: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 13},
		tokens.TokenType("string"): {Kind: JSONParserActionReduce, Target: 13},
	},
	63: {
		tokens.TokenType("["):      {Kind: JSONParserActionShift, Target: 50},
		tokens.TokenType("false"):  {Kind: JSONParserActionShift, Target: 51},
		tokens.TokenType("lcurly"): {Kind: JSONParserActionShift, Target: 52},
		tokens.TokenType("null"):   {Kind: JSONParserActionShift, Target: 53},
		tokens.TokenType("number"): {Kind: JSONParserActionShift, Target: 54},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 55},
		tokens.TokenType("true"):   {Kind: JSONParserActionShift, Target: 56},
	},
	64: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 17},
	},
	65: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 12},
	},
	66: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 14},
	},
}

var JSONParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Array"):  1,
		asts.NodeType("Json"):   2,
		asts.NodeType("Object"): 3,
		asts.NodeType("String"): 4,
		asts.NodeType("Value"):  5,
	},
	6: {
		asts.NodeType("Array"):           13,
		asts.NodeType("Element"):         14,
		asts.NodeType("Elements"):        15,
		asts.NodeType("Object"):          16,
		asts.NodeType("String"):          17,
		asts.NodeType("Value"):           18,
		asts.NodeType("__pgpg_repeat_2"): 19,
	},
	8: {
		asts.NodeType("Member"):          27,
		asts.NodeType("Members"):         28,
		asts.NodeType("__pgpg_repeat_1"): 29,
	},
	15: {
		asts.NodeType("Array"):           13,
		asts.NodeType("Element"):         14,
		asts.NodeType("Elements"):        15,
		asts.NodeType("Object"):          16,
		asts.NodeType("String"):          17,
		asts.NodeType("Value"):           18,
		asts.NodeType("__pgpg_repeat_2"): 32,
	},
	20: {
		asts.NodeType("Array"):           13,
		asts.NodeType("Element"):         14,
		asts.NodeType("Elements"):        15,
		asts.NodeType("Object"):          16,
		asts.NodeType("String"):          17,
		asts.NodeType("Value"):           18,
		asts.NodeType("__pgpg_repeat_2"): 34,
	},
	22: {
		asts.NodeType("Member"):          27,
		asts.NodeType("Members"):         28,
		asts.NodeType("__pgpg_repeat_1"): 35,
	},
	28: {
		asts.NodeType("Member"):          27,
		asts.NodeType("Members"):         28,
		asts.NodeType("__pgpg_repeat_1"): 37,
	},
	39: {
		asts.NodeType("Array"):  13,
		asts.NodeType("Object"): 16,
		asts.NodeType("String"): 17,
		asts.NodeType("Value"):  44,
	},
	40: {
		asts.NodeType("Array"):   45,
		asts.NodeType("Element"): 46,
		asts.NodeType("Object"):  47,
		asts.NodeType("String"):  48,
		asts.NodeType("Value"):   49,
	},
	43: {
		asts.NodeType("Member"): 57,
	},
	50: {
		asts.NodeType("Array"):           13,
		asts.NodeType("Element"):         14,
		asts.NodeType("Elements"):        15,
		asts.NodeType("Object"):          16,
		asts.NodeType("String"):          17,
		asts.NodeType("Value"):           18,
		asts.NodeType("__pgpg_repeat_2"): 60,
	},
	52: {
		asts.NodeType("Member"):          27,
		asts.NodeType("Members"):         28,
		asts.NodeType("__pgpg_repeat_1"): 61,
	},
	63: {
		asts.NodeType("Array"):  45,
		asts.NodeType("Object"): 47,
		asts.NodeType("String"): 48,
		asts.NodeType("Value"):  66,
	},
}

var JSONParserProductions = []JSONParserProduction{
	{lhs: asts.NodeType("__pgpg_start_3"), rhsCount: 1},
	{lhs: asts.NodeType("Json"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("String"), rhsCount: 1},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 2},
	{lhs: asts.NodeType("Object"), rhsCount: 3},
	{lhs: asts.NodeType("Members"), rhsCount: 5},
	{lhs: asts.NodeType("Member"), rhsCount: 3},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 2},
	{lhs: asts.NodeType("Array"), rhsCount: 3},
	{lhs: asts.NodeType("Elements"), rhsCount: 5},
	{lhs: asts.NodeType("Element"), rhsCount: 1},
}
