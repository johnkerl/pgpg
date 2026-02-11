package parsegen

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

const eofSymbol = "EOF"

// SortOutput controls whether output is deterministically sorted.
// When true (default), all maps and item sets are sorted for stable JSON output.
// When false, sorting is skipped for faster table generation.
var SortOutput = true

// Tables captures LR(1) parsing tables and productions.
type Tables struct {
	StartSymbol string                    `json:"start_symbol"`
	Actions     map[int]map[string]Action `json:"actions"`
	Gotos       map[int]map[string]int    `json:"gotos"`
	Productions []Production              `json:"productions"`
	Metadata    map[string]string         `json:"metadata,omitempty"`
	HintMode    string                    `json:"hint_mode,omitempty"`
}

type Action struct {
	Type   string `json:"type"`
	Target int    `json:"target,omitempty"`
}

type Production struct {
	LHS  string   `json:"lhs"`
	RHS  []Symbol `json:"rhs"`
	Hint *ASTHint `json:"hint,omitempty"`
}

// ASTHint captures AST-construction directives for a production.
type ASTHint struct {
	ParentIndex  int    `json:"parent"`
	ChildIndices []int  `json:"children"`
	NodeType     string `json:"type,omitempty"`
}

type Symbol struct {
	Name     string `json:"name"`
	Terminal bool   `json:"terminal"`
}

// MarshalJSON ensures deterministic map ordering for stable output.
// When SortOutput is false, falls back to standard JSON encoding for speed.
func (tables *Tables) MarshalJSON() ([]byte, error) {
	if tables == nil {
		return []byte("null"), nil
	}
	if !SortOutput {
		type tablesAlias Tables
		return json.Marshal((*tablesAlias)(tables))
	}
	var fields []jsonField

	startSymbolBytes, err := json.Marshal(tables.StartSymbol)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "start_symbol", value: startSymbolBytes})

	actionsBytes, err := marshalMapIntActionMap(tables.Actions)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "actions", value: actionsBytes})

	gotosBytes, err := marshalMapIntIntMap(tables.Gotos)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "gotos", value: gotosBytes})

	productionsBytes, err := json.Marshal(tables.Productions)
	if err != nil {
		return nil, err
	}
	fields = append(fields, jsonField{name: "productions", value: productionsBytes})

	if len(tables.Metadata) > 0 {
		metadataBytes, err := marshalMapStringString(tables.Metadata)
		if err != nil {
			return nil, err
		}
		fields = append(fields, jsonField{name: "metadata", value: metadataBytes})
	}

	if tables.HintMode != "" {
		hintModeBytes, err := json.Marshal(tables.HintMode)
		if err != nil {
			return nil, err
		}
		fields = append(fields, jsonField{name: "hint_mode", value: hintModeBytes})
	}

	return marshalOrderedFields(fields), nil
}

// GenerateTablesFromEBNF parses an EBNF grammar and produces LR(1) parser tables.
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
	lexerRuleSet := map[string]bool{}
	for _, name := range lexerRuleNames {
		lexerRuleSet[name] = true
	}

	parserRuleNames := selectParserRuleNames(ruleDefs)
	if len(parserRuleNames) == 0 {
		return nil, fmt.Errorf("no parser rules found")
	}

	startSymbol := selectStartSymbol(parserRuleNames)
	builder := newGrammarBuilder(parserRuleNames, lexerRuleSet)
	for _, rule := range ruleDefs {
		if !builder.parserRuleSet[rule.name] {
			continue
		}
		if err := builder.addRule(rule.name, rule.expr); err != nil {
			return nil, err
		}
	}

	if err := validateHints(builder.productions); err != nil {
		return nil, err
	}

	hintMode := ""
	for _, prod := range builder.productions {
		if prod.Hint != nil {
			hintMode = "hints"
			break
		}
	}

	grammar := newGrammar(builder, startSymbol)
	actions, gotos, err := buildLR1Tables(grammar)
	if err != nil {
		return nil, err
	}

	return &Tables{
		StartSymbol: startSymbol,
		Actions:     actions,
		Gotos:       gotos,
		Productions: grammar.productions,
		HintMode:    hintMode,
	}, nil
}

