package common

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/words"
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
		}
	}

	return tags
}

func MustLoadSubstitutions(filename string) []words.Substitution {
	substs := make([]words.Substitution, 0)

	if filename == "" {
		return substs
	}

	f, err := os.Open(filename)
	ExitIfError("cannot open substitution file", err)
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line != "" {
			parts := strings.Split(line, "\t")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Incorrect substitution: %s\n", line)
				os.Exit(1)
			}

			pattern := regexp.MustCompile(parts[0])
			substs = append(substs, words.Substitution{pattern, parts[1]})
		}
	}

	return substs
}
