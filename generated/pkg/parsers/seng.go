package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type SENGParser struct {
	Trace *SENGParserTraceHooks
}

type SENGParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action SENGParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewSENGParser() *SENGParser { return &SENGParser{} }

func (parser *SENGParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := SENGParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case SENGParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case SENGParserActionReduce:
			prod := SENGParserProductions[action.Target]
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
			nextState, ok := SENGParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case SENGParserActionAccept:
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
func (parser *SENGParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &SENGParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatSENGParserToken(tok))
		},
		OnAction: func(state int, action SENGParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatSENGParserAction(action), tokenTypeNameSENGParser(lookahead), tokenLexemeSENGParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatSENGParserIntStack(stateStack), formatSENGParserNodeStack(nodeStack))
		},
	}
}

type SENGParserActionKind int

const (
	SENGParserActionShift SENGParserActionKind = iota
	SENGParserActionReduce
	SENGParserActionAccept
)

type SENGParserAction struct {
	Kind   SENGParserActionKind
	Target int
}

func formatSENGParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNameSENGParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemeSENGParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatSENGParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatSENGParserNodeStack(stack []*asts.ASTNode) string {
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

func formatSENGParserAction(action SENGParserAction) string {
	switch action.Kind {
	case SENGParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case SENGParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case SENGParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type SENGParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var SENGParserActions = map[int]map[tokens.TokenType]SENGParserAction{
	0: {
		tokens.TokenType("adjective"):                  {Kind: SENGParserActionShift, Target: 6},
		tokens.TokenType("adverb"):                     {Kind: SENGParserActionShift, Target: 7},
		tokens.TokenType("article"):                    {Kind: SENGParserActionShift, Target: 8},
		tokens.TokenType("intransitiveImperativeVerb"): {Kind: SENGParserActionShift, Target: 9},
		tokens.TokenType("noun"):                       {Kind: SENGParserActionShift, Target: 10},
		tokens.TokenType("transitiveImperativeVerb"):   {Kind: SENGParserActionShift, Target: 11},
	},
	1: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 4},
	},
	2: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionShift, Target: 14},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionShift, Target: 15},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionShift, Target: 16},
	},
	3: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionReduce, Target: 5},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionReduce, Target: 5},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionReduce, Target: 5},
	},
	4: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionAccept},
	},
	5: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("article"):   {Kind: SENGParserActionShift, Target: 20},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	6: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 6},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 10},
	},
	7: {
		tokens.TokenType("adverb"):                     {Kind: SENGParserActionShift, Target: 7},
		tokens.TokenType("intransitiveImperativeVerb"): {Kind: SENGParserActionShift, Target: 9},
		tokens.TokenType("transitiveImperativeVerb"):   {Kind: SENGParserActionShift, Target: 11},
	},
	8: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 6},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 10},
	},
	9: {
		tokens.TokenTypeEOF:             {Kind: SENGParserActionReduce, Target: 16},
		tokens.TokenType("preposition"): {Kind: SENGParserActionShift, Target: 26},
	},
	10: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionReduce, Target: 7},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionReduce, Target: 7},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionReduce, Target: 7},
	},
	11: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionReduce, Target: 14},
		tokens.TokenType("article"):   {Kind: SENGParserActionReduce, Target: 14},
		tokens.TokenType("noun"):      {Kind: SENGParserActionReduce, Target: 14},
	},
	12: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 2},
	},
	13: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("article"):   {Kind: SENGParserActionShift, Target: 20},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	14: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionShift, Target: 14},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionShift, Target: 15},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionShift, Target: 16},
	},
	15: {
		tokens.TokenTypeEOF:             {Kind: SENGParserActionReduce, Target: 11},
		tokens.TokenType("preposition"): {Kind: SENGParserActionShift, Target: 30},
	},
	16: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionReduce, Target: 9},
		tokens.TokenType("article"):   {Kind: SENGParserActionReduce, Target: 9},
		tokens.TokenType("noun"):      {Kind: SENGParserActionReduce, Target: 9},
	},
	17: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 3},
	},
	18: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 5},
	},
	19: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	20: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	21: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 7},
	},
	22: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionReduce, Target: 8},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionReduce, Target: 8},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionReduce, Target: 8},
	},
	23: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 17},
	},
	24: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionReduce, Target: 15},
		tokens.TokenType("article"):   {Kind: SENGParserActionReduce, Target: 15},
		tokens.TokenType("noun"):      {Kind: SENGParserActionReduce, Target: 15},
	},
	25: {
		tokens.TokenType("adverb"):           {Kind: SENGParserActionReduce, Target: 6},
		tokens.TokenType("intransitiveVerb"): {Kind: SENGParserActionReduce, Target: 6},
		tokens.TokenType("transitiveVerb"):   {Kind: SENGParserActionReduce, Target: 6},
	},
	26: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("article"):   {Kind: SENGParserActionShift, Target: 20},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	27: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 1},
	},
	28: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 12},
	},
	29: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionReduce, Target: 10},
		tokens.TokenType("article"):   {Kind: SENGParserActionReduce, Target: 10},
		tokens.TokenType("noun"):      {Kind: SENGParserActionReduce, Target: 10},
	},
	30: {
		tokens.TokenType("adjective"): {Kind: SENGParserActionShift, Target: 19},
		tokens.TokenType("article"):   {Kind: SENGParserActionShift, Target: 20},
		tokens.TokenType("noun"):      {Kind: SENGParserActionShift, Target: 21},
	},
	31: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 8},
	},
	32: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 6},
	},
	33: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 18},
	},
	34: {
		tokens.TokenTypeEOF: {Kind: SENGParserActionReduce, Target: 13},
	},
}

var SENGParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("IntransitiveImperativeVerbPhrase"): 1,
		asts.NodeType("NounPhrase"):                       2,
		asts.NodeType("NounPhraseWithoutArticle"):         3,
		asts.NodeType("Root"):                             4,
		asts.NodeType("TransitiveImperativeVerbPhrase"):   5,
	},
	2: {
		asts.NodeType("IntransitiveVerbPhrase"): 12,
		asts.NodeType("TransitiveVerbPhrase"):   13,
	},
	5: {
		asts.NodeType("NounPhrase"):               17,
		asts.NodeType("NounPhraseWithoutArticle"): 18,
	},
	6: {
		asts.NodeType("NounPhraseWithoutArticle"): 22,
	},
	7: {
		asts.NodeType("IntransitiveImperativeVerbPhrase"): 23,
		asts.NodeType("TransitiveImperativeVerbPhrase"):   24,
	},
	8: {
		asts.NodeType("NounPhraseWithoutArticle"): 25,
	},
	13: {
		asts.NodeType("NounPhrase"):               27,
		asts.NodeType("NounPhraseWithoutArticle"): 18,
	},
	14: {
		asts.NodeType("IntransitiveVerbPhrase"): 28,
		asts.NodeType("TransitiveVerbPhrase"):   29,
	},
	19: {
		asts.NodeType("NounPhraseWithoutArticle"): 31,
	},
	20: {
		asts.NodeType("NounPhraseWithoutArticle"): 32,
	},
	26: {
		asts.NodeType("NounPhrase"):               33,
		asts.NodeType("NounPhraseWithoutArticle"): 18,
	},
	30: {
		asts.NodeType("NounPhrase"):               34,
		asts.NodeType("NounPhraseWithoutArticle"): 18,
	},
}

var SENGParserProductions = []SENGParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1},
	{lhs: asts.NodeType("Root"), rhsCount: 3},
	{lhs: asts.NodeType("Root"), rhsCount: 2},
	{lhs: asts.NodeType("Root"), rhsCount: 2},
	{lhs: asts.NodeType("Root"), rhsCount: 1},
	{lhs: asts.NodeType("NounPhrase"), rhsCount: 1},
	{lhs: asts.NodeType("NounPhrase"), rhsCount: 2},
	{lhs: asts.NodeType("NounPhraseWithoutArticle"), rhsCount: 1},
	{lhs: asts.NodeType("NounPhraseWithoutArticle"), rhsCount: 2},
	{lhs: asts.NodeType("TransitiveVerbPhrase"), rhsCount: 1},
	{lhs: asts.NodeType("TransitiveVerbPhrase"), rhsCount: 2},
	{lhs: asts.NodeType("IntransitiveVerbPhrase"), rhsCount: 1},
	{lhs: asts.NodeType("IntransitiveVerbPhrase"), rhsCount: 2},
	{lhs: asts.NodeType("IntransitiveVerbPhrase"), rhsCount: 3},
	{lhs: asts.NodeType("TransitiveImperativeVerbPhrase"), rhsCount: 1},
	{lhs: asts.NodeType("TransitiveImperativeVerbPhrase"), rhsCount: 2},
	{lhs: asts.NodeType("IntransitiveImperativeVerbPhrase"), rhsCount: 1},
	{lhs: asts.NodeType("IntransitiveImperativeVerbPhrase"), rhsCount: 2},
	{lhs: asts.NodeType("IntransitiveImperativeVerbPhrase"), rhsCount: 3},
}