// EncodeTables returns pretty-printed JSON for tables.
func EncodeTables(tables *Tables) ([]byte, error) {
	return json.MarshalIndent(tables, "", "  ")
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

func marshalMapIntActionMap(m map[int]map[string]Action) ([]byte, error) {
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
		valueBytes, err := marshalMapStringAction(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalMapStringAction(m map[string]Action) ([]byte, error) {
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

func marshalMapIntIntMap(m map[int]map[string]int) ([]byte, error) {
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
		valueBytes, err := marshalMapStringInt(m[key])
		if err != nil {
			return nil, err
		}
		buf.Write(valueBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func marshalMapStringInt(m map[string]int) ([]byte, error) {
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

func selectParserRuleNames(ruleDefs []ruleDef) []string {
	var names []string
	for _, rule := range ruleDefs {
		if !isLexerRuleName(rule.name) {
			names = append(names, rule.name)
		}
	}
	return names
}

func selectStartSymbol(parserRuleNames []string) string {
	for _, name := range parserRuleNames {
		if name == "Root" {
			return name
		}
	}
	return parserRuleNames[0]
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

type grammarBuilder struct {
	parserRuleSet map[string]bool
	lexerRuleSet  map[string]bool
	usedNames     map[string]bool
	productions   []Production
	synthCounter  int
}

func newGrammarBuilder(parserRuleNames []string, lexerRuleSet map[string]bool) *grammarBuilder {
	parserRuleSet := map[string]bool{}
	usedNames := map[string]bool{}
	for _, name := range parserRuleNames {
		parserRuleSet[name] = true
		usedNames[name] = true
	}
	for name := range lexerRuleSet {
		usedNames[name] = true
	}
	return &grammarBuilder{
		parserRuleSet: parserRuleSet,
		lexerRuleSet:  lexerRuleSet,
		usedNames:     usedNames,
	}
}

type expandedAlternative struct {
	symbols []Symbol
	hint    *ASTHint
}

func (builder *grammarBuilder) addRule(ruleName string, expr *asts.ASTNode) error {
	alts, err := builder.expandExpr(expr)
	if err != nil {
		return fmt.Errorf("rule %q: %w", ruleName, err)
	}
	for _, alt := range alts {
		builder.productions = append(builder.productions, Production{
			LHS:  ruleName,
			RHS:  alt.symbols,
			Hint: alt.hint,
		})
	}
	return nil
}

func (builder *grammarBuilder) expandExpr(node *asts.ASTNode) ([]expandedAlternative, error) {
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
		return []expandedAlternative{{symbols: []Symbol{{Name: unquoted, Terminal: true}}}}, nil
	case parsers.EBNFParserNodeTypeRange:
		return nil, fmt.Errorf("range expressions are only allowed in lexer rules")
	case parsers.EBNFParserNodeTypeWildcard:
		return nil, fmt.Errorf("wildcard '.' is only allowed in lexer rules")
	case parsers.EBNFParserNodeTypeIdentifier:
		if node.Token == nil {
			return nil, fmt.Errorf("identifier node missing token")
		}
		identifier := string(node.Token.Lexeme)
		if builder.lexerRuleSet[identifier] {
			return []expandedAlternative{{symbols: []Symbol{{Name: identifier, Terminal: true}}}}, nil
		}
		if !builder.parserRuleSet[identifier] {
			return nil, fmt.Errorf("undefined rule %q", identifier)
		}
		return []expandedAlternative{{symbols: []Symbol{{Name: identifier, Terminal: false}}}}, nil
	case parsers.EBNFParserNodeTypeSequence:
		if len(node.Children) == 0 {
			return []expandedAlternative{{symbols: []Symbol{}}}, nil
		}
		alts := []expandedAlternative{{symbols: []Symbol{}}}
		for _, child := range node.Children {
			childAlts, err := builder.expandExpr(child)
			if err != nil {
				return nil, err
			}
			var next []expandedAlternative
			for _, alt := range alts {
				for _, childAlt := range childAlts {
					combined := make([]Symbol, 0, len(alt.symbols)+len(childAlt.symbols))
					combined = append(combined, alt.symbols...)
					combined = append(combined, childAlt.symbols...)
					next = append(next, expandedAlternative{symbols: combined})
				}
			}
			alts = next
		}
		return alts, nil
	case parsers.EBNFParserNodeTypeAlternates:
		var out []expandedAlternative
		for _, child := range node.Children {
			childAlts, err := builder.expandExpr(child)
			if err != nil {
				return nil, err
			}
			out = append(out, childAlts...)
		}
		return out, nil
	case parsers.EBNFParserNodeTypeOptional:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		childAlts, err := builder.expandExpr(node.Children[0])
		if err != nil {
			return nil, err
		}
		return append(childAlts, expandedAlternative{symbols: []Symbol{}}), nil
	case parsers.EBNFParserNodeTypeRepeat:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		childAlts, err := builder.expandExpr(node.Children[0])
		if err != nil {
			return nil, err
		}
		repeatName := builder.newSyntheticName("repeat")
		builder.parserRuleSet[repeatName] = true
		builder.usedNames[repeatName] = true
		builder.productions = append(builder.productions, Production{
			LHS: repeatName,
			RHS: []Symbol{},
		})
		for _, childAlt := range childAlts {
			if len(childAlt.symbols) == 0 {
				continue
			}
			combined := make([]Symbol, 0, len(childAlt.symbols)+1)
			combined = append(combined, childAlt.symbols...)
			combined = append(combined, Symbol{Name: repeatName, Terminal: false})
			builder.productions = append(builder.productions, Production{
				LHS: repeatName,
				RHS: combined,
			})
		}
		return []expandedAlternative{{symbols: []Symbol{{Name: repeatName, Terminal: false}}}}, nil
	case parsers.EBNFParserNodeTypeHintedSequence:
		if err := node.CheckArity(2); err != nil {
			return nil, err
		}
		seqNode := node.Children[0]
		hintNode := node.Children[1]

		alts, err := builder.expandExpr(seqNode)
		if err != nil {
			return nil, err
		}

		hint, err := parseHintNode(hintNode)
		if err != nil {
			return nil, err
		}

		for i := range alts {
			alts[i].hint = hint
		}
		return alts, nil
	default:
		return nil, fmt.Errorf("unsupported node type %q", node.Type)
	}
}

func parseHintNode(node *asts.ASTNode) (*ASTHint, error) {
	if node.Type != parsers.EBNFParserNodeTypeHint {
		return nil, fmt.Errorf("expected hint node, got %q", node.Type)
	}
	hint := &ASTHint{}
	hasParent := false
	hasChildren := false
	for _, field := range node.Children {
		if field.Type != parsers.EBNFParserNodeTypeHintField || field.Token == nil {
			return nil, fmt.Errorf("invalid hint field node")
		}
		key := string(field.Token.Lexeme)
		// Strip quotes from key
		unquoted, err := strconv.Unquote(key)
		if err != nil {
			return nil, fmt.Errorf("invalid hint key %q: %w", key, err)
		}
		if len(field.Children) != 1 {
			return nil, fmt.Errorf("hint field %q must have exactly one value", unquoted)
		}
		valueNode := field.Children[0]
		switch unquoted {
		case "parent":
			if valueNode.Type != parsers.EBNFParserNodeTypeHintInt || valueNode.Token == nil {
				return nil, fmt.Errorf("hint \"parent\" must be an integer")
			}
			val, err := strconv.Atoi(string(valueNode.Token.Lexeme))
			if err != nil {
				return nil, fmt.Errorf("invalid hint parent value: %w", err)
			}
			hint.ParentIndex = val
			hasParent = true
		case "children":
			if valueNode.Type != parsers.EBNFParserNodeTypeHintArray {
				return nil, fmt.Errorf("hint \"children\" must be an array")
			}
			indices := make([]int, 0, len(valueNode.Children))
			for _, elem := range valueNode.Children {
				if elem.Type != parsers.EBNFParserNodeTypeHintInt || elem.Token == nil {
					return nil, fmt.Errorf("hint children elements must be integers")
				}
				val, err := strconv.Atoi(string(elem.Token.Lexeme))
				if err != nil {
					return nil, fmt.Errorf("invalid hint child index: %w", err)
				}
				indices = append(indices, val)
			}
			hint.ChildIndices = indices
			hasChildren = true
		case "type":
			if valueNode.Type != parsers.EBNFParserNodeTypeHintString || valueNode.Token == nil {
				return nil, fmt.Errorf("hint \"type\" must be a string")
			}
			unquotedType, err := strconv.Unquote(string(valueNode.Token.Lexeme))
			if err != nil {
				return nil, fmt.Errorf("invalid hint type value: %w", err)
			}
			hint.NodeType = unquotedType
		default:
			return nil, fmt.Errorf("unknown hint field %q", unquoted)
		}
	}
	if !hasParent {
		return nil, fmt.Errorf("hint missing required \"parent\" field")
	}
	if !hasChildren {
		return nil, fmt.Errorf("hint missing required \"children\" field")
	}
	return hint, nil
}

func validateHints(productions []Production) error {
	hasAnyHint := false
	for _, prod := range productions {
		if prod.Hint != nil {
			hasAnyHint = true
			break
		}
	}
	if !hasAnyHint {
		return nil
	}
	for _, prod := range productions {
		if strings.HasPrefix(prod.LHS, "__pgpg_") {
			continue
		}
		if prod.Hint == nil {
			if len(prod.RHS) <= 1 {
				continue
			}
			rhsNames := make([]string, len(prod.RHS))
			for i, sym := range prod.RHS {
				rhsNames[i] = sym.Name
			}
			return fmt.Errorf(
				"production %s ::= %s has %d RHS symbols but no AST hint; "+
					"in hint mode, multi-element productions require hints",
				prod.LHS, strings.Join(rhsNames, " "), len(prod.RHS))
		}
		if prod.Hint.ParentIndex < 0 || prod.Hint.ParentIndex >= len(prod.RHS) {
			return fmt.Errorf("production %s: parent index %d out of range [0, %d)",
				prod.LHS, prod.Hint.ParentIndex, len(prod.RHS))
		}
		for _, ci := range prod.Hint.ChildIndices {
			if ci < 0 || ci >= len(prod.RHS) {
				return fmt.Errorf("production %s: child index %d out of range [0, %d)",
					prod.LHS, ci, len(prod.RHS))
			}
		}
	}
	return nil
}

func (builder *grammarBuilder) newSyntheticName(kind string) string {
	for {
		builder.synthCounter++
		name := fmt.Sprintf("__pgpg_%s_%d", kind, builder.synthCounter)
		if !builder.usedNames[name] {
			builder.usedNames[name] = true
			return name
		}
	}
}

type grammar struct {
	startSymbol string
	productions []Production
	byLHS       map[string][]int
	terminals   map[string]bool
	nonterms    map[string]bool
}

func newGrammar(builder *grammarBuilder, startSymbol string) *grammar {
	productions := append([]Production{}, builder.productions...)
	byLHS := map[string][]int{}
	nonterms := map[string]bool{}
	terminals := map[string]bool{}
	for i, prod := range productions {
		byLHS[prod.LHS] = append(byLHS[prod.LHS], i)
		nonterms[prod.LHS] = true
		for _, sym := range prod.RHS {
			if sym.Terminal {
				terminals[sym.Name] = true
			} else {
				nonterms[sym.Name] = true
			}
		}
	}

	augmented := builder.newSyntheticName("start")
	productions = append([]Production{{
		LHS: augmented,
		RHS: []Symbol{{Name: startSymbol, Terminal: false}},
	}}, productions...)

	byLHS = map[string][]int{}
	nonterms = map[string]bool{}
	terminals = map[string]bool{}
	for i, prod := range productions {
		byLHS[prod.LHS] = append(byLHS[prod.LHS], i)
		nonterms[prod.LHS] = true
		for _, sym := range prod.RHS {
			if sym.Terminal {
				terminals[sym.Name] = true
			} else {
				nonterms[sym.Name] = true
			}
		}
	}
	terminals[eofSymbol] = true

	return &grammar{
		startSymbol: augmented,
		productions: productions,
		byLHS:       byLHS,
		terminals:   terminals,
		nonterms:    nonterms,
	}
}

type item struct {
	prod      int
	dot       int
	lookahead string
}

func buildLR1Tables(grammar *grammar) (map[int]map[string]Action, map[int]map[string]int, error) {
	first := computeFirstSets(grammar)
	stateMap := map[uint64][]int{} // hash → state IDs (for collision resolution)
	var states []map[item]struct{}
	stateLabels := map[int]string{}
	var queue []int

	startItem := item{prod: 0, dot: 0, lookahead: eofSymbol}
	startSet := closure(grammar, first, map[item]struct{}{startItem: {}})
	startHash := itemSetHash(startSet)
	stateMap[startHash] = []int{0}
	states = append(states, startSet)
	stateLabels[0] = stateLabel(startSet, grammar)
	queue = append(queue, 0)

	actions := map[int]map[string]Action{}
	gotos := map[int]map[string]int{}

	for len(queue) > 0 {
		stateID := queue[0]
		queue = queue[1:]
		itemSet := states[stateID]

		transitions := map[Symbol]map[item]struct{}{}
		orderedItems := sortedItems(itemSet)
		for _, it := range orderedItems {
			prod := grammar.productions[it.prod]
			if it.dot >= len(prod.RHS) {
				continue
			}
			nextSym := prod.RHS[it.dot]
			nextItem := item{prod: it.prod, dot: it.dot + 1, lookahead: it.lookahead}
			set := transitions[nextSym]
			if set == nil {
				set = map[item]struct{}{}
				transitions[nextSym] = set
			}
			set[nextItem] = struct{}{}
		}

		for _, sym := range sortedSymbols(transitions) {
			seedSet := transitions[sym]
			gotoSet := closure(grammar, first, seedSet)
			hash := itemSetHash(gotoSet)
			target := -1
			for _, id := range stateMap[hash] {
				if itemSetsEqual(states[id], gotoSet) {
					target = id
					break
				}
			}
			if target < 0 {
				target = len(states)
				stateMap[hash] = append(stateMap[hash], target)
				states = append(states, gotoSet)
				stateLabels[target] = stateLabel(gotoSet, grammar)
				queue = append(queue, target)
			}
			if sym.Terminal {
				if err := setAction(actions, stateID, sym.Name, Action{Type: "shift", Target: target}, itemSet, grammar, stateLabels); err != nil {
					return nil, nil, err
				}
			} else {
				if gotos[stateID] == nil {
					gotos[stateID] = map[string]int{}
				}
				if existing, ok := gotos[stateID][sym.Name]; ok && existing != target {
					return nil, nil, fmt.Errorf("goto conflict in state %d on %q", stateID, sym.Name)
				}
				gotos[stateID][sym.Name] = target
			}
		}

		for _, it := range orderedItems {
			prod := grammar.productions[it.prod]
			if it.dot < len(prod.RHS) {
				continue
			}
			if it.prod == 0 && it.lookahead == eofSymbol {
				if err := setAction(actions, stateID, eofSymbol, Action{Type: "accept"}, itemSet, grammar, stateLabels); err != nil {
					return nil, nil, err
				}
				continue
			}
			if err := setAction(actions, stateID, it.lookahead, Action{Type: "reduce", Target: it.prod}, itemSet, grammar, stateLabels); err != nil {
				return nil, nil, err
			}
		}
	}

	return actions, gotos, nil
}

func setAction(actions map[int]map[string]Action, state int, terminal string, action Action, itemSet map[item]struct{}, grammar *grammar, stateLabels map[int]string) error {
	if actions[state] == nil {
		actions[state] = map[string]Action{}
	}
	if existing, ok := actions[state][terminal]; ok {
		if existing.Type != action.Type || existing.Target != action.Target {
			return conflictError(state, terminal, existing, action, itemSet, grammar, stateLabels)
		}
		return nil
	}
	actions[state][terminal] = action
	return nil
}

func conflictError(state int, terminal string, existing Action, next Action, itemSet map[item]struct{}, grammar *grammar, stateLabels map[int]string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "Action conflict\n  State: %s\n  Lookahead: %q\n\n", formatStateLabel(state, stateLabels), terminal)
	b.WriteString(formatAction("Existing", existing, grammar, stateLabels))
	b.WriteString(formatAction("New", next, grammar, stateLabels))
	if grammar != nil && itemSet != nil {
		b.WriteString("\n  Items in state:\n")
		for _, it := range sortedItems(itemSet) {
			b.WriteString("    ")
			b.WriteString(formatItem(it, grammar))
			b.WriteByte('\n')
		}
	}
	if hint := buildConflictHint(existing, next, grammar); hint != "" {
		b.WriteString("\n  Hint:\n")
		b.WriteString(hint)
	}
	return fmt.Errorf(b.String())
}

func formatAction(label string, action Action, grammar *grammar, stateLabels map[int]string) string {
	switch action.Type {
	case "shift":
		return fmt.Sprintf("  %s action: shift to state %s\n", label, formatStateLabel(action.Target, stateLabels))
	case "reduce":
		prod := formatProduction(action.Target, grammar)
		return fmt.Sprintf("  %s action: reduce by production %d: %s\n", label, action.Target, prod)
	case "accept":
		return fmt.Sprintf("  %s action: accept\n", label)
	default:
		return fmt.Sprintf("  %s action: %s\n", label, action.Type)
	}
}

func formatProduction(prodIndex int, grammar *grammar) string {
	if grammar == nil || prodIndex < 0 || prodIndex >= len(grammar.productions) {
		return "<unknown production>"
	}
	prod := grammar.productions[prodIndex]
	var rhs []string
	for _, sym := range prod.RHS {
		rhs = append(rhs, sym.Name)
	}
	return fmt.Sprintf("%s ::= %s", prod.LHS, strings.Join(rhs, " "))
}

func formatItem(it item, grammar *grammar) string {
	if grammar == nil || it.prod < 0 || it.prod >= len(grammar.productions) {
		return fmt.Sprintf("<?> (prod=%d dot=%d), lookahead=%s", it.prod, it.dot, it.lookahead)
	}
	prod := grammar.productions[it.prod]
	var rhs []string
	for i, sym := range prod.RHS {
		if i == it.dot {
			rhs = append(rhs, ".")
		}
		rhs = append(rhs, sym.Name)
	}
	if it.dot >= len(prod.RHS) {
		rhs = append(rhs, ".")
	}
	return fmt.Sprintf("%s ::= %s , lookahead=%s", prod.LHS, strings.Join(rhs, " "), it.lookahead)
}

func stateLabel(itemSet map[item]struct{}, grammar *grammar) string {
	if grammar == nil || itemSet == nil {
		return ""
	}
	items := sortedItems(itemSet)
	for _, it := range items {
		if it.dot > 0 {
			return formatItem(it, grammar)
		}
	}
	if len(items) > 0 {
		return formatItem(items[0], grammar)
	}
	return ""
}

func formatStateLabel(state int, stateLabels map[int]string) string {
	if stateLabels == nil {
		return fmt.Sprintf("%d", state)
	}
	label := stateLabels[state]
	if label == "" {
		return fmt.Sprintf("%d", state)
	}
	return fmt.Sprintf("%d (%s)", state, label)
}

func buildConflictHint(existing Action, next Action, grammar *grammar) string {
	if grammar == nil {
		return ""
	}
	var hints []string
	if existing.Type == "reduce" {
		hints = append(hints, reduceHint(existing.Target, grammar)...)
	}
	if next.Type == "reduce" {
		hints = append(hints, reduceHint(next.Target, grammar)...)
	}
	if existing.Type == "shift" || next.Type == "shift" {
		hints = append(hints, "- Shift/reduce conflicts often come from ambiguous operator precedence or unintended recursion")
	}
	if len(hints) == 0 {
		return ""
	}
	return strings.Join(hints, "\n") + "\n"
}

func reduceHint(prodIndex int, grammar *grammar) []string {
	prod, ok := productionAt(prodIndex, grammar)
	if !ok {
		return nil
	}
	var hints []string
	userStart := userStartSymbol(grammar)
	if userStart != "" && containsSymbol(prod.RHS, userStart) {
		hints = append(hints, fmt.Sprintf("- Production reduces to %s via %s; check for cycles involving the start symbol", prod.LHS, userStart))
	}
	if containsSymbol(prod.RHS, prod.LHS) {
		hints = append(hints, fmt.Sprintf("- Production %s ::= ... %s ... is directly recursive; verify it appears only where intended", prod.LHS, prod.LHS))
	}
	return hints
}

func productionAt(index int, grammar *grammar) (Production, bool) {
	if grammar == nil || index < 0 || index >= len(grammar.productions) {
		return Production{}, false
	}
	return grammar.productions[index], true
}

func userStartSymbol(grammar *grammar) string {
	if grammar == nil || len(grammar.productions) == 0 || len(grammar.productions[0].RHS) == 0 {
		return ""
	}
	return grammar.productions[0].RHS[0].Name
}

func containsSymbol(symbols []Symbol, name string) bool {
	for _, sym := range symbols {
		if sym.Name == name {
			return true
		}
	}
	return false
}

func closure(grammar *grammar, first *firstSets, items map[item]struct{}) map[item]struct{} {
	queue := make([]item, 0, len(items))
	for _, it := range sortedItems(items) {
		queue = append(queue, it)
	}

	for len(queue) > 0 {
		it := queue[0]
		queue = queue[1:]
		prod := grammar.productions[it.prod]
		if it.dot >= len(prod.RHS) {
			continue
		}
		next := prod.RHS[it.dot]
		if next.Terminal {
			continue
		}
		beta := prod.RHS[it.dot+1:]
		lookaheads := firstOfSequence(grammar, first, beta, it.lookahead)
		for _, la := range lookaheads {
			for _, prodIndex := range grammar.byLHS[next.Name] {
				newItem := item{prod: prodIndex, dot: 0, lookahead: la}
				if _, ok := items[newItem]; ok {
					continue
				}
				items[newItem] = struct{}{}
				queue = append(queue, newItem)
			}
		}
	}
	return items
}

func sortedItems(items map[item]struct{}) []item {
	out := make([]item, 0, len(items))
	for it := range items {
		out = append(out, it)
	}
	if SortOutput {
		sort.Slice(out, func(i, j int) bool {
			if out[i].prod != out[j].prod {
				return out[i].prod < out[j].prod
			}
			if out[i].dot != out[j].dot {
				return out[i].dot < out[j].dot
			}
			return out[i].lookahead < out[j].lookahead
		})
	}
	return out
}

func sortedSymbols(transitions map[Symbol]map[item]struct{}) []Symbol {
	syms := make([]Symbol, 0, len(transitions))
	for sym := range transitions {
		syms = append(syms, sym)
	}
	if SortOutput {
		sort.Slice(syms, func(i, j int) bool {
			if syms[i].Terminal != syms[j].Terminal {
				return !syms[i].Terminal && syms[j].Terminal
			}
			if syms[i].Name != syms[j].Name {
				return syms[i].Name < syms[j].Name
			}
			return false
		})
	}
	return syms
}

// itemSetHash computes an order-independent hash of an item set.
// Uses commutative addition so iteration order doesn't matter.
func itemSetHash(items map[item]struct{}) uint64 {
	var h uint64
	for it := range items {
		ih := uint64(it.prod)*2654435761 ^ uint64(it.dot)*40503
		for i := 0; i < len(it.lookahead); i++ {
			ih = ih*1099511628211 ^ uint64(it.lookahead[i])
		}
		h += ih
	}
	return h
}

// itemSetsEqual checks whether two item sets contain the same items.
func itemSetsEqual(a, b map[item]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for it := range a {
		if _, ok := b[it]; !ok {
			return false
		}
	}
	return true
}

type firstSets struct {
	terminals map[string]map[string]bool
	nullable  map[string]bool
}

func computeFirstSets(grammar *grammar) *firstSets {
	terminals := map[string]map[string]bool{}
	nullable := map[string]bool{}

	for term := range grammar.terminals {
		terminals[term] = map[string]bool{term: true}
	}
	for nonterm := range grammar.nonterms {
		if terminals[nonterm] == nil {
			terminals[nonterm] = map[string]bool{}
		}
	}

	changed := true
	for changed {
		changed = false
		for _, prod := range grammar.productions {
			if len(prod.RHS) == 0 {
				if !nullable[prod.LHS] {
					nullable[prod.LHS] = true
					changed = true
				}
				continue
			}
			allNullable := true
			for _, sym := range prod.RHS {
				if sym.Terminal {
					if !terminals[prod.LHS][sym.Name] {
						terminals[prod.LHS][sym.Name] = true
						changed = true
					}
					allNullable = false
					break
				}
				for term := range terminals[sym.Name] {
					if term == "" {
						continue
					}
					if !terminals[prod.LHS][term] {
						terminals[prod.LHS][term] = true
						changed = true
					}
				}
				if !nullable[sym.Name] {
					allNullable = false
					break
				}
			}
			if allNullable && !nullable[prod.LHS] {
				nullable[prod.LHS] = true
				changed = true
			}
		}
	}

	return &firstSets{terminals: terminals, nullable: nullable}
}

func firstOfSequence(grammar *grammar, first *firstSets, seq []Symbol, lookahead string) []string {
	out := map[string]bool{}
	if len(seq) == 0 {
		out[lookahead] = true
		return sortedKeys(out)
	}
	for _, sym := range seq {
		if sym.Terminal {
			out[sym.Name] = true
			return sortedKeys(out)
		}
		for term := range first.terminals[sym.Name] {
			if term == "" {
				continue
			}
			out[term] = true
		}
		if !first.nullable[sym.Name] {
			return sortedKeys(out)
		}
	}
	out[lookahead] = true
	return sortedKeys(out)
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	if SortOutput {
		sort.Strings(keys)
	}
	return keys
}
