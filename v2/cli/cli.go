package cli

import (
	"fmt"
	"io"
	"strings"
)

const usage = `PGPG - combined lexer and LALR(1) parser generator

Usage:
  pgpg help
  pgpg bnf-to-ir <input.bnf>
  pgpg ir-to-go <grammar.json> --out <dir>
  pgpg all-in-one <input.bnf> --lang go --out <dir>
`

func Run(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprint(errOut, usage)
		return 2
	}

	switch strings.ToLower(args[1]) {
	case "help", "-h", "--help":
		fmt.Fprint(out, usage)
		return 0
	case "bnf-to-ir":
		return runBNFToIR(args[2:], out, errOut)
	case "ir-to-go":
		return runIRToGo(args[2:], out, errOut)
	case "all-in-one":
		return runAllInOne(args[2:], out, errOut)
	default:
		fmt.Fprintf(errOut, "unknown command: %s\n\n%s", args[1], usage)
		return 2
	}
}
