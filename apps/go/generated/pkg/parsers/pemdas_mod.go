package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/lib/go/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
	"github.com/johnkerl/pgpg/lib/go/pkg/tokens"
)

type PEMDASModParser struct {
	Trace *PEMDASModParserTraceHooks
}

type PEMDASModParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action PEMDASModParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewPEMDASModParser() *PEMDASModParser { return &PEMDASModParser{} }

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var PEMDASModParserNoASTSentinel = &asts.ASTNode{}

func (parser *PEMDASModParser) Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error) {
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
		action, ok := PEMDASModParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case PEMDASModParserActionShift:
			if astMode == "noast" {
				nodeStack = append(nodeStack, PEMDASModParserNoASTSentinel)
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
		case PEMDASModParserActionReduce:
			prod := PEMDASModParserProductions[action.Target]
			rhsNodes := make([]*asts.ASTNode, prod.rhsCount)
			for i := prod.rhsCount - 1; i >= 0; i-- {
				stateStack = stateStack[:len(stateStack)-1]
				rhsNodes[i] = nodeStack[len(nodeStack)-1]
				nodeStack = nodeStack[:len(nodeStack)-1]
			}
			if astMode == "noast" {
				nodeStack = append(nodeStack, PEMDASModParserNoASTSentinel)
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
			nextState, ok := PEMDASModParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASModParserActionAccept:
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
func (parser *PEMDASModParser) AttachCLITrace(traceTokens bool, traceStates bool, traceStack bool) {
	if !traceTokens && !traceStates && !traceStack {
		return
	}
	parser.Trace = &PEMDASModParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !traceTokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatPEMDASModParserToken(tok))
		},
		OnAction: func(state int, action PEMDASModParserAction, lookahead *tokens.Token) {
			if !traceStates {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n",
				state, formatPEMDASModParserAction(action), tokenTypeNamePEMDASModParser(lookahead), tokenLexemePEMDASModParser(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !traceStack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n",
				formatPEMDASModParserIntStack(stateStack), formatPEMDASModParserNodeStack(nodeStack))
		},
	}
}

type PEMDASModParserActionKind int

const (
	PEMDASModParserActionShift PEMDASModParserActionKind = iota
	PEMDASModParserActionReduce
	PEMDASModParserActionAccept
)

type PEMDASModParserAction struct {
	Kind   PEMDASModParserActionKind
	Target int
}

func formatPEMDASModParserToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d",
		tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeNamePEMDASModParser(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexemePEMDASModParser(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatPEMDASModParserIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatPEMDASModParserNodeStack(stack []*asts.ASTNode) string {
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

func formatPEMDASModParserAction(action PEMDASModParserAction) string {
	switch action.Kind {
	case PEMDASModParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case PEMDASModParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case PEMDASModParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

type PEMDASModParserProduction struct {
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

var PEMDASModParserActions = map[int]map[tokens.TokenType]PEMDASModParserAction{
	0: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: PEMDASModParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: PEMDASModParserActionShift, Target: 15},
		tokens.TokenType("plus"):  {Kind: PEMDASModParserActionShift, Target: 16},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 13},
	},
	3: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 17},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 18},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 19},
	},
	4: {
		tokens.TokenTypeEOF:                {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionShift, Target: 20},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 16},
	},
	5: {
		tokens.TokenTypeEOF:                {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 18},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: PEMDASModParserActionReduce, Target: 2},
	},
	7: {
		tokens.TokenTypeEOF: {Kind: PEMDASModParserActionAccept},
	},
	8: {
		tokens.TokenTypeEOF: {Kind: PEMDASModParserActionReduce, Target: 1},
	},
	9: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 10},
	},
	10: {
		tokens.TokenTypeEOF:                {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 20},
	},
	11: {
		tokens.TokenTypeEOF:                {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 19},
	},
	12: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	13: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	14: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
	},
	15: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	16: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	17: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	18: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	19: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 13},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 14},
	},
	20: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 41},
	},
	21: {
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionShift, Target: 42},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionShift, Target: 43},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 3},
	},
	22: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 13},
	},
	23: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 44},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 45},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 46},
	},
	24: {
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionShift, Target: 47},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("rparen"):         {Kind: PEMDASModParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 16},
	},
	25: {
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("rparen"):         {Kind: PEMDASModParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 18},
	},
	26: {
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionShift, Target: 48},
	},
	27: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 10},
	},
	28: {
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("rparen"):         {Kind: PEMDASModParserActionReduce, Target: 20},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 20},
	},
	29: {
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("rparen"):         {Kind: PEMDASModParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 19},
	},
	30: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	31: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	32: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
	},
	33: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 12},
	},
	34: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 11},
	},
	35: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 17},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 18},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 19},
	},
	36: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 17},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 18},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 19},
	},
	37: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 8},
	},
	38: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 9},
	},
	39: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 7},
	},
	40: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 14},
	},
	41: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 10},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 11},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 12},
	},
	42: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	43: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	44: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	45: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	46: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 31},
		tokens.TokenType("plus"):        {Kind: PEMDASModParserActionShift, Target: 32},
	},
	47: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
		tokens.TokenType("minus"):       {Kind: PEMDASModParserActionShift, Target: 59},
	},
	48: {
		tokens.TokenTypeEOF:                {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 17},
	},
	49: {
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionShift, Target: 60},
	},
	50: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 12},
	},
	51: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 11},
	},
	52: {
		tokens.TokenTypeEOF:        {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 15},
	},
	53: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 44},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 45},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 46},
	},
	54: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionShift, Target: 44},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionShift, Target: 45},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionShift, Target: 46},
	},
	55: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 8},
	},
	56: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 9},
	},
	57: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 7},
	},
	58: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 14},
	},
	59: {
		tokens.TokenType("hex_literal"): {Kind: PEMDASModParserActionShift, Target: 28},
		tokens.TokenType("int_literal"): {Kind: PEMDASModParserActionShift, Target: 29},
		tokens.TokenType("lparen"):      {Kind: PEMDASModParserActionShift, Target: 30},
	},
	60: {
		tokens.TokenType("divide"):         {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("rparen"):         {Kind: PEMDASModParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASModParserActionReduce, Target: 17},
	},
	61: {
		tokens.TokenType("divide"): {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("rparen"): {Kind: PEMDASModParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASModParserActionReduce, Target: 15},
	},
}

var PEMDASModParserGotos = map[int]map[asts.NodeType]int{
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
	12: {
		asts.NodeType("AddSubTerm"):           21,
		asts.NodeType("ExponentiationTerm"):   22,
		asts.NodeType("MulDivTerm"):           23,
		asts.NodeType("ParenTerm"):            24,
		asts.NodeType("PrecedenceChainEnd"):   25,
		asts.NodeType("PrecedenceChainStart"): 26,
		asts.NodeType("UnaryTerm"):            27,
	},
	13: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          33,
	},
	14: {
		asts.NodeType("ExponentiationTerm"): 34,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	15: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         35,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	16: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         36,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	17: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          37,
	},
	18: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          38,
	},
	19: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          39,
	},
	20: {
		asts.NodeType("ExponentiationTerm"): 40,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	30: {
		asts.NodeType("AddSubTerm"):           21,
		asts.NodeType("ExponentiationTerm"):   22,
		asts.NodeType("MulDivTerm"):           23,
		asts.NodeType("ParenTerm"):            24,
		asts.NodeType("PrecedenceChainEnd"):   25,
		asts.NodeType("PrecedenceChainStart"): 49,
		asts.NodeType("UnaryTerm"):            27,
	},
	31: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          50,
	},
	32: {
		asts.NodeType("ExponentiationTerm"): 51,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
	},
	41: {
		asts.NodeType("ExponentiationTerm"): 52,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	42: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("MulDivTerm"):         53,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          27,
	},
	43: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("MulDivTerm"):         54,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          27,
	},
	44: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          55,
	},
	45: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          56,
	},
	46: {
		asts.NodeType("ExponentiationTerm"): 22,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
		asts.NodeType("UnaryTerm"):          57,
	},
	47: {
		asts.NodeType("ExponentiationTerm"): 58,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
	},
	59: {
		asts.NodeType("ExponentiationTerm"): 61,
		asts.NodeType("ParenTerm"):          24,
		asts.NodeType("PrecedenceChainEnd"): 25,
	},
}

var PEMDASModParserProductions = []PEMDASModParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Root"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("Rvalue"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("PrecedenceChainStart"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("unary")},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{1}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("unary")},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 3, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 2}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 4, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 1, passthroughIndex: 0, parentLiteral: "", childIndices: []int{0, 3}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("operator")},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 3, hasHint: false, hasPassthrough: true, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 1, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 1, hasHint: false, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}},
	{lhs: asts.NodeType("PrecedenceChainEnd"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("int_literal")},
	{lhs: asts.NodeType("PrecedenceChainEnd"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("hex_literal")},
}
