package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

type JSONPlainParser struct {
	Trace            *JSONPlainParserTraceHooks
	stashedLookahead *tokens.Token
}

type JSONPlainParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action JSONPlainParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewJSONPlainParser() *JSONPlainParser { return &JSONPlainParser{} }

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var JSONPlainParserNoASTSentinel = &asts.ASTNode{}

func (parser *JSONPlainParser) Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error) {
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
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONPlainParserNoASTSentinel)
			} else {
				nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			}
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
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONPlainParserNoASTSentinel)
			} else {
				if prod.rhsCount == 0 {
					rhsNodes = []*asts.ASTNode{}
				}
				node := asts.NewASTNode(nil, prod.lhs, rhsNodes)
				nodeStack = append(nodeStack, node)
			}
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
			if astMode == "noast" {
				return nil, nil
			}
			return asts.NewAST(nodeStack[0]), nil
		case JSONPlainParserActionAcceptAndYield:
			return nil, fmt.Errorf("parse error: multiple objects; use ParseOne for multi-object input")
		default:
			return nil, fmt.Errorf("parse error: no action")
		}
	}
}

// ParseOne parses one record from the lexer. It is for multi-object input: call in a loop until done.
// Returns (ast, true, nil) on EOF after a record, (ast, false, nil) when more input follows, or (nil, false, err) on error.
func (parser *JSONPlainParser) ParseOne(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, bool, error) {
	if lexer == nil {
		return nil, false, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	var lookahead *tokens.Token
	if parser.stashedLookahead != nil {
		lookahead = parser.stashedLookahead
		parser.stashedLookahead = nil
	} else {
		lookahead = lexer.Scan()
	}
	if parser.Trace != nil && parser.Trace.OnToken != nil {
		parser.Trace.OnToken(lookahead)
	}
	for {
		if lookahead == nil {
			return nil, false, fmt.Errorf("parser: lexer returned nil token")
		}
		if lookahead.Type == tokens.TokenTypeError {
			return nil, false, fmt.Errorf("lexer error: %s", string(lookahead.Lexeme))
		}
		state := stateStack[len(stateStack)-1]
		action, ok := JSONPlainParserActions[state][lookahead.Type]
		if !ok {
			return nil, false, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case JSONPlainParserActionShift:
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONPlainParserNoASTSentinel)
			} else {
				nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			}
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
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONPlainParserNoASTSentinel)
			} else {
				if prod.rhsCount == 0 {
					rhsNodes = []*asts.ASTNode{}
				}
				node := asts.NewASTNode(nil, prod.lhs, rhsNodes)
				nodeStack = append(nodeStack, node)
			}
			state = stateStack[len(stateStack)-1]
			nextState, ok := JSONPlainParserGotos[state][prod.lhs]
			if !ok {
				return nil, false, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case JSONPlainParserActionAccept:
			if len(nodeStack) != 1 {
				return nil, false, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
			if astMode == "noast" {
				return nil, true, nil
			}
			return asts.NewAST(nodeStack[0]), true, nil
		case JSONPlainParserActionAcceptAndYield:
			if len(nodeStack) != 1 {
				return nil, false, fmt.Errorf("parse error: unexpected parse stack size %d", len(nodeStack))
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
			parser.stashedLookahead = lookahead
			if astMode == "noast" {
				return nil, false, nil
			}
			return asts.NewAST(nodeStack[0]), false, nil
		default:
			return nil, false, fmt.Errorf("parse error: no action")
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
	JSONPlainParserActionAcceptAndYield
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
	case JSONPlainParserActionAcceptAndYield:
		return "accept_and_yield"
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
		tokens.TokenTypeEOF:          {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	2: {
		tokens.TokenTypeEOF:          {Kind: JSONPlainParserActionAccept},
		tokens.TokenType("colon"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionAcceptAndYield},
	},
	3: {
		tokens.TokenTypeEOF:        {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	4: {
		tokens.TokenTypeEOF:          {Kind: JSONPlainParserActionReduce, Target: 1},
		tokens.TokenType("colon"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionAcceptAndYield},
	},
	5: {
		tokens.TokenTypeEOF:       {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("false"): {Kind: JSONPlainParserActionReduce, Target: 7},
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
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 30},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 31},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 32},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 33},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 34},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionShift, Target: 35},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 36},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 37},
	},
	8: {
		tokens.TokenTypeEOF:      {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("null"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	9: {
		tokens.TokenTypeEOF:        {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("number"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	10: {
		tokens.TokenTypeEOF:        {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	11: {
		tokens.TokenTypeEOF:      {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("true"): {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	12: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	13: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 38},
	},
	14: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	15: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionShift, Target: 40},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 17},
	},
	16: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	17: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 42},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	18: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 44},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 45},
	},
	19: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	20: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	21: {
		tokens.TokenTypeEOF:          {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	22: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	23: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	24: {
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	25: {
		tokens.TokenType("colon"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 0},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionAcceptAndYield},
	},
	26: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionShift, Target: 47},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 11},
	},
	27: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 48},
	},
	28: {
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	29: {
		tokens.TokenType("colon"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionAcceptAndYield},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionAcceptAndYield},
	},
	30: {
		tokens.TokenType("false"):  {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	31: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 50},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	32: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 52},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 45},
	},
	33: {
		tokens.TokenType("null"):   {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	34: {
		tokens.TokenType("number"): {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	35: {
		tokens.TokenTypeEOF:        {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	36: {
		tokens.TokenType("colon"):  {Kind: JSONPlainParserActionShift, Target: 53},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	37: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("true"):   {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	38: {
		tokens.TokenTypeEOF:          {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	39: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 19},
	},
	40: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	41: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 55},
	},
	42: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	43: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 56},
	},
	44: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	45: {
		tokens.TokenType("colon"): {Kind: JSONPlainParserActionShift, Target: 53},
	},
	46: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 13},
	},
	47: {
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 45},
	},
	48: {
		tokens.TokenTypeEOF:        {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 9},
	},
	49: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 58},
	},
	50: {
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	51: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 59},
	},
	52: {
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	53: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 63},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 64},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 65},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 66},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 67},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 68},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 69},
	},
	54: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionShift, Target: 40},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 17},
	},
	55: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	56: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 9},
	},
	57: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionShift, Target: 47},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 11},
	},
	58: {
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	59: {
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 9},
	},
	60: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 3},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 3},
	},
	61: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 2},
	},
	62: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 14},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 14},
	},
	63: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("false"):  {Kind: JSONPlainParserActionReduce, Target: 7},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 7},
	},
	64: {
		tokens.TokenType("false"):    {Kind: JSONPlainParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONPlainParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONPlainParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONPlainParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 73},
		tokens.TokenType("string"):   {Kind: JSONPlainParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONPlainParserActionShift, Target: 23},
	},
	65: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 75},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionShift, Target: 45},
	},
	66: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("null"):   {Kind: JSONPlainParserActionReduce, Target: 8},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 8},
	},
	67: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("number"): {Kind: JSONPlainParserActionReduce, Target: 5},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 5},
	},
	68: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 4},
		tokens.TokenType("string"): {Kind: JSONPlainParserActionReduce, Target: 4},
	},
	69: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 6},
		tokens.TokenType("true"):   {Kind: JSONPlainParserActionReduce, Target: 6},
	},
	70: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionReduce, Target: 18},
	},
	71: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 12},
	},
	72: {
		tokens.TokenType("rbracket"): {Kind: JSONPlainParserActionShift, Target: 76},
	},
	73: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 16},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 16},
	},
	74: {
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionShift, Target: 77},
	},
	75: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
		tokens.TokenType("rcurly"): {Kind: JSONPlainParserActionReduce, Target: 10},
	},
	76: {
		tokens.TokenType("comma"):    {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("lbracket"): {Kind: JSONPlainParserActionReduce, Target: 15},
		tokens.TokenType("rcurly"):   {Kind: JSONPlainParserActionReduce, Target: 15},
	},
	77: {
		tokens.TokenType("comma"):  {Kind: JSONPlainParserActionReduce, Target: 9},
		tokens.TokenType("lcurly"): {Kind: JSONPlainParserActionReduce, Target: 9},
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
		asts.NodeType("Array"):   24,
		asts.NodeType("Json"):    25,
		asts.NodeType("Member"):  26,
		asts.NodeType("Members"): 27,
		asts.NodeType("Object"):  28,
		asts.NodeType("Value"):   29,
	},
	15: {
		asts.NodeType("__pgpg_repeat_2"): 39,
	},
	17: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 41,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	18: {
		asts.NodeType("Member"):  26,
		asts.NodeType("Members"): 43,
	},
	26: {
		asts.NodeType("__pgpg_repeat_1"): 46,
	},
	31: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 49,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	32: {
		asts.NodeType("Member"):  26,
		asts.NodeType("Members"): 51,
	},
	40: {
		asts.NodeType("Array"):  12,
		asts.NodeType("Object"): 14,
		asts.NodeType("Value"):  54,
	},
	47: {
		asts.NodeType("Member"): 57,
	},
	53: {
		asts.NodeType("Array"):  60,
		asts.NodeType("Object"): 61,
		asts.NodeType("Value"):  62,
	},
	54: {
		asts.NodeType("__pgpg_repeat_2"): 70,
	},
	57: {
		asts.NodeType("__pgpg_repeat_1"): 71,
	},
	64: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 72,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	65: {
		asts.NodeType("Member"):  26,
		asts.NodeType("Members"): 74,
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
