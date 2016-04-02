// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"

	"github.com/danieldk/citar/cmd/common"
	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/tagger"
	"github.com/danieldk/citar/trigrams"
	"github.com/danieldk/citar/words"
	"github.com/danieldk/conllx"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] config input.conllx\n\n", os.Args[0])
		flag.PrintDefaults()
	}
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var nFolds = flag.Int("nfolds", 10, "number of cross-validation folds")

func trainFolds(testFold int) conllx.FoldSet {
	folds := make(conllx.FoldSet)

	for fold := 0; fold < *nFolds; fold++ {
		if fold != testFold {
			folds[fold] = nil
		}
	}

	return folds
}

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	if *nFolds < 2 {
		fmt.Fprintln(os.Stderr, "Data should be splitted in at least 2 folds.")
		os.Exit(1)
	}

	config := common.MustParseConfig(flag.Arg(0))

	var knownCorrect uint
	var knownIncorrect uint
	var unknownCorrect uint
	var unknownIncorrect uint

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	for fold := 0; fold < *nFolds; fold++ {
		fc := model.NewFrequencyCollector()

		err := processFolds(flag.Arg(1), trainFolds(fold), func(sent []conllx.Token) error {
			return fc.Process(sent)
		})
		common.ExitIfError("Error processing training folds", err)

		model := fc.Model()

		sh, err := config.UnknownWordHandler(model)
		common.ExitIfError("Could construct unknown word handler", err)

		lh := words.NewLexiconWithFallback(model.WordTagFreqs(), model.UnigramFreqs(), sh)
		lim := trigrams.NewLinearInterpolationModel(model)
		tagger := tagger.NewHMMTagger(model, lh, lim, 1000.0)

		eval := common.NewEvaluator(tagger, model)

		err = processFolds(flag.Arg(1), conllx.FoldSet{fold: nil}, func(sent []conllx.Token) error {
			return eval.Process(sent)
		})
		common.ExitIfError("Error processing testing fold", err)

		fmt.Printf("Fold %d accuracy: %2f (known: %2f, unknown: %2f)\n", fold, eval.Accuracy(),
			eval.KnownAccuracy(), eval.UnknownAccuracy())

		knownCorrect += eval.KnownCorrect()
		knownIncorrect += eval.KnownIncorrect()
		unknownCorrect += eval.UnknownCorrect()
		unknownIncorrect += eval.UnknownIncorrect()
	}

	accuracy := float64(knownCorrect+unknownCorrect) /
		float64(knownCorrect+unknownCorrect+knownIncorrect+unknownIncorrect)
	knownAccuracy := float64(knownCorrect) / float64(knownCorrect+knownIncorrect)
	unknownAccuracy := float64(unknownCorrect) / float64(unknownCorrect+unknownIncorrect)

	fmt.Printf("Overall accuracy: %2f (known: %2f, unknown: %2f)\n", accuracy,
		knownAccuracy, unknownAccuracy)

}

func processFolds(filename string, folds conllx.FoldSet, fun func(sent []conllx.Token) error) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := conllx.NewReader(bufio.NewReader(f))

	trainReader, err := conllx.NewSplittingReader(reader, *nFolds, folds)
	if err != nil {
		return err
	}

	for {
		sent, err := trainReader.ReadSentence()

		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		if err := fun(sent); err != nil {
			return err
		}
	}

	return nil
}
