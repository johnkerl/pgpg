package parsegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"sort"
	"strconv"
	"strings"
	"text/template"

	_ "embed"
)

//go:embed templates/parser.go.tmpl
var parserTemplateText string

var parserTemplateFuncs = template.FuncMap{
	"childIndicesLiteral": func(indices []int) string {
		if len(indices) == 0 {
			return "[]int{}"
		}
		parts := make([]string, len(indices))
		for i, idx := range indices {
			parts[i] = strconv.Itoa(idx)
		}
		return "[]int{" + strings.Join(parts, ", ") + "}"
	},
	"quote": strconv.Quote,
}

var parserTemplate = template.Must(
	template.New("parser").Funcs(parserTemplateFuncs).Parse(parserTemplateText),
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
	raw, err := GenerateGoParserCodeRaw(tables, packageName, typeName)
	if err != nil {
		return nil, err
	}
	formatted, err := format.Source(raw)
	if err != nil {
		return nil, fmt.Errorf("format generated code: %w", err)
	}
	return formatted, nil
}

// GenerateGoParserCodeRaw creates unformatted Go source implementing an LR(1) parser from tables.
func GenerateGoParserCodeRaw(tables *Tables, packageName string, typeName string) ([]byte, error) {
	if tables == nil {
		return nil, fmt.Errorf("nil tables")
	}
	if packageName == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if typeName == "" {
		return nil, fmt.Errorf("type name is required")
	}

	data := parserTemplateData{
		PackageName: packageName,
		TypeName:    typeName,
		Actions:     buildParserActions(tables, typeName),
		Gotos:       buildParserGotos(tables),
		Productions: buildParserProductions(tables),
		HintMode:    tables.HintMode,
	}

	var buf bytes.Buffer
	if err := parserTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render parser template: %w", err)
	}
	return buf.Bytes(), nil
}

type parserTemplateData struct {
	PackageName string
	TypeName    string
	Actions     []parserActionState
	Gotos       []parserGotoState
	Productions []parserProductionInfo
	HintMode    string
}

type parserActionState struct {
	State   int
	Entries []parserActionEntry
}

type parserActionEntry struct {
	TerminalLiteral string
	KindLiteral     string
	Target          int
	HasTarget       bool
}

type parserGotoState struct {
	State   int
	Entries []parserGotoEntry
}

type parserGotoEntry struct {
	NontermLiteral string
	Target         int
}

type parserProductionInfo struct {
	LHSLiteral       string
	RHSCount         int
	HasHint          bool
	HasPassthrough   bool
	ParentIndex      int
	PassthroughIndex int
	ChildIndices     []int
	NodeType         string
}

func buildParserActions(tables *Tables, typeName string) []parserActionState {
	stateIDs := make([]int, 0, len(tables.Actions))
	for state := range tables.Actions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)

	out := make([]parserActionState, 0, len(stateIDs))
	for _, state := range stateIDs {
		entries := tables.Actions[state]
		terminals := make([]string, 0, len(entries))
		for term := range entries {
			terminals = append(terminals, term)
		}
		sort.Strings(terminals)

		actionEntries := make([]parserActionEntry, 0, len(terminals))
		for _, term := range terminals {
			action := entries[term]
			hasTarget := action.Type == "shift" || action.Type == "reduce"
			actionEntries = append(actionEntries, parserActionEntry{
				TerminalLiteral: tokenTypeLiteral(term),
				KindLiteral:     actionKindLiteral(action.Type, typeName),
				Target:          action.Target,
				HasTarget:       hasTarget,
			})
		}

		out = append(out, parserActionState{
			State:   state,
			Entries: actionEntries,
		})
	}
	return out
}

func buildParserGotos(tables *Tables) []parserGotoState {
	stateIDs := make([]int, 0, len(tables.Gotos))
	for state := range tables.Gotos {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)

	out := make([]parserGotoState, 0, len(stateIDs))
	for _, state := range stateIDs {
		entries := tables.Gotos[state]
		names := make([]string, 0, len(entries))
		for name := range entries {
			names = append(names, name)
		}
		sort.Strings(names)

		gotoEntries := make([]parserGotoEntry, 0, len(names))
		for _, name := range names {
			gotoEntries = append(gotoEntries, parserGotoEntry{
				NontermLiteral: "asts.NodeType(" + strconv.Quote(name) + ")",
				Target:         entries[name],
			})
		}

		out = append(out, parserGotoState{
			State:   state,
			Entries: gotoEntries,
		})
	}
	return out
}

func buildParserProductions(tables *Tables) []parserProductionInfo {
	out := make([]parserProductionInfo, 0, len(tables.Productions))
	for _, prod := range tables.Productions {
		info := parserProductionInfo{
			LHSLiteral: "asts.NodeType(" + strconv.Quote(prod.LHS) + ")",
			RHSCount:   len(prod.RHS),
		}
		if prod.Hint != nil {
			if prod.Hint.PassthroughIndex != nil {
				info.HasPassthrough = true
				info.PassthroughIndex = *prod.Hint.PassthroughIndex
			} else {
				info.HasHint = true
				info.ParentIndex = prod.Hint.ParentIndex
				info.ChildIndices = prod.Hint.ChildIndices
				info.NodeType = prod.Hint.NodeType
			}
		}
		out = append(out, info)
	}
	return out
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
