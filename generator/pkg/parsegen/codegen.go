package parsegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// DecodeTables reads tables JSON into Tables.
func DecodeTables(data []byte) (*Tables, error) {
	var tables Tables
	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	}
	return &tables, nil
}

// GenerateGoParserCode creates Go source implementing an LR(1) parser from tables.
func GenerateGoParserCode(tables *Tables, packageName string, typeName string) ([]byte, error) {
	if tables == nil {
		return nil, fmt.Errorf("nil tables")
	}
	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if typeName == "" {
		return nil, fmt.Errorf("type name is required")
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n\n")
	buf.WriteString("\tmanuallexers \"github.com/johnkerl/pgpg/manual/pkg/lexers\"\n")
	buf.WriteString("\t\"github.com/johnkerl/pgpg/manual/pkg/asts\"\n")
	buf.WriteString("\t\"github.com/johnkerl/pgpg/manual/pkg/tokens\"\n")
	buf.WriteString(")\n\n")

	fmt.Fprintf(&buf, "type %s struct {}\n\n", typeName)
	fmt.Fprintf(&buf, "func New%s() *%s { return &%s{} }\n\n", typeName, typeName, typeName)

	buf.WriteString("func (parser *" + typeName + ") Parse(lexer manuallexers.AbstractLexer) (*asts.AST, error) {\n")
	buf.WriteString("\tif lexer == nil {\n")
	buf.WriteString("\t\treturn nil, fmt.Errorf(\"parser: nil lexer\")\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tstateStack := []int{0}\n")
	buf.WriteString("\tnodeStack := []*asts.ASTNode{}\n")
	buf.WriteString("\tlookahead := lexer.Scan()\n")
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tif lookahead == nil {\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"parser: lexer returned nil token\")\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif lookahead.Type == tokens.TokenTypeError {\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"lexer error: %s\", string(lookahead.Lexeme))\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tstate := stateStack[len(stateStack)-1]\n")
	buf.WriteString("\t\taction, ok := " + typeName + "Actions[state][lookahead.Type]\n")
	buf.WriteString("\t\tif !ok {\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"parse error: unexpected %s (%q)\", lookahead.Type, string(lookahead.Lexeme))\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tswitch action.kind {\n")
	buf.WriteString("\t\tcase " + typeName + "ActionShift:\n")
	buf.WriteString("\t\t\tnodeStack = append(nodeStack, asts.NewASTNodeTerminal(lookahead, asts.NodeType(lookahead.Type)))\n")
	buf.WriteString("\t\t\tstateStack = append(stateStack, action.target)\n")
	buf.WriteString("\t\t\tlookahead = lexer.Scan()\n")
	buf.WriteString("\t\tcase " + typeName + "ActionReduce:\n")
	buf.WriteString("\t\t\tprod := " + typeName + "Productions[action.target]\n")
	buf.WriteString("\t\t\tchildren := make([]*asts.ASTNode, prod.rhsCount)\n")
	buf.WriteString("\t\t\tfor i := prod.rhsCount - 1; i >= 0; i-- {\n")
	buf.WriteString("\t\t\t\tstateStack = stateStack[:len(stateStack)-1]\n")
	buf.WriteString("\t\t\t\tchildren[i] = nodeStack[len(nodeStack)-1]\n")
	buf.WriteString("\t\t\t\tnodeStack = nodeStack[:len(nodeStack)-1]\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tif prod.rhsCount == 0 {\n")
	buf.WriteString("\t\t\t\tchildren = []*asts.ASTNode{}\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tnode := asts.NewASTNode(nil, prod.lhs, children)\n")
	buf.WriteString("\t\t\tnodeStack = append(nodeStack, node)\n")
	buf.WriteString("\t\t\tstate = stateStack[len(stateStack)-1]\n")
	buf.WriteString("\t\t\tnextState, ok := " + typeName + "Gotos[state][prod.lhs]\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\treturn nil, fmt.Errorf(\"parse error: missing goto for %s\", prod.lhs)\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tstateStack = append(stateStack, nextState)\n")
	buf.WriteString("\t\tcase " + typeName + "ActionAccept:\n")
	buf.WriteString("\t\t\tif len(nodeStack) != 1 {\n")
	buf.WriteString("\t\t\t\treturn nil, fmt.Errorf(\"parse error: unexpected parse stack size %d\", len(nodeStack))\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\treturn asts.NewAST(nodeStack[0]), nil\n")
	buf.WriteString("\t\tdefault:\n")
	buf.WriteString("\t\t\treturn nil, fmt.Errorf(\"parse error: no action\")\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString("type " + typeName + "ActionKind int\n\n")
	buf.WriteString("const (\n")
	buf.WriteString("\t" + typeName + "ActionShift " + typeName + "ActionKind = iota\n")
	buf.WriteString("\t" + typeName + "ActionReduce\n")
	buf.WriteString("\t" + typeName + "ActionAccept\n")
	buf.WriteString(")\n\n")

	buf.WriteString("type " + typeName + "Action struct {\n")
	buf.WriteString("\tkind   " + typeName + "ActionKind\n")
	buf.WriteString("\ttarget int\n")
	buf.WriteString("}\n\n")

	buf.WriteString("type " + typeName + "Production struct {\n")
	buf.WriteString("\tlhs      asts.NodeType\n")
	buf.WriteString("\trhsCount int\n")
	buf.WriteString("}\n\n")

	buf.WriteString("var " + typeName + "Actions = map[int]map[tokens.TokenType]" + typeName + "Action{\n")
	writeActions(&buf, tables, typeName)
	buf.WriteString("}\n\n")

	buf.WriteString("var " + typeName + "Gotos = map[int]map[asts.NodeType]int{\n")
	writeGotos(&buf, tables)
	buf.WriteString("}\n\n")

	buf.WriteString("var " + typeName + "Productions = []" + typeName + "Production{\n")
	writeProductions(&buf, tables)
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

func writeActions(buf *bytes.Buffer, tables *Tables, typeName string) {
	stateIDs := make([]int, 0, len(tables.Actions))
	for state := range tables.Actions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)
	for _, state := range stateIDs {
		buf.WriteString(fmt.Sprintf("\t%d: {\n", state))
		entries := tables.Actions[state]
		terms := make([]string, 0, len(entries))
		for term := range entries {
			terms = append(terms, term)
		}
		sort.Strings(terms)
		for _, term := range terms {
			action := entries[term]
			buf.WriteString("\t\t")
			buf.WriteString(tokenTypeLiteral(term))
			buf.WriteString(": {kind: " + actionKindLiteral(action.Type, typeName))
			if action.Type == "shift" || action.Type == "reduce" {
				buf.WriteString(fmt.Sprintf(", target: %d", action.Target))
			}
			buf.WriteString("},\n")
		}
		buf.WriteString("\t},\n")
	}
}

func writeGotos(buf *bytes.Buffer, tables *Tables) {
	stateIDs := make([]int, 0, len(tables.Gotos))
	for state := range tables.Gotos {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)
	for _, state := range stateIDs {
		buf.WriteString(fmt.Sprintf("\t%d: {\n", state))
		entries := tables.Gotos[state]
		names := make([]string, 0, len(entries))
		for name := range entries {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			buf.WriteString("\t\tasts.NodeType(" + strconv.Quote(name) + "): ")
			buf.WriteString(fmt.Sprintf("%d", entries[name]))
			buf.WriteString(",\n")
		}
		buf.WriteString("\t},\n")
	}
}

func writeProductions(buf *bytes.Buffer, tables *Tables) {
	for _, prod := range tables.Productions {
		buf.WriteString("\t{lhs: asts.NodeType(" + strconv.Quote(prod.LHS) + "), rhsCount: ")
		buf.WriteString(fmt.Sprintf("%d", len(prod.RHS)))
		buf.WriteString("},\n")
	}
}

func actionKindLiteral(kind string, typeName string) string {
	switch kind {
	case "shift":
		return typeName + "ActionShift"
	case "reduce":
		return typeName + "ActionReduce"
	case "accept":
		return typeName + "ActionAccept"
	default:
		return typeName + "ActionAccept"
	}
}

func tokenTypeLiteral(term string) string {
	if term == eofSymbol {
		return "tokens.TokenTypeEOF"
	}
	return "tokens.TokenType(" + strconv.Quote(term) + ")"
}
