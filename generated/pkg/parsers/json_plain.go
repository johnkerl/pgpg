package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type JSONPlainParser struct {
	Trace *JSONPlainParserTraceHooks
}

type JSONPlainParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action JSONPlainParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewJSONPlainParser() *JSONPlainParser { return &JSONPlainParser{} }

func (parser *JSONPlainParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := JSONPlainParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case JSONPlainParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case JSONPlainParserActionReduce:
			prod := JSONPlainParserProductions[action.Target]
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
			nextState, ok := JSONPlainParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case JSONPlainParserActionAccept:
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
func (parser *JSONPlainParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &JSONPlainParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatJSONPlainParserToken(tok))
		},
		OnAction: func(state int, action JSONPlainParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatJSONPlainParserAction(action), tokenTypeNameJSONPlainParser(lookahead), tokenLexemeJSONPlainParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatJSONPlainParserIntStack(stateStack), formatJSONPlainParserNodeStack(nodeStack))
		},
	}
}

type JSONPlainParserActionKind int

const (
	JSONPlainParserActionShift JSONPlainParserActionKind = iota
	JSONPlainParserActionReduce
	JSONPlainParserActionAccept
)

type JSONPlainParserAction struct {
	Kind   JSONPlainParserActionKind
	Target int
}

func formatJSONPlainParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNameJSONPlainParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemeJSONPlainParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatJSONPlainParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatJSONPlainParserNodeStack(stack []*asts.ASTNode) string {
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

func formatJSONPlainParserAction(action JSONPlainParserAction) string {
	switch action.Kind {
	case JSONPlainParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case JSONPlainParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case JSONPlainParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type JSONPlainParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var JSONPlainParserActions = map[int]map[tokens.TokenType]JSONPlainParserAction{
	0: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 5},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 6},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 7},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 8},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 9},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 10},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 11},
	},
	1: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	2: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionAccept},
	},
	3: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 1},
	},
	5: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	6: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 21},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	7: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 26},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 27},
	},
	8: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	9: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	10: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	11: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	12: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	13: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 28},
	},
	14: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	15: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionShift, Target: 30},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 17},
	},
	16: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	17: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 32},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	18: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 34},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 27},
	},
	19: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	20: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	21: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	22: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	23: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	24: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionShift, Target: 36},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 11},
	},
	25: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 37},
	},
	26: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	27: {
		tokens.TokenType("colon"): {Kind: JSONPlainParserActionShift, Target: 38},
	},
	28: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	29: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 19},
	},
	30: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	31: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 40},
	},
	32: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	33: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 41},
	},
	34: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	35: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 13},
	},
	36: {
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 27},
	},
	37: {
		tokens.TokenTypeEOF: {Kind: JSONPlainParserActionReduce, Target: 9},
	},
	38: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 46},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 47},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 48},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 49},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 50},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 51},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 52},
	},
	39: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionShift, Target: 30},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 17},
	},
	40: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	41: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 9},
	},
	42: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionShift, Target: 36},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 11},
	},
	43: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	44: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	45: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 14},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 14},
	},
	46: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	47: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 56},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	48: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 58},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 27},
	},
	49: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	50: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	51: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	52: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	53: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 18},
	},
	54: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 12},
	},
	55: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 59},
	},
	56: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	57: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 60},
	},
	58: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	59: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	60: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 9},
	},
}

var JSONPlainParserGotos = map[int]map[asts.NodeType]int{
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

var JSONPlainParserProductions = []JSONPlainParserProduction{
	{lhs: asts.NodeType("__pgpg_start_3"), rhsCount: 1},
	{lhs: asts.NodeType("Json"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Value"), rhsCount: 1},
	{lhs: asts.NodeType("Object"), rhsCount: 3},
	{lhs: asts.NodeType("Object"), rhsCount: 2},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 3},
	{lhs: asts.NodeType("Members"), rhsCount: 2},
	{lhs: asts.NodeType("Member"), rhsCount: 3},
	{lhs: asts.NodeType("Array"), rhsCount: 3},
	{lhs: asts.NodeType("Array"), rhsCount: 2},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_2"), rhsCount: 3},
	{lhs: asts.NodeType("Elements"), rhsCount: 2},
}
