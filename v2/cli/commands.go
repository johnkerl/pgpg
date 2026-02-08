package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/johnkerl/pgpg/tmp/codegen"
	"github.com/johnkerl/pgpg/tmp/grammar"
	"github.com/johnkerl/pgpg/tmp/ir"
)

func runBNFToIR(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("bnf-to-ir", flag.ContinueOnError)
	fs.SetOutput(errOut)

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(errOut, "bnf-to-ir expects exactly one input file")
		return 2
	}

	inputPath := fs.Arg(0)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(errOut, "bnf-to-ir: read %s: %v\n", inputPath, err)
		return 1
	}

	grammarAST, err := grammar.ParseBNF(data)
	if err != nil {
		fmt.Fprintf(errOut, "bnf-to-ir: parse %s: %v\n", inputPath, err)
		return 1
	}

	irDoc := ir.FromGrammar(grammarAST, inputPath)
	if err := ir.EncodeJSON(out, irDoc); err != nil {
		fmt.Fprintf(errOut, "bnf-to-ir: write JSON: %v\n", err)
		return 1
	}
	return 0
}

func runIRToGo(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("ir-to-go", flag.ContinueOnError)
	fs.SetOutput(errOut)
	outDir := fs.String("out", "", "output directory")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(errOut, "ir-to-go expects exactly one input file")
		return 2
	}
	if *outDir == "" {
		fmt.Fprintln(errOut, "ir-to-go requires --out <dir>")
		return 2
	}

	inputPath := fs.Arg(0)
	input, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(errOut, "ir-to-go: open %s: %v\n", inputPath, err)
		return 1
	}
	defer input.Close()

	doc, err := ir.DecodeJSON(input)
	if err != nil {
		fmt.Fprintf(errOut, "ir-to-go: decode %s: %v\n", inputPath, err)
		return 1
	}

	gen, err := codegen.NewGo(out)
	if err != nil {
		fmt.Fprintf(errOut, "ir-to-go: init generator: %v\n", err)
		return 1
	}
	if err := gen.Generate(doc, *outDir); err != nil {
		fmt.Fprintf(errOut, "ir-to-go: generate: %v\n", err)
		return 1
	}
	return 0
}

func runAllInOne(args []string, out io.Writer, errOut io.Writer) int {
	fs := flag.NewFlagSet("all-in-one", flag.ContinueOnError)
	fs.SetOutput(errOut)
	lang := fs.String("lang", "go", "output language")
	outDir := fs.String("out", "", "output directory")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(errOut, "all-in-one expects exactly one input file")
		return 2
	}
	if *outDir == "" {
		fmt.Fprintln(errOut, "all-in-one requires --out <dir>")
		return 2
	}
	if *lang != "go" {
		fmt.Fprintf(errOut, "unsupported language: %s\n", *lang)
		return 2
	}

	inputPath := fs.Arg(0)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(errOut, "all-in-one: read %s: %v\n", inputPath, err)
		return 1
	}

	grammarAST, err := grammar.ParseBNF(data)
	if err != nil {
		fmt.Fprintf(errOut, "all-in-one: parse %s: %v\n", inputPath, err)
		return 1
	}

	irDoc := ir.FromGrammar(grammarAST, inputPath)
	gen, err := codegen.NewGo(out)
	if err != nil {
		fmt.Fprintf(errOut, "all-in-one: init generator: %v\n", err)
		return 1
	}
	if err := gen.Generate(irDoc, *outDir); err != nil {
		fmt.Fprintf(errOut, "all-in-one: generate: %v\n", err)
		return 1
	}
	return 0
}

func notImplemented(errOut io.Writer, cmd string) int {
	_ = errors.New("placeholder to keep this function non-trivial")
	fmt.Fprintf(errOut, "%s: not implemented yet\n", cmd)
	return 1
}
