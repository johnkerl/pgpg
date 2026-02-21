package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
	"github.com/johnkerl/pgpg/manual/go/pkg/asts"

	"github.com/johnkerl/goffl/pkg/f2poly"
	"github.com/johnkerl/goffl/pkg/f2polymod"
	"github.com/johnkerl/goffl/pkg/intmod"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e | -l] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  -l: read stdin line-by-line, evaluate each line, print result (REPL mode).\n")
	fmt.Fprintf(os.Stderr, "  -mode: int (default), float, mod, intmod, f2poly, f2polymod.\n")
	fmt.Fprintf(os.Stderr, "  -mod: integer modulus for -mode=mod or -mode=intmod (required).\n")
	fmt.Fprintf(os.Stderr, "  -mod-poly: hex modulus polynomial for -mode=f2polymod (e.g. 0x11).\n")
	fmt.Fprintf(os.Stderr, "  With -l and stdin a TTY, -p sets the prompt (default \"> \"); use -p \"\" to disable.\n")
	fmt.Fprintf(os.Stderr, "  Without -e/-l: zero arguments = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var verbose bool
	var exprMode bool
	var lineMode bool
	var prompt string
	var mode string
	var modN int
	var modPolyHex string
	flag.BoolVar(&verbose, "v", false, "Print AST before evaluation")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.BoolVar(&lineMode, "l", false, "Read stdin line-by-line, evaluate each, print result (REPL)")
	flag.StringVar(&prompt, "p", "> ", "In -l mode with TTY stdin, prompt string (default \"> \"; use \"\" to disable)")
	flag.StringVar(&mode, "mode", "int", "Arithmetic mode: int, float, mod, intmod, f2poly, f2polymod")
	flag.IntVar(&modN, "mod", 0, "Integer modulus for -mode=mod or -mode=intmod")
	flag.StringVar(&modPolyHex, "mod-poly", "", "Hex modulus polynomial for -mode=f2polymod (e.g. 0x11)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	validModes := map[string]bool{"int": true, "float": true, "mod": true, "intmod": true, "f2poly": true, "f2polymod": true}
	if !validModes[mode] {
		fmt.Fprintf(os.Stderr, "pemdas-eval: -mode must be int, float, mod, intmod, f2poly, or f2polymod (got %q)\n", mode)
		os.Exit(1)
	}
	if (mode == "mod" || mode == "intmod") && modN <= 0 {
		fmt.Fprintln(os.Stderr, "pemdas-eval: -mode=mod and -mode=intmod require -mod N with N > 0")
		os.Exit(1)
	}
	if mode == "f2polymod" && modPolyHex == "" {
		fmt.Fprintln(os.Stderr, "pemdas-eval: -mode=f2polymod requires -mod-poly HEX (e.g. -mod-poly 0x11)")
		os.Exit(1)
	}

	if lineMode {
		if exprMode {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e and -l are mutually exclusive")
			os.Exit(1)
		}
		runREPL(verbose, prompt, mode, modN, modPolyHex)
		return
	}

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e requires at least one argument")
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, verbose, mode, modN, modPolyHex); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	} else {
		if len(args) == 0 {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := runParserOnce(string(content), verbose, mode, modN, modPolyHex); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := runParserOnFiles(args, verbose, mode, modN, modPolyHex); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func runREPL(verbose bool, prompt string, mode string, modN int, modPolyHex string) {
	usePrompt := stdinIsTTY() && prompt != ""
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if usePrompt {
			fmt.Fprint(os.Stdout, prompt)
			os.Stdout.Sync()
		}
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := runParserOnce(line, verbose, mode, modN, modPolyHex); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParserOnFiles(filenames []string, verbose bool, mode string, modN int, modPolyHex string) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(string(content), verbose, mode, modN, modPolyHex); err != nil {
			return err
		}
	}
	return nil
}

func runParserOnce(input string, verbose bool, mode string, modN int, modPolyHex string) error {
	ast, err := parseWithMode(input, mode)
	if err != nil {
		return err
	}
	switch mode {
	case "int":
		var b IntNumeric
		result, err := evaluateAST[int, int](ast, b, verbose)
		if err != nil {
			return err
		}
		fmt.Println(b.String(result))
	case "float":
		var b FloatNumeric
		result, err := evaluateAST[float64, float64](ast, b, verbose)
		if err != nil {
			return err
		}
		fmt.Println(b.String(result))
	case "mod":
		numeric, err := NewModNumeric(modN)
		if err != nil {
			return err
		}
		result, err := evaluateAST[ModInt, int](ast, numeric, verbose)
		if err != nil {
			return err
		}
		fmt.Println(numeric.String(result))
	case "intmod":
		backend, err := NewIntModNumeric(int64(modN))
		if err != nil {
			return err
		}
		result, err := evaluateAST[*intmod.IntMod, int](ast, backend, verbose)
		if err != nil {
			return err
		}
		fmt.Println(backend.String(result))
	case "f2poly":
		var b F2PolyNumeric
		result, err := evaluateAST[*f2poly.F2Poly, int](ast, b, verbose)
		if err != nil {
			return err
		}
		fmt.Println(b.String(result))
	case "f2polymod":
		modBits, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(modPolyHex, "0x"), "0X"), 16, 64)
		if err != nil {
			return fmt.Errorf("invalid -mod-poly hex %q: %w", modPolyHex, err)
		}
		modulus := f2poly.New(modBits)
		backend, err := NewF2PolyModNumeric(modulus)
		if err != nil {
			return err
		}
		result, err := evaluateAST[*f2polymod.F2PolyMod, int](ast, backend, verbose)
		if err != nil {
			return err
		}
		fmt.Println(backend.String(result))
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
	return nil
}

// parseWithMode returns an AST using the lexer/parser for the given mode.
func parseWithMode(input string, mode string) (*asts.AST, error) {
	switch mode {
	case "int":
		lexer := generatedlexers.NewPEMDASIntLexer(input)
		return generatedparsers.NewPEMDASIntParser().Parse(lexer, "")
	case "float":
		lexer := generatedlexers.NewPEMDASFloatLexer(input)
		return generatedparsers.NewPEMDASFloatParser().Parse(lexer, "")
	case "mod", "intmod":
		lexer := generatedlexers.NewPEMDASModLexer(input)
		return generatedparsers.NewPEMDASModParser().Parse(lexer, "")
	case "f2poly", "f2polymod":
		lexer := generatedlexers.NewPEMDASF2PolyLexer(input)
		return generatedparsers.NewPEMDASF2PolyParser().Parse(lexer, "")
	default:
		return nil, fmt.Errorf("unsupported mode %q", mode)
	}
}
