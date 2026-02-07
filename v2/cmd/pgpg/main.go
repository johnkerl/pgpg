package main

import (
	"os"

	"github.com/johnkerl/pgpg/v2/cli"
)

func main() {
	os.Exit(cli.Run(os.Args, os.Stdout, os.Stderr))
}
