package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io"
	"os"

	"github.com/danieldk/citar/cmd/common"
	"github.com/danieldk/citar/model"
	"github.com/danieldk/conllx"
)

func main() {
	flag.Parse()

	if flag.NArg() != 2 {
		os.Exit(1)
	}

	config := common.MustParseConfig(flag.Arg(0))

	f, err := os.Open(flag.Arg(1))
	common.ExitIfError("Cannot open training data", err)
	defer f.Close()

	out, err := os.Create(config.Model)
	common.ExitIfError("Cannot open model for writing", err)
	defer out.Close()

	bufOut := bufio.NewWriter(out)
	defer bufOut.Flush()

	reader := conllx.NewReader(bufio.NewReader(f))

	fc := model.NewFrequencyCollector()

	for {
		sent, err := reader.ReadSentence()
		if err == io.EOF {
			break
		}

		common.ExitIfError("Cannot read sentence", err)

		err = fc.Process(sent)
		common.ExitIfError("Cannot process sentence", err)
	}

	model := fc.Model()
	enc := gob.NewEncoder(bufOut)
	err = enc.Encode(model)
	common.ExitIfError("Cannot encode model", err)
}
