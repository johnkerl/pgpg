package lexgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/parsers"
)

// Tables captures DFA transitions and accepting actions for a lexer.
type Tables struct {
	StartState  int                       `json:"start_state"`
	Transitions map[int][]RangeTransition `json:"transitions"`
	Actions     map[int]string            `json:"actions"`
	Rules       map[string]string         `json:"rules,omitempty"`
	Metadata    map[string]string         `json:"metadata,omitempty"`
}

// RangeTransition is a DFA transition on an inclusive rune range.
type RangeTransition struct {
	From rune `json:"from"`
	To   rune `json:"to"`
	Next int  `json:"next"`
}

// MarshalJSON ensures deterministic map ordering for stable output.
func (tables *Tables) MarshalJSON() ([]byte, error) {
	if tables == nil {
		return []byte("null"), nil
	}
	var fields []jsonField

	startStateBytes, err := json.Marshal(tables.StartState)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "start_state", value: startStateBytes})

	transitionsBytes, err := marshalMapIntRangeTransitions(tables.Transitions)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "transitions", value: transitionsBytes})

	actionsBytes, err := marshalMapIntString(tables.Actions)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "actions", value: actionsBytes})

	if len(tables.Rules) > 0 {
		rulesBytes, err := marshalMapStringString(tables.Rules)
		if err != nil {
			return nil, err
		}
		fields = append(fields, jsonField{name: "rules", value: rulesBytes})
	}

	if len(tables.Metadata) > 0 {
		metadataBytes, err := marshalMapStringString(tables.Metadata)
		if err != nil {
			return nil, err
		}
		fields = append(fields, jsonField{name: "metadata", value: metadataBytes})
	}

	return marshalOrderedFields(fields), nil
}

type jsonField struct {
	name  string
	value []byte
}

func marshalOrderedFields(fields []jsonField) []byte {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(field.name))
		buf.WriteByte(':')
		buf.Write(field.value)
	}
	buf.WriteByte('}')
	return buf.Bytes()
}

func marshalMapIntRangeTransitions(m map[int][]RangeTransition) ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(strconv.Itoa(key)))
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalMapIntString(m map[int]string) ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Ints(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(strconv.Itoa(key)))
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalMapStringString(m map[string]string) ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(strconv.Quote(key))
		buf.WriteByte(':')
		valueBytes, err := json.Marshal(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// GenerateTablesFromEBNF parses an EBNF grammar and produces lexer tables.
// Lexer rules must expand to regex-compatible forms; repeats and references are supported.
func GenerateTablesFromEBNF(inputText string) (*Tables, error) {
	return GenerateTablesFromEBNFWithSourceName(inputText, "")
}

func GenerateTablesFromEBNFWithSourceName(inputText string, sourceName string) (*Tables, error) {
	parser := parsers.NewEBNFParserWithSourceName(sourceName)
	ast, err := parser.Parse(inputText)
	if err != nil {
		return nil, err
	}

	ruleDefs, err := extractRuleDefs(ast)
	if err != nil {
		return nil, err
	}

	lexerRuleNames := selectLexerRuleNames(ruleDefs)
	if len(lexerRuleNames) == 0 {
		return nil, fmt.Errorf("no lexer rules found")
	}
	tokenRuleNames := selectTokenRuleNames(ruleDefs)
	if len(tokenRuleNames) == 0 {
		return nil, fmt.Errorf("no lexer token rules found")
	}
	lexerRuleSet := map[string]bool{}
	for _, name := range lexerRuleNames {
		lexerRuleSet[name] = true
	}

	ruleMap := map[string]*asts.ASTNode{}
	for _, rule := range ruleDefs {
		ruleMap[rule.name] = rule.expr
	}

	regexCache := map[string]*regexNode{}
	regexRules := map[string]*regexNode{}
	for _, ruleName := range lexerRuleNames {
		node, err := buildRegexForRule(ruleName, ruleMap, lexerRuleSet, regexCache, map[string]bool{})
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", ruleName, err)
		}
		if canBeEmpty(node) {
			return nil, fmt.Errorf("rule %q expands to empty literal", ruleName)
		}
		regexRules[ruleName] = node
	}

	nfaBuilder := &nfaBuilder{}
	globalStart := nfaBuilder.newState()
	for i, ruleName := range tokenRuleNames {
		node := regexRules[ruleName]
		fragment, err := nfaBuilder.build(node)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", ruleName, err)
		}
		accept := acceptRule{name: ruleName, priority: i}
		for _, state := range fragment.accepts {
			state.accepts = append(state.accepts, accept)
		}
		globalStart.epsilon = append(globalStart.epsilon, fragment.start)
	}

	dfa := buildDFA(globalStart)

	transitions := map[int][]RangeTransition{}
	actions := map[int]string{}
	for _, state := range dfa.states {
		if len(state.transitions) > 0 {
			transitions[state.id] = compressRuneTransitions(state.transitions)
		}
		if state.accept != nil {
			actions[state.id] = state.accept.name
		}
	}

	return &Tables{
		StartState:  dfa.startID,
		Transitions: transitions,
		Actions:     actions,
		Rules:       stringifyRegexRules(regexRules, lexerRuleNames),
	}, nil
}

