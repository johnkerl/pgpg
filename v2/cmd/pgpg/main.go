package main

import (
	"os"

	"github.com/johnkerl/pgpgx/cli"
)

func main() {
	os.Exit(cli.Run(os.Args, os.Stdout, os.Stderr))
}
