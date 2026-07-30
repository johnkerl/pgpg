package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnkerl/pgpg/go/generators/pkg/lexgen"
	"github.com/johnkerl/pgpg/go/generators/pkg/parsegen"
)

// LexgenTablesOptions configures LexgenTables (BNF → JSON).
type LexgenTablesOptions struct {
	// SourceName is used in error messages (e.g. file path). If empty, the input path is used.
	SourceName string
	// Encode controls JSON encoding. Nil means deterministic key order.
	Encode *lexgen.EncodeOptions
}

// LexgenTables reads a BNF grammar from inputPath, generates lexer tables, and writes JSON to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func LexgenTables(ctx context.Context, inputPath, outputPath string, opts *LexgenTablesOptions) error {
	var sourceName string
	var encodeOpts *lexgen.EncodeOptions
	if opts != nil {
		sourceName = opts.SourceName
		encodeOpts = opts.Encode
	}
	return runTables(ctx, inputPath, outputPath, sourceName, func(grammarText, sourceName string) ([]byte, error) {
		tables, err := lexgen.GenerateTables(grammarText, &lexgen.LexTableOptions{SourceName: sourceName})
		if err != nil {
			return nil, err
		}
		jsonBytes, err := lexgen.EncodeTables(tables, encodeOpts)
		if err != nil {
			return nil, fmt.Errorf("encode tables: %w", err)
		}
		return jsonBytes, nil
	})
}

// LexgenCode reads tables JSON from tablesPath, generates Go lexer code, and writes to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func LexgenCode(ctx context.Context, tablesPath, outputPath string, opts lexgen.LexCodegenOptions) error {
	return runCode(ctx, tablesPath, outputPath, func(data []byte) ([]byte, error) {
		tables, err := lexgen.DecodeTables(data)
		if err != nil {
			return nil, err
		}
		return lexgen.GenerateCode(tables, opts)
	})
}

// ParsegenTablesOptions configures ParsegenTables (BNF → JSON).
type ParsegenTablesOptions struct {
	// SourceName is used in error messages (e.g. file path). If empty, the input path is used.
	SourceName string
	// Encode controls JSON encoding. Nil means deterministic key order.
	Encode *parsegen.EncodeOptions
}

// ParsegenTables reads a BNF grammar from inputPath, generates parser tables, and writes JSON to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func ParsegenTables(ctx context.Context, inputPath, outputPath string, opts *ParsegenTablesOptions) error {
	var sourceName string
	var encodeOpts *parsegen.EncodeOptions
	if opts != nil {
		sourceName = opts.SourceName
		encodeOpts = opts.Encode
	}
	return runTables(ctx, inputPath, outputPath, sourceName, func(grammarText, sourceName string) ([]byte, error) {
		tables, err := parsegen.GenerateTables(grammarText, &parsegen.ParseTableOptions{SourceName: sourceName})
		if err != nil {
			return nil, err
		}
		jsonBytes, err := parsegen.EncodeTables(tables, encodeOpts)
		if err != nil {
			return nil, fmt.Errorf("encode tables: %w", err)
		}
		return jsonBytes, nil
	})
}

// ParsegenCode reads tables JSON from tablesPath, generates Go parser code, and writes to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func ParsegenCode(ctx context.Context, tablesPath, outputPath string, opts parsegen.ParseCodegenOptions) error {
	return runCode(ctx, tablesPath, outputPath, func(data []byte) ([]byte, error) {
		tables, err := parsegen.DecodeTables(data)
		if err != nil {
			return nil, err
		}
		return parsegen.GenerateCode(tables, opts)
	})
}

// runTables reads a grammar file, runs generate (grammar text + resolved
// source name → encoded JSON tables), and writes the result plus a trailing
// newline to outputPath.
func runTables(ctx context.Context, inputPath, outputPath, sourceName string, generate func(grammarText, sourceName string) ([]byte, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	grammar, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read grammar: %w", err)
	}
	if sourceName == "" {
		if abs, err := filepath.Abs(inputPath); err == nil {
			sourceName = abs
		} else {
			sourceName = inputPath
		}
	}
	jsonBytes, err := generate(string(grammar), sourceName)
	if err != nil {
		return err
	}
	return writeOutput(ctx, outputPath, append(jsonBytes, '\n'))
}

// runCode reads a tables JSON file, runs generate (JSON bytes → Go source),
// and writes the result to outputPath.
func runCode(ctx context.Context, tablesPath, outputPath string, generate func(data []byte) ([]byte, error)) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := os.ReadFile(tablesPath)
	if err != nil {
		return fmt.Errorf("read tables: %w", err)
	}
	code, err := generate(data)
	if err != nil {
		return err
	}
	return writeOutput(ctx, outputPath, code)
}

func writeOutput(ctx context.Context, outputPath string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if outputPath == "" || outputPath == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(outputPath, data, 0o644)
}