type ruleDef struct {
	name string
	expr *asts.ASTNode
}

func extractRuleDefs(ast *asts.AST) ([]ruleDef, error) {
	if ast == nil || ast.RootNode == nil {
		return nil, fmt.Errorf("nil AST")
	}
	if ast.RootNode.Type != parsers.EBNFParserNodeTypeGrammar {
		return nil, fmt.Errorf("expected grammar root, got %q", ast.RootNode.Type)
	}
	var rules []ruleDef
	for _, ruleNode := range ast.RootNode.Children {
		if ruleNode.Type != parsers.EBNFParserNodeTypeRule {
			return nil, fmt.Errorf("expected rule node, got %q", ruleNode.Type)
		}
		if err := ruleNode.CheckArity(2); err != nil {
			return nil, err
		}
		nameNode := ruleNode.Children[0]
		exprNode := ruleNode.Children[1]
		if nameNode.Type != parsers.EBNFParserNodeTypeIdentifier || nameNode.Token == nil {
			return nil, fmt.Errorf("rule name must be identifier")
		}
		ruleName := string(nameNode.Token.Lexeme)
		rules = append(rules, ruleDef{name: ruleName, expr: exprNode})
	}
	return rules, nil
}

func selectLexerRuleNames(ruleDefs []ruleDef) []string {
	var names []string
	for _, rule := range ruleDefs {
		if isLexerRuleName(rule.name) {
			names = append(names, rule.name)
		}
	}
	return names
}

func selectTokenRuleNames(ruleDefs []ruleDef) []string {
	var names []string
	for _, rule := range ruleDefs {
		if isTokenRuleName(rule.name) {
			names = append(names, rule.name)
		}
	}
	return names
}

func isLexerRuleName(name string) bool {
	if name == "" {
		return false
	}
	first := []rune(name)[0]
	if first == '!' {
		return true
	}
	return first == '_' || unicode.IsLower(first)
}

func isTokenRuleName(name string) bool {
	return isLexerRuleName(name) && !strings.HasPrefix(name, "_")
}

type regexKind int

const (
	regexLiteral regexKind = iota
	regexConcat
	regexAlternate
	regexOptional
	regexStar
)

type regexNode struct {
	kind     regexKind
	literal  string
	children []*regexNode
}

func buildRegexForRule(
	ruleName string,
	ruleMap map[string]*asts.ASTNode,
	lexerRuleSet map[string]bool,
	cache map[string]*regexNode,
	visiting map[string]bool,
) (*regexNode, error) {
	if node, ok := cache[ruleName]; ok {
		return node, nil
	}
	if visiting[ruleName] {
		return nil, fmt.Errorf("recursive rule expansion is not supported for lexer rules")
	}
	exprNode, ok := ruleMap[ruleName]
	if !ok {
		return nil, fmt.Errorf("undefined rule %q", ruleName)
	}
	visiting[ruleName] = true
	node, err := regexFromAST(exprNode, ruleMap, lexerRuleSet, cache, visiting)
	visiting[ruleName] = false
	if err != nil {
		return nil, err
	}
	cache[ruleName] = node
	return node, nil
}

