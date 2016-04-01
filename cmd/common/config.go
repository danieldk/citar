// Copyright 2016 Daniël de Kok. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/words"
)

// CitarConfig stores the configuration of citar.
type CitarConfig struct {
	Model          string
	UnknownHandler string `toml:"unknown_handler"`
}

func (c CitarConfig) UnknownWordHandler(m model.Model) (words.WordHandler, error) {
	if cons, ok := unknownHandlers[c.UnknownHandler]; ok {
		return cons(m), nil
	}

	return nil, fmt.Errorf("Unknown word handler: %s", c.UnknownHandler)
}

func defaultConfiguration() *CitarConfig {
	return &CitarConfig{
		Model:          "model.gob",
		UnknownHandler: "lookup",
	}
}

func MustParseConfig(filename string) *CitarConfig {
	f, err := os.Open(filename)
	ExitIfError("Cannot open configuration file", err)
	defer f.Close()

	config, err := ParseConfig(f)
	ExitIfError("Cannot parse configuration file", err)

	config.Model = relToConfig(filename, config.Model)

	return config
}

// ParseConfig attempts to parse the configuration from the given reader.
func ParseConfig(reader io.Reader) (*CitarConfig, error) {
	config := defaultConfiguration()
	if _, err := toml.DecodeReader(reader, config); err != nil {
		return config, err
	}

	return config, nil
}

type unknownHandler func(m model.Model) words.WordHandler

// UnknownHandlers is a mapping from unknown words handlers to
// constructors of these handlers.
var unknownHandlers = map[string]unknownHandler{
	"tree": func(m model.Model) words.WordHandler {
		return words.NewSuffixHandler(words.DefaultSuffixHandlerConfig(), m)
	},
	"lookup": func(m model.Model) words.WordHandler {
		return words.NewLookupSuffixHandler(
			words.NewSuffixHandler(words.DefaultSuffixHandlerConfig(), m))
	},
}

// Return the path of a file, relative to the directory of
// the configuration file, unless the path is absolute.
func relToConfig(configPath, filePath string) string {
	if len(filePath) == 0 {
		return filePath
	}

	if filepath.IsAbs(filePath) {
		return filePath
	}

	return filepath.Join(filepath.Dir(configPath), filePath)
}
