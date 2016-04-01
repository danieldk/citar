package common

import (
	"flag"
	"os"
)

// FileOrStdin opens the file at the given index for reading when the
// index is valid. Otherwise, it returns os.Stdin.
func FileOrStdin(args []string, idx int) *os.File {
	if len(args) > idx {
		input, err := os.Open(flag.Arg(idx))
		exitIfError(err)
		return input
	}

	return os.Stdin
}

// FileOrStdout opens the file at the given index for writing when the
// index is valid. Otherwise, it returns os.Stdout.
func FileOrStdout(args []string, idx int) *os.File {
	if len(args) > idx {
		output, err := os.Create(flag.Arg(idx))
		exitIfError(err)
		return output
	}

	return os.Stdout
}

func exitIfError(err error) {
	if err != nil {
		panic(err)
	}
}
