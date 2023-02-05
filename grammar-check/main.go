package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"one/src/lexer"
	"one/src/parser"
)

func parseOne(input string) {
	green := "\033[32;01m"
	red := "\033[31;01m"
	blue := "\033[33;01m"
	textdefault := "\033[0m"

	if strings.HasPrefix(input, "#") {
		fmt.Printf("%s%s%s\n", blue, input, textdefault)
		return
	}

	theLexer := lexer.NewLexer([]byte(input))
	theParser := parser.NewParser()
	_, err := theParser.Parse(theLexer)

	if err != nil {
		fmt.Printf("%sFail%s %s\n", red, textdefault, input)
	} else {
		fmt.Printf("%sOK%s   %s\n", green, textdefault, input)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s expr 'expression ...'\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s file {one or more file names}\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	if os.Args[1] == "expr" {
		text := strings.Join(os.Args[2:], " ")
		parseOne(text)
	} else if os.Args[1] == "file" {
		for _, filename := range(os.Args[2:]) {
			handle, err := os.Open(filename)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			lineScanner := bufio.NewScanner(handle)
			for lineScanner.Scan() {
				parseOne(lineScanner.Text())
			}
		}
	} else {
		usage()
	}
}
