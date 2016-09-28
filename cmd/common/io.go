package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danieldk/citar/model"
)

func MustLoadClosedClass(filename string) model.ClosedClassSet {
	tags := make(model.ClosedClassSet)

	if filename == "" {
		return tags
	}

	f, err := os.Open(filename)
	ExitIfError("cannot open closed class tag file", err)
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		tag := strings.TrimSpace(line)

		if tag != "" {
			tags[tag] = nil
			fmt.Fprintf(os.Stderr, "Added closed-class tag: %s\n", tag)
		}
	}

	return tags
}