func regexFromAST(
	node *asts.ASTNode,
	ruleMap map[string]*asts.ASTNode,
	lexerRuleSet map[string]bool,
	cache map[string]*regexNode,
	visiting map[string]bool,
) (*regexNode, error) {
	switch node.Type {
	case parsers.EBNFParserNodeTypeLiteral:
		if node.Token == nil {
			return nil, fmt.Errorf("literal node missing token")
		}
		text := string(node.Token.Lexeme)
		unquoted, err := strconv.Unquote(text)
		if err != nil {
			return nil, fmt.Errorf("invalid literal %q: %w", text, err)
		}
		return &regexNode{kind: regexLiteral, literal: unquoted}, nil
	case parsers.EBNFParserNodeTypeSequence:
		if len(node.Children) == 0 {
			return &regexNode{kind: regexLiteral, literal: ""}, nil
		}
		var children []*regexNode
		for _, child := range node.Children {
			part, err := regexFromAST(child, ruleMap, lexerRuleSet, cache, visiting)
			if err != nil {
				return nil, err
			}
			children = append(children, part)
		}
		if len(children) == 1 {
			return children[0], nil
		}
		return &regexNode{kind: regexConcat, children: children}, nil
	case parsers.EBNFParserNodeTypeAlternates:
		var children []*regexNode
		for _, child := range node.Children {
			part, err := regexFromAST(child, ruleMap, lexerRuleSet, cache, visiting)
			if err != nil {
				return nil, err
			}
			children = append(children, part)
		}
		if len(children) == 1 {
			return children[0], nil
		}
		return &regexNode{kind: regexAlternate, children: children}, nil
	case parsers.EBNFParserNodeTypeOptional:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		part, err := regexFromAST(node.Children[0], ruleMap, lexerRuleSet, cache, visiting)
		if err != nil {
			return nil, err
		}
		return &regexNode{kind: regexOptional, children: []*regexNode{part}}, nil
	case parsers.EBNFParserNodeTypeRepeat:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		part, err := regexFromAST(node.Children[0], ruleMap, lexerRuleSet, cache, visiting)
		if err != nil {
			return nil, err
		}
		return &regexNode{kind: regexStar, children: []*regexNode{part}}, nil
	case parsers.EBNFParserNodeTypeIdentifier:
		if node.Token == nil {
			return nil, fmt.Errorf("identifier node missing token")
		}
		identifier := string(node.Token.Lexeme)
		if !lexerRuleSet[identifier] {
			return nil, fmt.Errorf("identifier %q is not a lexer rule", identifier)
		}
		return buildRegexForRule(identifier, ruleMap, lexerRuleSet, cache, visiting)
	default:
		return nil, fmt.Errorf("unsupported node type %q", node.Type)
	}
}

func canBeEmpty(node *regexNode) bool {
	switch node.kind {
	case regexLiteral:
		return node.literal == ""
	case regexConcat:
		for _, child := range node.children {
			if !canBeEmpty(child) {
				return false
			}
		}
		return true
	case regexAlternate:
		for _, child := range node.children {
			if canBeEmpty(child) {
				return true
			}
		}
		return false
	case regexOptional, regexStar:
		return true
	default:
		return false
	}
}

type acceptRule struct {
	name     string
	priority int
}

type nfaState struct {
	id          int
	epsilon     []*nfaState
	transitions map[rune][]*nfaState
	accepts     []acceptRule
}

type nfaFragment struct {
	start   *nfaState
	accepts []*nfaState
}

type nfaBuilder struct {
	nextID int
}

func (builder *nfaBuilder) newState() *nfaState {
	state := &nfaState{
		id:          builder.nextID,
		transitions: map[rune][]*nfaState{},
	}
	builder.nextID++
	return state
}

func (builder *nfaBuilder) build(node *regexNode) (*nfaFragment, error) {
	switch node.kind {
	case regexLiteral:
		start := builder.newState()
		current := start
		for _, r := range []rune(node.literal) {
			next := builder.newState()
			current.transitions[r] = append(current.transitions[r], next)
			current = next
		}
		return &nfaFragment{start: start, accepts: []*nfaState{current}}, nil
	case regexConcat:
		if len(node.children) == 0 {
			start := builder.newState()
			return &nfaFragment{start: start, accepts: []*nfaState{start}}, nil
		}
		first, err := builder.build(node.children[0])
		if err != nil {
			return nil, err
		}
		current := first
		for _, child := range node.children[1:] {
			next, err := builder.build(child)
			if err != nil {
				return nil, err
			}
			for _, accept := range current.accepts {
				accept.epsilon = append(accept.epsilon, next.start)
			}
			current = &nfaFragment{start: first.start, accepts: next.accepts}
		}
		return current, nil
	case regexAlternate:
		start := builder.newState()
		accept := builder.newState()
		for _, child := range node.children {
			fragment, err := builder.build(child)
			if err != nil {
				return nil, err
			}
			start.epsilon = append(start.epsilon, fragment.start)
			for _, childAccept := range fragment.accepts {
				childAccept.epsilon = append(childAccept.epsilon, accept)
			}
		}
		return &nfaFragment{start: start, accepts: []*nfaState{accept}}, nil
	case regexOptional:
		if len(node.children) != 1 {
			return nil, fmt.Errorf("optional node must have one child")
		}
		start := builder.newState()
		accept := builder.newState()
		start.epsilon = append(start.epsilon, accept)
		fragment, err := builder.build(node.children[0])
		if err != nil {
			return nil, err
		}
		start.epsilon = append(start.epsilon, fragment.start)
		for _, childAccept := range fragment.accepts {
			childAccept.epsilon = append(childAccept.epsilon, accept)
		}
		return &nfaFragment{start: start, accepts: []*nfaState{accept}}, nil
	case regexStar:
		if len(node.children) != 1 {
			return nil, fmt.Errorf("star node must have one child")
		}
		start := builder.newState()
		accept := builder.newState()
		start.epsilon = append(start.epsilon, accept)
		fragment, err := builder.build(node.children[0])
		if err != nil {
			return nil, err
		}
		start.epsilon = append(start.epsilon, fragment.start)
		for _, childAccept := range fragment.accepts {
			childAccept.epsilon = append(childAccept.epsilon, fragment.start, accept)
		}
		return &nfaFragment{start: start, accepts: []*nfaState{accept}}, nil
	default:
		return nil, fmt.Errorf("unsupported regex node type")
	}
}

