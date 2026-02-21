package lexgen

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

type lexerTemplateData struct {
	PackageName string
	TypeName    string
	StartState  int
	HasIgnored  bool
	Transitions []lexerTransitionState
	Actions     []lexerActionState
}

type lexerTransitionState struct {
	State  int
	Ranges []RangeTransition
}

type lexerActionState struct {
	State     int
	TokenType string
}

//go:embed templates/lexer.go.tmpl
var lexerTemplateText string

var lexerTemplate = template.Must(
	template.New("lexer").Funcs(template.FuncMap{
		"quote": strconv.Quote,
	}).Parse(lexerTemplateText),
)

// LexCodegenOptions configures Go lexer code generation from tables.
type LexCodegenOptions struct {
	Package string // Go package name for generated code
	Type    string // Go type name for the lexer
	Format  bool   // run go/format.Source on output
}

// DecodeTables reads tables JSON into Tables.
func DecodeTables(data []byte) (*Tables, error) {
	var tables Tables
	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	}
	return &tables, nil
}

// GenerateCode creates Go source implementing a lexer from tables.
// When opts.Format is true, output is run through go/format.Source.
func GenerateCode(tables *Tables, opts LexCodegenOptions) ([]byte, error) {
	if tables == nil {
		return nil, fmt.Errorf("nil tables")
	}
	if opts.Package == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if opts.Type == "" {
		return nil, fmt.Errorf("type name is required")
	}
	raw, err := generateCodeRaw(tables, opts.Package, opts.Type)
	if err != nil {
		return nil, err
	}
	if !opts.Format {
		return raw, nil
	}
	formatted, err := format.Source(raw)
	if err != nil {
		return nil, fmt.Errorf("format generated code: %w", err)
	}
	return formatted, nil
}

func generateCodeRaw(tables *Tables, packageName string, typeName string) ([]byte, error) {
	hasIgnoredActions := false
	for _, action := range tables.Actions {
		if strings.HasPrefix(action, "!") {
			hasIgnoredActions = true
			break
		}
	}

	data := lexerTemplateData{
		PackageName: packageName,
		TypeName:    typeName,
		StartState:  tables.StartState,
		HasIgnored:  hasIgnoredActions,
		Transitions: buildLexerTransitions(tables),
		Actions:     buildLexerActions(tables),
	}

	var buf bytes.Buffer
	if err := lexerTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render lexer template: %w", err)
	}
	return buf.Bytes(), nil
}

func buildLexerTransitions(tables *Tables) []lexerTransitionState {
	stateIDs := make([]int, 0, len(tables.Transitions))
	for state := range tables.Transitions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)

	out := make([]lexerTransitionState, 0, len(stateIDs))
	for _, state := range stateIDs {
		ranges := tables.Transitions[state]
		sort.Slice(ranges, func(i, j int) bool {
			if ranges[i].From == ranges[j].From {
				return ranges[i].To < ranges[j].To
			}
			return ranges[i].From < ranges[j].From
		})
		out = append(out, lexerTransitionState{
			State:  state,
			Ranges: ranges,
		})
	}
	return out
}

func buildLexerActions(tables *Tables) []lexerActionState {
	stateIDs := make([]int, 0, len(tables.Actions))
	for state := range tables.Actions {
		stateIDs = append(stateIDs, state)
	}
	sort.Ints(stateIDs)

	out := make([]lexerActionState, 0, len(stateIDs))
	for _, state := range stateIDs {
		out = append(out, lexerActionState{
			State:     state,
			TokenType: tables.Actions[state],
		})
	}
	return out
}
