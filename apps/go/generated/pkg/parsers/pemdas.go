package parsers

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/go/lib/pkg/lexers"
	"github.com/johnkerl/pgpg/go/lib/pkg/tokens"
)

type PEMDASParser struct {
	Trace            *PEMDASParserTraceHooks
	stashedLookahead *tokens.Token
}

type PEMDASParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action PEMDASParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewPEMDASParser() *PEMDASParser { return &PEMDASParser{} }

// noASTSentinel is used as a placeholder on the node stack when astMode == "noast".
var PEMDASParserNoASTSentinel = &asts.ASTNode{}

func (parser *PEMDASParser) Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error) {
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
		case PEMDASParserActionAcceptAndYield:
			return nil, fmt.Errorf("parse error: multiple objects; use ParseOne for multi-object input")
		default:
			return nil, fmt.Errorf("parse error: no action")
		}
	}
}

// ParseOne parses one record from the lexer. It is for multi-object input: call in a loop until done.
// Returns (ast, true, nil) on EOF after a record, (ast, false, nil) when more input follows, or (nil, false, err) on error.
func (parser *PEMDASParser) ParseOne(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, bool, error) {
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
		action, ok := PEMDASParserActions[state][lookahead.Type]
		if !ok {
			return nil, false, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
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
			nextState, ok := PEMDASParserGotos[state][prod.lhs]
			if !ok {
				return nil, false, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASParserActionAccept:
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
		case PEMDASParserActionAcceptAndYield:
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
	PEMDASParserActionAcceptAndYield
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
	case PEMDASParserActionAcceptAndYield:
		return "accept_and_yield"
	default:
		return "unknown"
	}
}

type PEMDASParserProduction struct {
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

var PEMDASParserActions = map[int]map[tokens.TokenType]PEMDASParserAction{
	0: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	1: {
		tokens.TokenTypeEOF:               {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 16},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 17},
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
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 18},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 19},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 20},
	},
	4: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionShift, Target: 21},
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
		tokens.TokenTypeEOF:               {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionReduce, Target: 2},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionReduce, Target: 2},
	},
	7: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionAccept},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("float_literal"):  {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("hex_literal"):    {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("int_literal"):    {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("lparen"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionAcceptAndYield},
	},
	8: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 1},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("float_literal"):  {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("hex_literal"):    {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("int_literal"):    {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("lparen"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionAcceptAndYield},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionAcceptAndYield},
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
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 21},
	},
	11: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 20},
	},
	12: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 19},
	},
	13: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	14: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	15: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	16: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	17: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	18: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	19: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	20: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 14},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 15},
	},
	21: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 43},
	},
	22: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionReduce, Target: 3},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 44},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 45},
		tokens.TokenType("rparen"):        {Kind: PEMDASParserActionReduce, Target: 3},
	},
	23: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 13},
	},
	24: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 46},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 47},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 48},
	},
	25: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionShift, Target: 49},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 16},
	},
	26: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 18},
	},
	27: {
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionShift, Target: 50},
	},
	28: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 10},
	},
	29: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 21},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 21},
	},
	30: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 20},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 20},
	},
	31: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 19},
	},
	32: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	33: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	34: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
	},
	35: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 12},
	},
	36: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 11},
	},
	37: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 18},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 19},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 20},
	},
	38: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 18},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 19},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 20},
	},
	39: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 8},
	},
	40: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 9},
	},
	41: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 7},
	},
	42: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 14},
	},
	43: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 10},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 11},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 12},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 13},
	},
	44: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	45: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	46: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	47: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	48: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 33},
		tokens.TokenType("plus"):          {Kind: PEMDASParserActionShift, Target: 34},
	},
	49: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
		tokens.TokenType("minus"):         {Kind: PEMDASParserActionShift, Target: 61},
	},
	50: {
		tokens.TokenTypeEOF:                {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 17},
	},
	51: {
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionShift, Target: 62},
	},
	52: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 12},
	},
	53: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 11},
	},
	54: {
		tokens.TokenTypeEOF:        {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 15},
	},
	55: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 46},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 47},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 48},
	},
	56: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionShift, Target: 46},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionShift, Target: 47},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionShift, Target: 48},
	},
	57: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 8},
	},
	58: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 9},
	},
	59: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 7},
	},
	60: {
		tokens.TokenType("divide"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("rparen"): {Kind: PEMDASParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASParserActionReduce, Target: 14},
	},
	61: {
		tokens.TokenType("float_literal"): {Kind: PEMDASParserActionShift, Target: 29},
		tokens.TokenType("hex_literal"):   {Kind: PEMDASParserActionShift, Target: 30},
		tokens.TokenType("int_literal"):   {Kind: PEMDASParserActionShift, Target: 31},
		tokens.TokenType("lparen"):        {Kind: PEMDASParserActionShift, Target: 32},
	},
	62: {
		tokens.TokenType("divide"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("rparen"):         {Kind: PEMDASParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASParserActionReduce, Target: 17},
	},
	63: {
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
	13: {
		asts.NodeType("AddSubTerm"):           22,
		asts.NodeType("ExponentiationTerm"):   23,
		asts.NodeType("MulDivTerm"):           24,
		asts.NodeType("ParenTerm"):            25,
		asts.NodeType("PrecedenceChainEnd"):   26,
		asts.NodeType("PrecedenceChainStart"): 27,
		asts.NodeType("UnaryTerm"):            28,
	},
	14: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          35,
	},
	15: {
		asts.NodeType("ExponentiationTerm"): 36,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	16: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         37,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	17: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("MulDivTerm"):         38,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          9,
	},
	18: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          39,
	},
	19: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          40,
	},
	20: {
		asts.NodeType("ExponentiationTerm"): 2,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
		asts.NodeType("UnaryTerm"):          41,
	},
	21: {
		asts.NodeType("ExponentiationTerm"): 42,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	32: {
		asts.NodeType("AddSubTerm"):           22,
		asts.NodeType("ExponentiationTerm"):   23,
		asts.NodeType("MulDivTerm"):           24,
		asts.NodeType("ParenTerm"):            25,
		asts.NodeType("PrecedenceChainEnd"):   26,
		asts.NodeType("PrecedenceChainStart"): 51,
		asts.NodeType("UnaryTerm"):            28,
	},
	33: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          52,
	},
	34: {
		asts.NodeType("ExponentiationTerm"): 53,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
	},
	43: {
		asts.NodeType("ExponentiationTerm"): 54,
		asts.NodeType("ParenTerm"):          4,
		asts.NodeType("PrecedenceChainEnd"): 5,
	},
	44: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("MulDivTerm"):         55,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          28,
	},
	45: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("MulDivTerm"):         56,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          28,
	},
	46: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          57,
	},
	47: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          58,
	},
	48: {
		asts.NodeType("ExponentiationTerm"): 23,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
		asts.NodeType("UnaryTerm"):          59,
	},
	49: {
		asts.NodeType("ExponentiationTerm"): 60,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
	},
	61: {
		asts.NodeType("ExponentiationTerm"): 63,
		asts.NodeType("ParenTerm"):          25,
		asts.NodeType("PrecedenceChainEnd"): 26,
	},
}

var PEMDASParserProductions = []PEMDASParserProduction{
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
	{lhs: asts.NodeType("PrecedenceChainEnd"), rhsCount: 1, hasHint: true, hasPassthrough: false, hasParentLiteral: false, hasWithAppendedChildren: false, hasWithPrependedChildren: false, hasWithAdoptedGrandchildren: false, parentIndex: 0, passthroughIndex: 0, parentLiteral: "", childIndices: []int{}, withAppendedChildren: []int{}, withPrependedChildren: []int{}, withAdoptedGrandchildren: []int{}, nodeType: asts.NodeType("float_literal")},
}