type dfaState struct {
	id          int
	nfaSet      map[int]*nfaState
	transitions map[rune]int
	accept      *acceptRule
}

type dfaResult struct {
	startID int
	states  []*dfaState
}

func buildDFA(start *nfaState) *dfaResult {
	startSet := epsilonClosure(map[int]*nfaState{start.id: start})
	stateMap := map[string]*dfaState{}
	var states []*dfaState
	queue := []*dfaState{}
	nextID := 0

	newState := func(set map[int]*nfaState) *dfaState {
		state := &dfaState{
			id:          nextID,
			nfaSet:      set,
			transitions: map[rune]int{},
			accept:      selectAcceptRule(set),
		}
		nextID++
		key := setKey(set)
		stateMap[key] = state
		states = append(states, state)
		queue = append(queue, state)
		return state
	}

	startState := newState(startSet)
	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]
		runeTargets := map[rune]map[int]*nfaState{}
		for _, nfa := range state.nfaSet {
			for r, targets := range nfa.transitions {
				targetSet := runeTargets[r]
				if targetSet == nil {
					targetSet = map[int]*nfaState{}
					runeTargets[r] = targetSet
				}
				for _, target := range targets {
					targetSet[target.id] = target
				}
			}
		}
		runes := make([]rune, 0, len(runeTargets))
		for r := range runeTargets {
			runes = append(runes, r)
		}
		sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
		for _, r := range runes {
			targetSet := epsilonClosure(runeTargets[r])
			key := setKey(targetSet)
			targetState, ok := stateMap[key]
			if !ok {
				targetState = newState(targetSet)
			}
			state.transitions[r] = targetState.id
		}
	}

	return &dfaResult{startID: startState.id, states: states}
}

func epsilonClosure(initial map[int]*nfaState) map[int]*nfaState {
	out := map[int]*nfaState{}
	stack := make([]*nfaState, 0, len(initial))
	for _, state := range initial {
		out[state.id] = state
		stack = append(stack, state)
	}
	for len(stack) > 0 {
		state := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range state.epsilon {
			if _, ok := out[next.id]; ok {
				continue
			}
			out[next.id] = next
			stack = append(stack, next)
		}
	}
	return out
}

func setKey(set map[int]*nfaState) string {
	ids := make([]int, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	var b strings.Builder
	for i, id := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, "%d", id)
	}
	return b.String()
}

func selectAcceptRule(set map[int]*nfaState) *acceptRule {
	var best *acceptRule
	for _, state := range set {
		for _, accept := range state.accepts {
			if best == nil || accept.priority < best.priority {
				candidate := accept
				best = &candidate
			}
		}
	}
	return best
}

func compressRuneTransitions(transitions map[rune]int) []RangeTransition {
	if len(transitions) == 0 {
		return nil
	}
	runes := make([]rune, 0, len(transitions))
	for r := range transitions {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	var out []RangeTransition
	start := runes[0]
	prev := runes[0]
	currentNext := transitions[start]
	for i := 1; i < len(runes); i++ {
		r := runes[i]
		next := transitions[r]
		if r == prev+1 && next == currentNext {
			prev = r
			continue
		}
		out = append(out, RangeTransition{From: start, To: prev, Next: currentNext})
		start = r
		prev = r
		currentNext = next
	}
	out = append(out, RangeTransition{From: start, To: prev, Next: currentNext})
	return out
}

func stringifyRegexRules(rules map[string]*regexNode, ruleOrder []string) map[string]string {
	if len(rules) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, name := range ruleOrder {
		node, ok := rules[name]
		if !ok {
			continue
		}
		out[name] = regexToString(node)
	}
	return out
}

func regexToString(node *regexNode) string {
	switch node.kind {
	case regexLiteral:
		return strconv.Quote(node.literal)
	case regexConcat:
		var parts []string
		for _, child := range node.children {
			parts = append(parts, regexToString(child))
		}
		return strings.Join(parts, " ")
	case regexAlternate:
		var parts []string
		for _, child := range node.children {
			parts = append(parts, regexToString(child))
		}
		return "(" + strings.Join(parts, " | ") + ")"
	case regexOptional:
		return "(" + regexToString(node.children[0]) + ")?"
	case regexStar:
		return "(" + regexToString(node.children[0]) + ")*"
	default:
		return "<?>"
	}
}

// EncodeTables returns pretty-printed JSON for tables.
func EncodeTables(tables *Tables) ([]byte, error) {
	return json.MarshalIndent(tables, "", "  ")
}
