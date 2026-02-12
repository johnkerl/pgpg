package parsers

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	manuallexers "github.com/johnkerl/pgpg/manual/pkg/lexers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

type PEMDASPlainParser struct {
	Trace *PEMDASPlainParserTraceHooks
}

type PEMDASPlainParserTraceHooks struct {
	OnToken  func(tok *tokens.Token)
	OnAction func(state int, action PEMDASPlainParserAction, lookahead *tokens.Token)
	OnStack  func(stateStack []int, nodeStack []*asts.ASTNode)
}

func NewPEMDASPlainParser() *PEMDASPlainParser { return &PEMDASPlainParser{} }

func (parser *PEMDASPlainParser) Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {
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
		action, ok := PEMDASPlainParserActions[state][lookahead.Type]
		if !ok {
			return nil, fmt.Errorf("parse error: unexpected %s (%q)", lookahead.Type, string(lookahead.Lexeme))
		}
		if parser.Trace != nil && parser.Trace.OnAction != nil {
			parser.Trace.OnAction(state, action, lookahead)
		}
		switch action.Kind {
		case PEMDASPlainParserActionShift:
			nodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))
			stateStack = append(stateStack, action.Target)
			lookahead = lexer.Scan()
			if parser.Trace != nil && parser.Trace.OnToken != nil {
				parser.Trace.OnToken(lookahead)
			}
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASPlainParserActionReduce:
			prod := PEMDASPlainParserProductions[action.Target]
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
			nextState, ok := PEMDASPlainParserGotos[state][prod.lhs]
			if !ok {
				return nil, fmt.Errorf("parse error: missing goto for %s", prod.lhs)
			}
			stateStack = append(stateStack, nextState)
			if parser.Trace != nil && parser.Trace.OnStack != nil {
				parser.Trace.OnStack(stateStack, nodeStack)
			}
		case PEMDASPlainParserActionAccept:
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

type PEMDASPlainParserActionKind int

const (
	PEMDASPlainParserActionShift PEMDASPlainParserActionKind = iota
	PEMDASPlainParserActionReduce
	PEMDASPlainParserActionAccept
)

type PEMDASPlainParserAction struct {
	Kind   PEMDASPlainParserActionKind
	Target int
}

type PEMDASPlainParserProduction struct {
	lhs      asts.NodeType
	rhsCount int
}

var PEMDASPlainParserActions = map[int]map[tokens.TokenType]PEMDASPlainParserAction{
	0: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	1: {
		tokens.TokenTypeEOF:       {Kind: PEMDASPlainParserActionReduce, Target: 3},
		tokens.TokenType("minus"): {Kind: PEMDASPlainParserActionShift, Target: 14},
		tokens.TokenType("plus"):  {Kind: PEMDASPlainParserActionShift, Target: 15},
	},
	2: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 13},
	},
	3: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 18},
	},
	4: {
		tokens.TokenTypeEOF:                {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionShift, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 16},
	},
	5: {
		tokens.TokenTypeEOF:                {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 18},
	},
	6: {
		tokens.TokenTypeEOF: {Kind: PEMDASPlainParserActionReduce, Target: 2},
	},
	7: {
		tokens.TokenTypeEOF: {Kind: PEMDASPlainParserActionAccept},
	},
	8: {
		tokens.TokenTypeEOF: {Kind: PEMDASPlainParserActionReduce, Target: 1},
	},
	9: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 10},
	},
	10: {
		tokens.TokenTypeEOF:                {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 19},
	},
	11: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	12: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	13: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
	},
	14: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	15: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	16: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	17: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	18: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 12},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 13},
	},
	19: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 39},
	},
	20: {
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionShift, Target: 40},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionShift, Target: 41},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 3},
	},
	21: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 13},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 13},
	},
	22: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 6},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 44},
	},
	23: {
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionShift, Target: 45},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("rparen"):         {Kind: PEMDASPlainParserActionReduce, Target: 16},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 16},
	},
	24: {
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("rparen"):         {Kind: PEMDASPlainParserActionReduce, Target: 18},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 18},
	},
	25: {
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionShift, Target: 46},
	},
	26: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 10},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 10},
	},
	27: {
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("rparen"):         {Kind: PEMDASPlainParserActionReduce, Target: 19},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 19},
	},
	28: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	29: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	30: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
	},
	31: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 12},
	},
	32: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 11},
	},
	33: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 18},
	},
	34: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 16},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 17},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 18},
	},
	35: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 8},
	},
	36: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 9},
	},
	37: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 7},
	},
	38: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 14},
	},
	39: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 10},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 11},
	},
	40: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	41: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	42: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	43: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	44: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 29},
		tokens.TokenType("plus"):        {Kind: PEMDASPlainParserActionShift, Target: 30},
	},
	45: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
		tokens.TokenType("minus"):       {Kind: PEMDASPlainParserActionShift, Target: 57},
	},
	46: {
		tokens.TokenTypeEOF:                {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 17},
	},
	47: {
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionShift, Target: 58},
	},
	48: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 12},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 12},
	},
	49: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 11},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 11},
	},
	50: {
		tokens.TokenTypeEOF:        {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 15},
	},
	51: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 5},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 44},
	},
	52: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionShift, Target: 42},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionShift, Target: 43},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 4},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionShift, Target: 44},
	},
	53: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 8},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 8},
	},
	54: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 9},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 9},
	},
	55: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 7},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 7},
	},
	56: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 14},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 14},
	},
	57: {
		tokens.TokenType("int_literal"): {Kind: PEMDASPlainParserActionShift, Target: 27},
		tokens.TokenType("lparen"):      {Kind: PEMDASPlainParserActionShift, Target: 28},
	},
	58: {
		tokens.TokenType("divide"):         {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("exponentiation"): {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("minus"):          {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("modulo"):         {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("plus"):           {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("rparen"):         {Kind: PEMDASPlainParserActionReduce, Target: 17},
		tokens.TokenType("times"):          {Kind: PEMDASPlainParserActionReduce, Target: 17},
	},
	59: {
		tokens.TokenType("divide"): {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("minus"):  {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("modulo"): {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("plus"):   {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("rparen"): {Kind: PEMDASPlainParserActionReduce, Target: 15},
		tokens.TokenType("times"):  {Kind: PEMDASPlainParserActionReduce, Target: 15},
	},
}

var PEMDASPlainParserGotos = map[int]map[asts.NodeType]int{
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

var PEMDASPlainParserProductions = []PEMDASPlainParserProduction{
	{lhs: asts.NodeType("__pgpg_start_1"), rhsCount: 1},
	{lhs: asts.NodeType("Root"), rhsCount: 1},
	{lhs: asts.NodeType("Rvalue"), rhsCount: 1},
	{lhs: asts.NodeType("PrecedenceChainStart"), rhsCount: 1},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 3},
	{lhs: asts.NodeType("AddSubTerm"), rhsCount: 1},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 3},
	{lhs: asts.NodeType("MulDivTerm"), rhsCount: 1},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 2},
	{lhs: asts.NodeType("UnaryTerm"), rhsCount: 1},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 3},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 4},
	{lhs: asts.NodeType("ExponentiationTerm"), rhsCount: 1},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 3},
	{lhs: asts.NodeType("ParenTerm"), rhsCount: 1},
	{lhs: asts.NodeType("PrecedenceChainEnd"), rhsCount: 1},
}
