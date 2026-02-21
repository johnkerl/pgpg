package run

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnkerl/pgpg/generators/go/pkg/lexgen"
	"github.com/johnkerl/pgpg/generators/go/pkg/parsegen"
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
	if err := ctx.Err(); err != nil {
		return err
	}
	grammar, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read grammar: %w", err)
	}
	sourceName := ""
	if opts != nil {
		sourceName = opts.SourceName
	}
	if sourceName == "" {
		sourceName, _ = filepath.Abs(inputPath)
	}
	tables, err := lexgen.GenerateTables(string(grammar), &lexgen.LexTableOptions{SourceName: sourceName})
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	encodeOpts := (*lexgen.EncodeOptions)(nil)
	if opts != nil {
		encodeOpts = opts.Encode
	}
	jsonBytes, err := lexgen.EncodeTables(tables, encodeOpts)
	if err != nil {
		return fmt.Errorf("encode tables: %w", err)
	}
	return writeOutput(ctx, outputPath, append(jsonBytes, '\n'))
}

// LexgenCode reads tables JSON from tablesPath, generates Go lexer code, and writes to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func LexgenCode(ctx context.Context, tablesPath, outputPath string, opts lexgen.LexCodegenOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := os.ReadFile(tablesPath)
	if err != nil {
		return fmt.Errorf("read tables: %w", err)
	}
	tables, err := lexgen.DecodeTables(data)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	code, err := lexgen.GenerateCode(tables, opts)
	if err != nil {
		return err
	}
	return writeOutput(ctx, outputPath, code)
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
	if err := ctx.Err(); err != nil {
		return err
	}
	grammar, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read grammar: %w", err)
	}
	sourceName := ""
	if opts != nil {
		sourceName = opts.SourceName
	}
	if sourceName == "" {
		sourceName, _ = filepath.Abs(inputPath)
	}
	tables, err := parsegen.GenerateTables(string(grammar), &parsegen.ParseTableOptions{SourceName: sourceName})
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	encodeOpts := (*parsegen.EncodeOptions)(nil)
	if opts != nil {
		encodeOpts = opts.Encode
	}
	jsonBytes, err := parsegen.EncodeTables(tables, encodeOpts)
	if err != nil {
		return fmt.Errorf("encode tables: %w", err)
	}
	return writeOutput(ctx, outputPath, append(jsonBytes, '\n'))
}

// ParsegenCode reads tables JSON from tablesPath, generates Go parser code, and writes to outputPath.
// If outputPath is "" or "-", writes to stdout. ctx is used for cancellation.
func ParsegenCode(ctx context.Context, tablesPath, outputPath string, opts parsegen.ParseCodegenOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := os.ReadFile(tablesPath)
	if err != nil {
		return fmt.Errorf("read tables: %w", err)
	}
	tables, err := parsegen.DecodeTables(data)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	code, err := parsegen.GenerateCode(tables, opts)
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
