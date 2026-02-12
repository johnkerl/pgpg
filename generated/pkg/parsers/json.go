package parsers

import (
	"fmt"
	"os"
	"strings"

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
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			var node *asts.ASTNode
			if prod.hasPassthrough {
				node = rhsNodes[prod.passthroughIndex]
			} else if prod.hasHint {
				nodeType := prod.nodeType
				if nodeType == "" {
					nodeType = prod.lhs
				}
				var parentToken *tokens.Token
				if prod.hasParentLiteral {
					parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
				} else if prod.parentIndex >= 0 && prod.parentIndex < len(rhsNodes) {
					parentToken = rhsNodes[prod.parentIndex].Token
				}
				hintChildren := make([]*asts.ASTNode, len(prod.childIndices))
				for i, ci := range prod.childIndices {
					hintChildren[i] = rhsNodes[ci]
				}
				node = asts.NewASTNode(parentToken, nodeType, hintChildren)
			} else if prod.rhsCount == 1 {
				node = rhsNodes[0]
			} else if prod.rhsCount == 0 {
				node = asts.NewASTNode(nil, prod.lhs, []*asts.ASTNode{})
			} else {
				node = asts.NewASTNode(nil, prod.lhs, rhsNodes)
			}
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

// AttachCLITrace installs tracing hooks for CLI debugging.
func (parser *JSONParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &JSONParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatJSONParserToken(tok))
		},
		OnAction: func(state int, action JSONParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatJSONParserAction(action), tokenTypeNameJSONParser(lookahead), tokenLexemeJSONParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatJSONParserIntStack(stateStack), formatJSONParserNodeStack(nodeStack))
		},
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

func formatJSONParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNameJSONParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemeJSONParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatJSONParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatJSONParserNodeStack(stack []*asts.ASTNode) string {
	parts := make([]string, len(stack))
	for i, node := range stack {
		if node == nil {
			parts[i] = "<nil>"
			continue
		}
		parts[i] = string(node.Type)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatJSONParserAction(action JSONParserAction) string {
	switch action.Kind {
	case JSONParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case JSONParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case JSONParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type JSONParserProduction struct {
	lhs              asts.NodeType
	rhsCount         int
	hasHint          bool
	hasPassthrough   bool
	hasParentLiteral bool
	parentIndex      int
	passthroughIndex int
	parentLiteral    string
	childIndices     []int
	nodeType         asts.NodeType
}

var JSONParserActions = map[int]map[tokens.TokenType]JSONParserAction{
	0: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 5},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 6},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 7},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 8},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 9},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 10},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 11},
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
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 1},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 7},
	},
	6: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 21},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	7: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 26},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	8: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 8},
	},
	9: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 5},
	},
	10: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 4},
	},
	11: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 6},
	},
	12: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 3},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 3},
	},
	13: {
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 28},
	},
	14: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 2},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 2},
	},
	15: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionShift, Target: 30},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 17},
	},
	16: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 7},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 7},
	},
	17: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 32},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	18: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 34},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	19: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 8},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 8},
	},
	20: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 5},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 5},
	},
	21: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 16},
	},
	22: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 4},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 4},
	},
	23: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 6},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 6},
	},
	24: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionShift, Target: 36},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 11},
	},
	25: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 37},
	},
	26: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 10},
	},
	27: {
		tokens.TokenType("colon"): {Kind: JSONParserActionShift, Target: 38},
	},
	28: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 15},
	},
	29: {
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 19},
	},
	30: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	31: {
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 40},
	},
	32: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 16},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 16},
	},
	33: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 41},
	},
	34: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 10},
	},
	35: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 13},
	},
	36: {
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	37: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 9},
	},
	38: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 46},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 47},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 48},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 49},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 50},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 51},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 52},
	},
	39: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionShift, Target: 30},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 17},
	},
	40: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 15},
	},
	41: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 9},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 9},
	},
	42: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionShift, Target: 36},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 11},
	},
	43: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 3},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 3},
	},
	44: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 2},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 2},
	},
	45: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 14},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 14},
	},
	46: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 7},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 7},
	},
	47: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 56},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	48: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 58},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	49: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 8},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 8},
	},
	50: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 5},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 5},
	},
	51: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 4},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 4},
	},
	52: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 6},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 6},
	},
	53: {
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 18},
	},
	54: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 12},
	},
	55: {
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 59},
	},
	56: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 16},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 16},
	},
	57: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 60},
	},
	58: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
	},
	59: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 15},
	},
	60: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 9},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 9},
	},
}

var JSONParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Array"):  1,
		asts.NodeType("Json"):   2,
		asts.NodeType("Object"): 3,
		asts.NodeType("Value"):  4,
	},
	6: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 13,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	7: {
		asts.NodeType("Member"):  24,
		asts.NodeType("Members"): 25,
	},
	15: {
		asts.NodeType("__pgpg_repeat_2"): 29,
	},
	17: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 31,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	18: {
		asts.NodeType("Member"):  24,
		asts.NodeType("Members"): 33,
	},
	24: {
		asts.NodeType("__pgpg_repeat_1"): 35,
	},
	30: {
		asts.NodeType("Array"):  12,
		asts.NodeType("Object"): 14,
		asts.NodeType("Value"):  39,
	},
	36: {
		asts.NodeType("Member"): 42,
	},
	38: {
		asts.NodeType("Array"):  43,
		asts.NodeType("Object"): 44,
		asts.NodeType("Value"):  45,
	},
	39: {
		asts.NodeType("__pgpg_repeat_2"): 53,
	},
	42: {
		asts.NodeType("__pgpg_repeat_1"): 54,
	},
	47: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 55,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	48: {
		asts.NodeType("Member"):  24,
		asts.NodeType("Members"): 57,
	},
}

var JSONParserProductions = []JSONParserProduction{
	{lhs: asts.NodeType("__pgpg_start_3"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Json"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Object"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "{}", childIndices: []int{1}},
	{lhs: asts.NodeType("Object"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "{}", childIndices: []int{1}},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 3, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Members"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}},
	{lhs: asts.NodeType("Member"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}},
	{lhs: asts.NodeType("Array"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "[]", childIndices: []int{1}},
	{lhs: asts.NodeType("Array"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "[]", childIndices: []int{1}},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 0, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 3, hasHint: false, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}},
	{lhs: asts.NodeType("Elements"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}},
}
