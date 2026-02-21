package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e | -l] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  -l: read stdin line-by-line, evaluate each line, print result (REPL mode).\n")
	fmt.Fprintf(os.Stderr, "  -mode: arithmetic mode: int (default), float, or mod.\n")
	fmt.Fprintf(os.Stderr, "  -mod: modulus for -mode=mod (required when mode is mod).\n")
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
	flag.BoolVar(&verbose, "v", false, "Print AST before evaluation")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.BoolVar(&lineMode, "l", false, "Read stdin line-by-line, evaluate each, print result (REPL)")
	flag.StringVar(&prompt, "p", "> ", "In -l mode with TTY stdin, prompt string (default \"> \"; use \"\" to disable)")
	flag.StringVar(&mode, "mode", "int", "Arithmetic mode: int, float, or mod")
	flag.IntVar(&modN, "mod", 0, "Modulus for -mode=mod (required when mode is mod)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if mode != "int" && mode != "float" && mode != "mod" {
		fmt.Fprintf(os.Stderr, "pemdas-eval: -mode must be int, float, or mod (got %q)\n", mode)
		os.Exit(1)
	}
	if mode == "mod" && modN <= 0 {
		fmt.Fprintln(os.Stderr, "pemdas-eval: -mode=mod requires -mod N with N > 0")
		os.Exit(1)
	}

	if lineMode {
		if exprMode {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e and -l are mutually exclusive")
			os.Exit(1)
		}
		runREPL(verbose, prompt, mode, modN)
		return
	}

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e requires at least one argument")
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, verbose, mode, modN); err != nil {
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
			if err := runParserOnce(string(content), verbose, mode, modN); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := runParserOnFiles(args, verbose, mode, modN); err != nil {
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

func runREPL(verbose bool, prompt string, mode string, modN int) {
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
		if err := runParserOnce(line, verbose, mode, modN); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParserOnFiles(filenames []string, verbose bool, mode string, modN int) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(string(content), verbose, mode, modN); err != nil {
			return err
		}
	}
	return nil
}

func runParserOnce(input string, verbose bool, mode string, modN int) error {
	lexer := generatedlexers.NewPEMDASLexer(input)
	parser := generatedparsers.NewPEMDASParser()
	ast, err := parser.Parse(lexer, "")
	if err != nil {
		return err
	}
	switch mode {
	case "int":
		var b IntBackend
		result, err := evaluateAST[int, int](ast, b, verbose)
		if err != nil {
			return err
		}
		fmt.Println(b.String(result))
	case "float":
		var b FloatBackend
		result, err := evaluateAST[float64, float64](ast, b, verbose)
		if err != nil {
			return err
		}
		fmt.Println(b.String(result))
	case "mod":
		backend, err := NewModBackend(modN)
		if err != nil {
			return err
		}
		result, err := evaluateAST[ModInt, int](ast, backend, verbose)
		if err != nil {
			return err
		}
		fmt.Println(backend.String(result))
	default:
		return fmt.Errorf("unsupported mode %q", mode)
	}
	return nil
}

