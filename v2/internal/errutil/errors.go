package errutil

import "fmt"

// NotImplemented returns a consistent error for unfinished code paths.
func NotImplemented(feature string) error {
	return fmt.Errorf("%s: not implemented", feature)
}
