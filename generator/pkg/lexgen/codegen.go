package lexgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DecodeTables reads tables JSON into Tables.
func DecodeTables(data []byte) (*Tables, error) {
	var tables Tables
	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	}
	return &tables, nil
}

// GenerateGoLexerCode creates Go source implementing a lexer from tables.
func GenerateGoLexerCode(tables *Tables, packageName string, typeName string) ([]byte, error) {
	if tables == nil {
		return nil, fmt.Errorf("nil tables")
	}
	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if typeName == "" {
		return nil, fmt.Errorf("type name is required")
	}

	hasIgnoredActions := false
	for _, action := range tables.Actions {
		if strings.HasPrefix(action, "!") {
			hasIgnoredActions = true
			break
		}
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	if hasIgnoredActions {
		buf.WriteString("\t\"strings\"\n")
	}
	buf.WriteString("\t\"unicode/utf8\"\n\n")
	buf.WriteString("\tmanuallexers \"github.com/johnkerl/pgpg/manual/pkg/lexers\"\n")
	buf.WriteString("\t\"github.com/johnkerl/pgpg/manual/pkg/tokens\"\n")
	buf.WriteString(")\n\n")

	fmt.Fprintf(&buf, "type %s struct {\n", typeName)
	buf.WriteString("\tinputText     string\n")
	buf.WriteString("\tinputLength   int\n")
	buf.WriteString("\ttokenLocation *tokens.TokenLocation\n")
	buf.WriteString("}\n\n")

	fmt.Fprintf(&buf, "var _ manuallexers.AbstractLexer = (*%s)(nil)\n\n", typeName)

	fmt.Fprintf(&buf, "func New%s(inputText string) manuallexers.AbstractLexer {\n", typeName)
	fmt.Fprintf(&buf, "\treturn &%s{\n", typeName)
	buf.WriteString("\t\tinputText:     inputText,\n")
	buf.WriteString("\t\tinputLength:   len(inputText),\n")
	buf.WriteString("\t\ttokenLocation: tokens.NewTokenLocation(),\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (lexer *" + typeName + ") Scan() *tokens.Token {\n")
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tif lexer.tokenLocation.ByteOffset >= lexer.inputLength {\n")
	buf.WriteString("\t\t\treturn tokens.NewEOFToken(lexer.tokenLocation)\n")
	buf.WriteString("\t\t}\n\n")
	buf.WriteString("\t\tstartLocation := *lexer.tokenLocation\n")
	buf.WriteString("\t\tscanLocation := *lexer.tokenLocation\n")
	buf.WriteString("\t\tstate := " + typeName + "StartState\n")
	buf.WriteString("\t\tlastAcceptState := -1\n")
	buf.WriteString("\t\tlastAcceptLocation := scanLocation\n\n")
	buf.WriteString("\t\tfor {\n")
	buf.WriteString("\t\t\tif scanLocation.ByteOffset >= lexer.inputLength {\n")
	buf.WriteString("\t\t\t\tbreak\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tr, width := lexer.peekRuneAt(scanLocation.ByteOffset)\n")
	buf.WriteString("\t\t\tnextState, ok := " + typeName + "LookupTransition(state, r)\n")
	buf.WriteString("\t\t\tif !ok {\n")
	buf.WriteString("\t\t\t\tbreak\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t\tscanLocation.LocateRune(r, width)\n")
	buf.WriteString("\t\t\tstate = nextState\n")
	buf.WriteString("\t\t\tif _, ok := " + typeName + "Actions[state]; ok {\n")
	buf.WriteString("\t\t\t\tlastAcceptState = state\n")
	buf.WriteString("\t\t\t\tlastAcceptLocation = scanLocation\n")
	buf.WriteString("\t\t\t}\n")
	buf.WriteString("\t\t}\n\n")
	buf.WriteString("\t\tif lastAcceptState < 0 {\n")
	buf.WriteString("\t\t\tr, _ := lexer.peekRuneAt(lexer.tokenLocation.ByteOffset)\n")
	buf.WriteString("\t\t\treturn tokens.NewErrorToken(fmt.Sprintf(\"lexer: unrecognized input %q\", r), lexer.tokenLocation)\n")
	buf.WriteString("\t\t}\n\n")
	buf.WriteString("\t\tlexemeText := lexer.inputText[lexer.tokenLocation.ByteOffset:lastAcceptLocation.ByteOffset]\n")
	buf.WriteString("\t\tlexeme := []rune(lexemeText)\n")
	buf.WriteString("\t\t*lexer.tokenLocation = lastAcceptLocation\n")
	buf.WriteString("\t\ttokenType := " + typeName + "Actions[lastAcceptState]\n")
	if hasIgnoredActions {
		buf.WriteString("\t\tif " + typeName + "IsIgnoredToken(tokenType) {\n")
		buf.WriteString("\t\t\tcontinue\n")
		buf.WriteString("\t\t}\n")
	}
	buf.WriteString("\t\treturn tokens.NewToken(lexeme, tokenType, &startLocation)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (lexer *" + typeName + ") peekRuneAt(byteOffset int) (rune, int) {\n")
	buf.WriteString("\tr, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])\n")
	buf.WriteString("\treturn r, width\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func " + typeName + "LookupTransition(state int, r rune) (int, bool) {\n")
	buf.WriteString("\ttransitionsForState, ok := " + typeName + "Transitions[state]\n")
	buf.WriteString("\tif !ok {\n")
	buf.WriteString("\t\treturn 0, false\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tfor _, tr := range transitionsForState {\n")
	buf.WriteString("\t\tif r < tr.from {\n")
	buf.WriteString("\t\t\treturn 0, false\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tif r >= tr.from && r <= tr.to {\n")
	buf.WriteString("\t\t\treturn tr.next, true\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn 0, false\n")
	buf.WriteString("}\n\n")

	if hasIgnoredActions {
		buf.WriteString("func " + typeName + "IsIgnoredToken(tokenType tokens.TokenType) bool {\n")
		buf.WriteString("\treturn strings.HasPrefix(string(tokenType), \"!\")\n")
		buf.WriteString("}\n\n")
	}

	buf.WriteString("const " + typeName + "StartState = ")
	buf.WriteString(fmt.Sprintf("%d\n\n", tables.StartState))

	buf.WriteString("type " + typeName + "RangeTransition struct {\n")
	buf.WriteString("\tfrom rune\n")
	buf.WriteString("\tto   rune\n")
	buf.WriteString("\tnext int\n")
	buf.WriteString("}\n\n")

	buf.WriteString("var " + typeName + "Transitions = map[int][]" + typeName + "RangeTransition{\n")
	writeTransitions(&buf, tables)
	buf.WriteString("}\n\n")

	buf.WriteString("var " + typeName + "Actions = map[int]tokens.TokenType{\n")
	writeActions(&buf, tables)
	buf.WriteString("}\n")

	return buf.Bytes(), nil
}

func writeTransitions(buf *bytes.Buffer, tables *Tables) {
	stateIDs := make([]int, 0, len(tables.Transitions))
	for state := range tables.Transitions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)
	for _, state := range stateIDs {
		buf.WriteString(fmt.Sprintf("\t%d: {\n", state))
		ranges := tables.Transitions[state]
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i].From == ranges[j].From {
				return ranges[i].To < ranges[j].To
			}
			return ranges[i].From < ranges[j].From
		})
		for _, tr := range ranges {
			buf.WriteString(fmt.Sprintf("\t\t{from: %q, to: %q, next: %d},\n", tr.From, tr.To, tr.Next))
		}
		buf.WriteString("\t},\n")
	}
}

func writeActions(buf *bytes.Buffer, tables *Tables) {
	stateIDs := make([]int, 0, len(tables.Actions))
	for state := range tables.Actions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)
	for _, state := range stateIDs {
		tokenType := tables.Actions[state]
		tokenType = strings.ReplaceAll(tokenType, "\"", "\\\"")
		buf.WriteString(fmt.Sprintf("\t%d: %q,\n", state, tokenType))
	}
}
