/*
go build github.com/johnkerl/pgpg/cmd/tmp
*/

package main

import (
	"fmt"

	"github.com/johnkerl/pgpg/internal/pkg/temp"
)

func main() {
	fmt.Printf("TEMP.FOO: %d\n", temp.Foo())
}
