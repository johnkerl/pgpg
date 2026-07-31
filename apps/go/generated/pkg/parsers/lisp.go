package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

type LISPParser struct {
	Trace            *LISPParserTraceHooks
	stashedLookahead *tokens.Token
}

type LISPParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action LISPParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewLISPParser() *LISPParser { return &LISPParser{} }

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var LISPParserNoASTSentinel = &asts.ASTNode{}

func (parser *LISPParser) Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error) {
	ast, _, err := parser.parseCore(lexer, astMode, false)
	return ast, err
}

// ParseOne parses one record from the lexer. It is for multi-object input: call in a loop until done.
// Returns (ast, true, nil) on EOF after a record, (ast, false, nil) when more input follows, or (nil, false, err) on error.
func (parser *LISPParser) ParseOne(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, bool, error) {
	return parser.parseCore(lexer, astMode, true)
}

// parseCore runs the LR(1) machine shared by Parse and ParseOne. In yield mode
// (ParseOne) it resumes from any stashed lookahead and honors accept-and-yield
// actions, stashing the lookahead for the next record and returning done=false;
// outside yield mode multi-object input is an error. done is true on accept at
// end of input. The returned AST is nil when astMode == "noast".
func (parser *LISPParser) parseCore(lexer liblexers.AbstractLexer, astMode string, yieldMode bool) (*asts.AST, bool, error) {
	if lexer == nil {
		return nil, false, fmt.Errorf("parser: nil lexer")
	}
	stateStack := []int{0}
	nodeStack := []*asts.ASTNode{}
	var lookahead *tokens.Token
	if yieldMode && parser.stashedLookahead != nil {
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
		action, ok := LISPParserActions[state][lookahead.Type]
		if !ok {
			return nil, false, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case LISPParserActionShift:
			if astMode == "noast" {
				nodeStack = append(nodeStack, LISPParserNoASTSentinel)
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
		case LISPParserActionReduce:
			prod := LISPParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if astMode == "noast" {
				nodeStack = append(nodeStack, LISPParserNoASTSentinel)
			} else {
				if prod.rhsCount == 0 {
					rhsNodes = []*asts.ASTNode{}
				}
				node := asts.NewASTNode(nil, prod.lhs, rhsNodes)
				nodeStack = append(nodeStack, node)
			}
			state = stateStack[len(stateStack)-1]
			nextState, ok := LISPParserGotos[state][prod.lhs]
			if !ok {
				return nil, false, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case LISPParserActionAccept:
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
		case LISPParserActionAcceptAndYield:
			if !yieldMode {
				return nil, false, fmt.Errorf("parse error: multiple objects; use ParseOne for multi-object input")
			}
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
func (parser *LISPParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &LISPParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatLISPParserToken(tok))
		},
		OnAction: func(state int, action LISPParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatLISPParserAction(action), tokenTypeNameLISPParser(lookahead), tokenLexemeLISPParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatLISPParserIntStack(stateStack), formatLISPParserNodeStack(nodeStack))
		},
	}
}

type LISPParserActionKind int

const (
	LISPParserActionShift LISPParserActionKind = iota
	LISPParserActionReduce
	LISPParserActionAccept
	LISPParserActionAcceptAndYield
)

type LISPParserAction struct {
	Kind   LISPParserActionKind
	Target int
}

func formatLISPParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNameLISPParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemeLISPParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatLISPParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatLISPParserNodeStack(stack []*asts.ASTNode) string {
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

func formatLISPParserAction(action LISPParserAction) string {
	switch action.Kind {
	case LISPParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case LISPParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case LISPParserActionAccept:
		return "accept"
	case LISPParserActionAcceptAndYield:
		return "accept_and_yield"
	default:
		return "unknown"
	}
}

type LISPParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var LISPParserActions = map[int]map[tokens.TokenType]LISPParserAction{
	0: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 5},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 6},
	},
	1: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 2},
	},
	2: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionReduce, Target: 3},
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 3},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 3},
	},
	3: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionAccept},
		tokens.TokenType("identifier"): {Kind: LISPParserActionAcceptAndYield},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionAcceptAndYield},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionAcceptAndYield},
	},
	4: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionReduce, Target: 1},
		tokens.TokenType("identifier"): {Kind: LISPParserActionAcceptAndYield},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionAcceptAndYield},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionAcceptAndYield},
	},
	5: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionReduce, Target: 7},
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 7},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 7},
	},
	6: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 11},
	},
	7: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 2},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 2},
	},
	8: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 3},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 3},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 3},
	},
	9: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 11},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 4},
	},
	10: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 7},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 7},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 7},
	},
	11: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 11},
	},
	12: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 11},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 4},
	},
	13: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionShift, Target: 16},
	},
	14: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionShift, Target: 10},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionShift, Target: 11},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 4},
	},
	15: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionReduce, Target: 5},
	},
	16: {
		tokens.TokenTypeEOF:            {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 6},
	},
	17: {
		tokens.TokenType("rparen"): {Kind: LISPParserActionShift, Target: 18},
	},
	18: {
		tokens.TokenType("identifier"): {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("lparen"):     {Kind: LISPParserActionReduce, Target: 6},
		tokens.TokenType("rparen"):     {Kind: LISPParserActionReduce, Target: 6},
	},
}

var LISPParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("Atom"):         1,
		asts.NodeType("List"):         2,
		asts.NodeType("Root"):         3,
		asts.NodeType("S_expression"): 4,
	},
	6: {
		asts.NodeType("Atom"):         7,
		asts.NodeType("List"):         8,
		asts.NodeType("S_expression"): 9,
	},
	9: {
		asts.NodeType("Atom"):            7,
		asts.NodeType("List"):            8,
		asts.NodeType("S_expression"):    12,
		asts.NodeType("__pgpg_repeat_1"): 13,
	},
	11: {
		asts.NodeType("Atom"):         7,
		asts.NodeType("List"):         8,
		asts.NodeType("S_expression"): 14,
	},
	12: {
		asts.NodeType("Atom"):            7,
		asts.NodeType("List"):            8,
		asts.NodeType("S_expression"):    12,
		asts.NodeType("__pgpg_repeat_1"): 15,
	},
	14: {
		asts.NodeType("Atom"):            7,
		asts.NodeType("List"):            8,
		asts.NodeType("S_expression"):    12,
		asts.NodeType("__pgpg_repeat_1"): 17,
	},
}

var LISPParserProductions = []LISPParserProduction{
	{lhs: asts.NodeType("__pgpg_start_2"), rhsCount: 1},
	{lhs: asts.NodeType("Root"), rhsCount: 1},
	{lhs: asts.NodeType("S_expression"), rhsCount: 1},
	{lhs: asts.NodeType("S_expression"), rhsCount: 1},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 0},
	{lhs: asts.NodeType("__pgpg_repeat_1"), rhsCount: 2},
	{lhs: asts.NodeType("List"), rhsCount: 4},
	{lhs: asts.NodeType("Atom"), rhsCount: 1},
}
