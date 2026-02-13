package parsers

import (
	"fmt"
	"os"
	"strings"

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

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var PEMDASParserNoASTSentinel = &asts.ASTNode{}

func (parser *PEMDASParser) Parse(lexer manuallexers.AbstractLexer, astMode string) (*asts.AST, error) {
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
			if astMode == "noast" {
				nodeStack = append(nodeStack, PEMDASParserNoASTSentinel)
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
		case PEMDASParserActionReduce:
			prod := PEMDASParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if astMode == "noast" {
				nodeStack = append(nodeStack, PEMDASParserNoASTSentinel)
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
					newChildren := make([]*asts.ASTNode, 0)
					if parent != nil && parent.Children != nil {
						newChildren = append(newChildren, parent.Children...)
					}
					for _, ci := range prod.withAppendedChildren {
						newChildren = append(newChildren, rhsNodes[ci])
					}
					node = asts.NewASTNode(parentToken, parentType, newChildren)
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
func (parser *PEMDASParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &PEMDASParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatPEMDASParserToken(tok))
		},
		OnAction: func(state int, action PEMDASParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatPEMDASParserAction(action), tokenTypeNamePEMDASParser(lookahead), tokenLexemePEMDASParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatPEMDASParserIntStack(stateStack), formatPEMDASParserNodeStack(nodeStack))
		},
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

func formatPEMDASParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNamePEMDASParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemePEMDASParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatPEMDASParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatPEMDASParserNodeStack(stack []*asts.ASTNode) string {
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

func formatPEMDASParserAction(action PEMDASParserAction) string {
	switch action.Kind {
	case PEMDASParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case PEMDASParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case PEMDASParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type PEMDASParserProduction struct {
	lhs                     asts.NodeType
	rhsCount                int
	hasHint                 bool
	hasPassthrough          bool
	hasParentLiteral        bool
	hasWithAppendedChildren bool
	parentIndex             int
	passthroughIndex        int
	parentLiteral           string
	childIndices            []int
	withAppendedChildren    []int
	nodeType                asts.NodeType
}

var PEMDASParserActions = map[int]map[tokens.TokenType]PEMDASParserAction{
	0: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):  {Kind: PEMDASParserActionShift, Target: 15},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 13},
	},
	3: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 18},
	},
	4: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionShift, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 16},
	},
	5: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 18},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionReduce, Target: 2},
	},
	7: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionAccept},
	},
	8: {
		tokens.TokenTypeEOF: {Kind: PEMDASParserActionReduce, Target: 1},
	},
	9: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 10},
	},
	10: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 19},
	},
	11: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	12: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	13: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
	},
	14: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	15: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	16: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	17: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	18: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	19: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 39},
	},
	20: {
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionShift, Target: 40},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionShift, Target: 41},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 3},
	},
	21: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 13},
	},
	22: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 44},
	},
	23: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionShift, Target: 45},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 16},
	},
	24: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 18},
	},
	25: {
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionShift, Target: 46},
	},
	26: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 10},
	},
	27: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 19},
	},
	28: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	29: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	30: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
	},
	31: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 12},
	},
	32: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 11},
	},
	33: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 18},
	},
	34: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 18},
	},
	35: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 8},
	},
	36: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 9},
	},
	37: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 7},
	},
	38: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 14},
	},
	39: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 11},
	},
	40: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	41: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	42: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	43: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	44: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASParserActionShift, Target: 30},
	},
	45: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASParserActionShift, Target: 57},
	},
	46: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 17},
	},
	47: {
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionShift, Target: 58},
	},
	48: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 12},
	},
	49: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 11},
	},
	50: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 15},
	},
	51: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 44},
	},
	52: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 44},
	},
	53: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 8},
	},
	54: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 9},
	},
	55: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 7},
	},
	56: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 14},
	},
	57: {
		tokens.TokenType("int_literal"): {Kind: PEMDASParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASParserActionShift, Target: 28},
	},
	58: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 17},
	},
	59: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 15},
	},
}

var PEMDASParserGotos = map[int]map[asts.NodeType]int{
	0: {
		asts.NodeType("AddSubTerm"):           1,
		asts.NodeType("ExponentiationTerm"):   2,
		asts.NodeType("MulDivTerm"):           3,
		asts.NodeType("ParenTerm"):            4,
		asts.NodeType("PrecedenceChainEnd"):   5,
		asts.NodeType("PrecedenceChainStart"): 6,
		asts.NodeType("Root"):                 7,
		asts.NodeType("Rvalue"):               8,
		asts.NodeType("UnaryTerm"):            9,
	},
	11: {
		asts.NodeType("AddSubTerm"):           20,
		asts.NodeType("ExponentiationTerm"):   21,
		asts.NodeType("MulDivTerm"):           22,
		asts.NodeType("ParenTerm"):            23,
		asts.NodeType("PrecedenceChainEnd"):   24,
		asts.NodeType("PrecedenceChainStart"): 25,
		asts.NodeType("UnaryTerm"):            26,
	},
	12: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          31,
	},
	13: {
		asts.NodeType("ExponentiationTerm"): 32,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	14: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         33,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	15: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         34,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	16: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          35,
	},
	17: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          36,
	},
	18: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          37,
	},
	19: {
		asts.NodeType("ExponentiationTerm"): 38,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	28: {
		asts.NodeType("AddSubTerm"):           20,
		asts.NodeType("ExponentiationTerm"):   21,
		asts.NodeType("MulDivTerm"):           22,
		asts.NodeType("ParenTerm"):            23,
		asts.NodeType("PrecedenceChainEnd"):   24,
		asts.NodeType("PrecedenceChainStart"): 47,
		asts.NodeType("UnaryTerm"):            26,
	},
	29: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          48,
	},
	30: {
		asts.NodeType("ExponentiationTerm"): 49,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
	},
	39: {
		asts.NodeType("ExponentiationTerm"): 50,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	40: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("MulDivTerm"):         51,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          26,
	},
	41: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("MulDivTerm"):         52,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          26,
	},
	42: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          53,
	},
	43: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          54,
	},
	44: {
		asts.NodeType("ExponentiationTerm"): 21,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
		asts.NodeType("UnaryTerm"):          55,
	},
	45: {
		asts.NodeType("ExponentiationTerm"): 56,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
	},
	57: {
		asts.NodeType("ExponentiationTerm"): 59,
		asts.NodeType("ParenTerm"):          23,
		asts.NodeType("PrecedenceChainEnd"): 24,
	},
}

var PEMDASParserProductions = []PEMDASParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("Root"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("Rvalue"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("PrecedenceChainStart"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}, withAppendedChildren: []int{}, nodeType: asts.NodeType("unary")},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}, withAppendedChildren: []int{}, nodeType: asts.NodeType("unary")},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 4, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 3}, withAppendedChildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 3, hasHint: false, hasPassthrough: true, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 1, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}},
	{lhs: asts.NodeType("PrecedenceChainEnd"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, nodeType: asts.NodeType("int_literal")},
}
