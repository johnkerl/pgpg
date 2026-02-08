package parsegen

import (
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

// Tables captures LR(1) parsing tables and productions.
type Tables struct {
	StartSymbol string                 `json:"start_symbol"`
	Actions     map[int]map[string]Action `json:"actions"`
	Gotos       map[int]map[string]int `json:"gotos"`
	Productions []Production           `json:"productions"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

type Action struct {
	Type   string `json:"type"`
	Target int    `json:"target,omitempty"`
}

type Production struct {
	LHS string   `json:"lhs"`
	RHS []Symbol `json:"rhs"`
}

type Symbol struct {
	Name     string `json:"name"`
	Terminal bool   `json:"terminal"`
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
	}, nil
}

// EncodeTables returns pretty-printed JSON for tables.
func EncodeTables(tables *Tables) ([]byte, error) {
	return json.MarshalIndent(tables, "", "  ")
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

func (builder *grammarBuilder) addRule(ruleName string, expr *asts.ASTNode) error {
	sequences, err := builder.expandExpr(expr)
	if err != nil {
		return fmt.Errorf("rule %q: %w", ruleName, err)
	}
	for _, seq := range sequences {
		builder.productions = append(builder.productions, Production{
			LHS: ruleName,
			RHS: seq,
		})
	}
	return nil
}

func (builder *grammarBuilder) expandExpr(node *asts.ASTNode) ([][]Symbol, error) {
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
		return [][]Symbol{{{Name: unquoted, Terminal: true}}}, nil
	case parsers.EBNFParserNodeTypeIdentifier:
		if node.Token == nil {
			return nil, fmt.Errorf("identifier node missing token")
		}
		identifier := string(node.Token.Lexeme)
		if builder.lexerRuleSet[identifier] {
			return [][]Symbol{{{Name: identifier, Terminal: true}}}, nil
		}
		if !builder.parserRuleSet[identifier] {
			return nil, fmt.Errorf("undefined rule %q", identifier)
		}
		return [][]Symbol{{{Name: identifier, Terminal: false}}}, nil
	case parsers.EBNFParserNodeTypeSequence:
		if len(node.Children) == 0 {
			return [][]Symbol{{}}, nil
		}
		sequences := [][]Symbol{{}}
		for _, child := range node.Children {
			childSeqs, err := builder.expandExpr(child)
			if err != nil {
				return nil, err
			}
			var next [][]Symbol
			for _, seq := range sequences {
				for _, childSeq := range childSeqs {
					combined := make([]Symbol, 0, len(seq)+len(childSeq))
					combined = append(combined, seq...)
					combined = append(combined, childSeq...)
					next = append(next, combined)
				}
			}
			sequences = next
		}
		return sequences, nil
	case parsers.EBNFParserNodeTypeAlternates:
		var out [][]Symbol
		for _, child := range node.Children {
			childSeqs, err := builder.expandExpr(child)
			if err != nil {
				return nil, err
			}
			out = append(out, childSeqs...)
		}
		return out, nil
	case parsers.EBNFParserNodeTypeOptional:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		childSeqs, err := builder.expandExpr(node.Children[0])
		if err != nil {
			return nil, err
		}
		return append(childSeqs, []Symbol{}), nil
	case parsers.EBNFParserNodeTypeRepeat:
		if err := node.CheckArity(1); err != nil {
			return nil, err
		}
		childSeqs, err := builder.expandExpr(node.Children[0])
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
		for _, childSeq := range childSeqs {
			if len(childSeq) == 0 {
				continue
			}
			combined := make([]Symbol, 0, len(childSeq)+1)
			combined = append(combined, childSeq...)
			combined = append(combined, Symbol{Name: repeatName, Terminal: false})
			builder.productions = append(builder.productions, Production{
				LHS: repeatName,
				RHS: combined,
			})
		}
		return [][]Symbol{{{Name: repeatName, Terminal: false}}}, nil
	default:
		return nil, fmt.Errorf("unsupported node type %q", node.Type)
	}
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

func itemKey(it item) string {
	return fmt.Sprintf("%d:%d:%s", it.prod, it.dot, it.lookahead)
}

func buildLR1Tables(grammar *grammar) (map[int]map[string]Action, map[int]map[string]int, error) {
	first := computeFirstSets(grammar)
	stateMap := map[string]int{}
	var states []map[string]item
	var queue []int

	startItem := item{prod: 0, dot: 0, lookahead: eofSymbol}
	startSet := closure(grammar, first, map[string]item{itemKey(startItem): startItem})
	startKey := itemSetKey(startSet)
	stateMap[startKey] = 0
	states = append(states, startSet)
	queue = append(queue, 0)

	actions := map[int]map[string]Action{}
	gotos := map[int]map[string]int{}

	for len(queue) > 0 {
		stateID := queue[0]
		queue = queue[1:]
		itemSet := states[stateID]

		transitions := map[Symbol]map[string]item{}
		for _, it := range itemSet {
			prod := grammar.productions[it.prod]
			if it.dot >= len(prod.RHS) {
				continue
			}
			nextSym := prod.RHS[it.dot]
			nextItem := item{prod: it.prod, dot: it.dot + 1, lookahead: it.lookahead}
			set := transitions[nextSym]
			if set == nil {
				set = map[string]item{}
				transitions[nextSym] = set
			}
			set[itemKey(nextItem)] = nextItem
		}

		for sym, seedSet := range transitions {
			gotoSet := closure(grammar, first, seedSet)
			key := itemSetKey(gotoSet)
			target, ok := stateMap[key]
			if !ok {
				target = len(states)
				stateMap[key] = target
				states = append(states, gotoSet)
				queue = append(queue, target)
			}
			if sym.Terminal {
				if err := setAction(actions, stateID, sym.Name, Action{Type: "shift", Target: target}); err != nil {
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

		for _, it := range itemSet {
			prod := grammar.productions[it.prod]
			if it.dot < len(prod.RHS) {
				continue
			}
			if it.prod == 0 && it.lookahead == eofSymbol {
				if err := setAction(actions, stateID, eofSymbol, Action{Type: "accept"}); err != nil {
					return nil, nil, err
				}
				continue
			}
			if err := setAction(actions, stateID, it.lookahead, Action{Type: "reduce", Target: it.prod}); err != nil {
				return nil, nil, err
			}
		}
	}

	return actions, gotos, nil
}

func setAction(actions map[int]map[string]Action, state int, terminal string, action Action) error {
	if actions[state] == nil {
		actions[state] = map[string]Action{}
	}
	if existing, ok := actions[state][terminal]; ok {
		if existing.Type != action.Type || existing.Target != action.Target {
			return fmt.Errorf("action conflict in state %d on %q", state, terminal)
		}
		return nil
	}
	actions[state][terminal] = action
	return nil
}

func closure(grammar *grammar, first *firstSets, items map[string]item) map[string]item {
	queue := make([]item, 0, len(items))
	for _, it := range items {
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
				key := itemKey(newItem)
				if _, ok := items[key]; ok {
					continue
				}
				items[key] = newItem
				queue = append(queue, newItem)
			}
		}
	}
	return items
}

func itemSetKey(items map[string]item) string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
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
	sort.Strings(keys)
	return keys
}
