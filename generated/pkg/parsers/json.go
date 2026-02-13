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

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var JSONParserNoASTSentinel = &asts.ASTNode{}

func (parser *JSONParser) Parse(lexer manuallexers.AbstractLexer, astMode string) (*asts.AST, error) {
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
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONParserNoASTSentinel)
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
		case JSONParserActionReduce:
			prod := JSONParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if astMode == "noast" {
				nodeStack = append(nodeStack, JSONParserNoASTSentinel)
			} else {
				var node *asts.ASTNode
				useFullTree := (astMode == "fullast")
				if !useFullTree && prod.hasPassthrough {
					node = rhsNodes[prod.passthroughIndex]
				} else if !useFullTree && prod.hasWithAppendedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					for _, ci := range prod.withAppendedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithPrependedChildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withPrependedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasWithAdoptedGrandchildren {
					var parent *asts.ASTNode
					var parentToken *tokens.Token
					var parentType asts.NodeType
					if prod.hasParentLiteral {
						parentToken = tokens.NewToken([]rune(prod.parentLiteral), tokens.TokenType(prod.parentLiteral), tokens.NewTokenLocation())
						parentType = asts.NodeType(prod.parentLiteral)
						parent = nil
					} else {
						parent = rhsNodes[prod.parentIndex]
						parentToken = parent.Token
						parentType = parent.Type
					}
					nodeType := prod.nodeType
					if nodeType == "" {
						nodeType = parentType
					}
					newChildren := make([]*asts.ASTNode, 0)
					for _, ci := range prod.withAdoptedGrandchildren {
						childNode := rhsNodes[ci]
						if childNode != nil && childNode.Children != nil {
							newChildren = append(newChildren, childNode.Children...)
						}
					}
					node = asts.NewASTNode(parentToken, nodeType, newChildren)
				} else if !useFullTree && prod.hasHint {
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
			}
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
			if astMode == "noast" {
				return nil, nil
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
	lhs                         asts.NodeType
	rhsCount                    int
	hasHint                     bool
	hasPassthrough              bool
	hasParentLiteral            bool
	hasWithAppendedChildren     bool
	hasWithPrependedChildren    bool
	hasWithAdoptedGrandchildren bool
	parentIndex                 int
	passthroughIndex            int
	parentLiteral               string
	childIndices                []int
	withAppendedChildren        []int
	withPrependedChildren       []int
	withAdoptedGrandchildren    []int
	nodeType                    asts.NodeType
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
		tokens.TokenType("comma"):    {Kind: JSONParserActionShift, Target: 28},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 29},
	},
	14: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 2},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 2},
	},
	15: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 16},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 16},
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
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 31},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	18: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 33},
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
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 14},
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
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 11},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 11},
	},
	25: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionShift, Target: 34},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 35},
	},
	26: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 9},
	},
	27: {
		tokens.TokenType("colon"): {Kind: JSONParserActionShift, Target: 36},
	},
	28: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	29: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 15},
	},
	30: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionShift, Target: 28},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 38},
	},
	31: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 14},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 14},
	},
	32: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionShift, Target: 34},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 39},
	},
	33: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 9},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 9},
	},
	34: {
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	35: {
		tokens.TokenTypeEOF: {Kind: JSONParserActionReduce, Target: 10},
	},
	36: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 44},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 45},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 46},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 47},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 48},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 49},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 50},
	},
	37: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 17},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 17},
	},
	38: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 15},
	},
	39: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionReduce, Target: 10},
	},
	40: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 12},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 12},
	},
	41: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 3},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 3},
	},
	42: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 2},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 2},
	},
	43: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 13},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 13},
	},
	44: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 7},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 7},
	},
	45: {
		tokens.TokenType("false"):    {Kind: JSONParserActionShift, Target: 16},
		tokens.TokenType("lbracket"): {Kind: JSONParserActionShift, Target: 17},
		tokens.TokenType("lcurly"):   {Kind: JSONParserActionShift, Target: 18},
		tokens.TokenType("null"):     {Kind: JSONParserActionShift, Target: 19},
		tokens.TokenType("number"):   {Kind: JSONParserActionShift, Target: 20},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 52},
		tokens.TokenType("string"):   {Kind: JSONParserActionShift, Target: 22},
		tokens.TokenType("true"):     {Kind: JSONParserActionShift, Target: 23},
	},
	46: {
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 54},
		tokens.TokenType("string"): {Kind: JSONParserActionShift, Target: 27},
	},
	47: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 8},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 8},
	},
	48: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 5},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 5},
	},
	49: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 4},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 4},
	},
	50: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 6},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 6},
	},
	51: {
		tokens.TokenType("comma"):    {Kind: JSONParserActionShift, Target: 28},
		tokens.TokenType("rbracket"): {Kind: JSONParserActionShift, Target: 55},
	},
	52: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 14},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 14},
	},
	53: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionShift, Target: 34},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionShift, Target: 56},
	},
	54: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 9},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 9},
	},
	55: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 15},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 15},
	},
	56: {
		tokens.TokenType("comma"):  {Kind: JSONParserActionReduce, Target: 10},
		tokens.TokenType("rcurly"): {Kind: JSONParserActionReduce, Target: 10},
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
	17: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 30,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	18: {
		asts.NodeType("Member"):  24,
		asts.NodeType("Members"): 32,
	},
	28: {
		asts.NodeType("Array"):  12,
		asts.NodeType("Object"): 14,
		asts.NodeType("Value"):  37,
	},
	34: {
		asts.NodeType("Member"): 40,
	},
	36: {
		asts.NodeType("Array"):  41,
		asts.NodeType("Object"): 42,
		asts.NodeType("Value"):  43,
	},
	45: {
		asts.NodeType("Array"):    12,
		asts.NodeType("Elements"): 51,
		asts.NodeType("Object"):   14,
		asts.NodeType("Value"):    15,
	},
	46: {
		asts.NodeType("Member"):  24,
		asts.NodeType("Members"): 53,
	},
}

var JSONParserProductions = []JSONParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Json"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Value"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Object"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "{}", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("object")},
	{lhs: asts.NodeType("Object"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "{}", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{1}, nodeType: asts.NodeType("object")},
	{lhs: asts.NodeType("Members"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "{temp}", childIndices: []int{0}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Members"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: true, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{2}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Member"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Array"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "[]", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("array")},
	{lhs: asts.NodeType("Array"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: true, parentIndex: 0, passthroughIndex: 0, parentLiteral: "[]", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{1}, nodeType: asts.NodeType("array")},
	{lhs: asts.NodeType("Elements"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: true, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "[temp]", childIndices: []int{0}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Elements"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: true, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{2}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
}
