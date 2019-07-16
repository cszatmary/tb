// TODO: Move to its own packages if we end up with more functions
package util

import (
	"fmt"
	"os"
)

func FatalErr(message string, err error) {
	fmt.Fprintf(os.Stderr, message+"\n")

	if err != nil {
		// TODO: Make this user-configurable in tbrc - they may not always want the stack trace
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
	}

	os.Exit(1)
}

func Fatal(message string) {
	fmt.Fprintf(os.Stderr, message+"\n")
	os.Exit(1)
}
