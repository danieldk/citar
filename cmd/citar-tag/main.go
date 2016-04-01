package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io"
	"os"

	"github.com/danieldk/citar/cmd/common"
	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/tagger"
	"github.com/danieldk/citar/trigrams"
	"github.com/danieldk/citar/words"
	"github.com/danieldk/conllx"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 || flag.NArg() > 3 {
		os.Exit(1)
	}

	config := common.MustParseConfig(flag.Arg(0))

	modelFile, err := os.Open(config.Model)
	common.ExitIfError("Cannot open model", err)
	defer modelFile.Close()

	inputFile := common.FileOrStdin(flag.Args(), 1)
	defer inputFile.Close()

	outputFile := common.FileOrStdout(flag.Args(), 2)
	defer outputFile.Close()

	var model model.Model
	decoder := gob.NewDecoder(modelFile)
	err = decoder.Decode(&model)
	common.ExitIfError("Could not load model", err)

	sh, err := config.UnknownWordHandler(model)
	common.ExitIfError("Could construct unknown word handler", err)

	lh := words.NewLexiconWithFallback(model.WordTagFreqs(), model.UnigramFreqs(), sh)
	lim := trigrams.NewLinearInterpolationModel(model)
	tagger := tagger.NewHMMTagger(model, lh, lim, 1000.0)

	reader := conllx.NewReader(bufio.NewReader(inputFile))
	bufWriter := bufio.NewWriter(outputFile)
	defer bufWriter.Flush()
	writer := conllx.NewWriter(bufWriter)

	for {
		sent, err := reader.ReadSentence()
		if err == io.EOF {
			break
		}
		common.ExitIfError("Cannot read sentence", err)

		words := tokenToWords(sent)
		tags, _ := tagger.Tag(words).Tags()
		addTags(sent, tags)

		err = writer.WriteSentence(sent)
		common.ExitIfError("Cannot write sentence", err)
	}
}

func tokenToWords(sent []conllx.Token) []string {
	words := make([]string, 0, len(sent))

	for _, token := range sent {
		if form, ok := token.Form(); ok {
			words = append(words, form)
		} else {
			// Unlikely that there was no token, probably an underscore interpreted
			// as absent.
			words = append(words, "_")
		}
	}

	return words
}

func addTags(sent []conllx.Token, tags []string) {
	for i := range sent {
		sent[i].SetPosTag(tags[i])
	}
}