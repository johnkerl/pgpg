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

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", packageName)
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"unicode/utf8\"\n\n")
	buf.WriteString("\t\"github.com/johnkerl/pgpg/manual/pkg/tokens\"\n")
	buf.WriteString(")\n\n")

	fmt.Fprintf(&buf, "type %s struct {\n", typeName)
	buf.WriteString("\tinputText     string\n")
	buf.WriteString("\tinputLength   int\n")
	buf.WriteString("\ttokenLocation *tokens.TokenLocation\n")
	buf.WriteString("}\n\n")

	fmt.Fprintf(&buf, "func New%s(inputText string) *%s {\n", typeName, typeName)
	fmt.Fprintf(&buf, "\treturn &%s{\n", typeName)
	buf.WriteString("\t\tinputText:     inputText,\n")
	buf.WriteString("\t\tinputLength:   len(inputText),\n")
	buf.WriteString("\t\ttokenLocation: tokens.NewTokenLocation(),\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (lexer *" + typeName + ") Scan() *tokens.Token {\n")
	buf.WriteString("\tif lexer.tokenLocation.ByteOffset >= lexer.inputLength {\n")
	buf.WriteString("\t\treturn tokens.NewEOFToken(lexer.tokenLocation)\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tstartLocation := *lexer.tokenLocation\n")
	buf.WriteString("\tscanLocation := *lexer.tokenLocation\n")
	buf.WriteString("\tstate := startState\n")
	buf.WriteString("\tlastAcceptState := -1\n")
	buf.WriteString("\tlastAcceptLocation := scanLocation\n\n")
	buf.WriteString("\tfor {\n")
	buf.WriteString("\t\tif scanLocation.ByteOffset >= lexer.inputLength {\n")
	buf.WriteString("\t\t\tbreak\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tr, width := lexer.peekRuneAt(scanLocation.ByteOffset)\n")
	buf.WriteString("\t\tnextState, ok := transitions[state][r]\n")
	buf.WriteString("\t\tif !ok {\n")
	buf.WriteString("\t\t\tbreak\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t\tscanLocation.LocateRune(r, width)\n")
	buf.WriteString("\t\tstate = nextState\n")
	buf.WriteString("\t\tif _, ok := actions[state]; ok {\n")
	buf.WriteString("\t\t\tlastAcceptState = state\n")
	buf.WriteString("\t\t\tlastAcceptLocation = scanLocation\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tif lastAcceptState < 0 {\n")
	buf.WriteString("\t\tr, _ := lexer.peekRuneAt(lexer.tokenLocation.ByteOffset)\n")
	buf.WriteString("\t\treturn tokens.NewErrorToken(fmt.Sprintf(\"lexer: unrecognized input %q\", r), lexer.tokenLocation)\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tlexemeText := lexer.inputText[lexer.tokenLocation.ByteOffset:lastAcceptLocation.ByteOffset]\n")
	buf.WriteString("\tlexeme := []rune(lexemeText)\n")
	buf.WriteString("\t*lexer.tokenLocation = lastAcceptLocation\n")
	buf.WriteString("\treturn tokens.NewToken(lexeme, actions[lastAcceptState], &startLocation)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func (lexer *" + typeName + ") peekRuneAt(byteOffset int) (rune, int) {\n")
	buf.WriteString("\tr, width := utf8.DecodeRuneInString(lexer.inputText[byteOffset:])\n")
	buf.WriteString("\treturn r, width\n")
	buf.WriteString("}\n\n")

	buf.WriteString("const startState = ")
	buf.WriteString(fmt.Sprintf("%d\n\n", tables.StartState))

	buf.WriteString("var transitions = map[int]map[rune]int{\n")
	writeTransitions(&buf, tables)
	buf.WriteString("}\n\n")

	buf.WriteString("var actions = map[int]tokens.TokenType{\n")
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
		keys := make([]string, 0, len(tables.Transitions[state]))
		for key := range tables.Transitions[state] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			runes := []rune(key)
			if len(runes) != 1 {
				continue
			}
			buf.WriteString(fmt.Sprintf("\t\t%q: %d,\n", runes[0], tables.Transitions[state][key]))
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
