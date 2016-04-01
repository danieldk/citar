// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"strings"
)

var _ gob.GobEncoder = &StringNumberer{}
var _ gob.GobDecoder = &StringNumberer{}

// A StringNumberer creates a bijection between (string-based) labels
// and numbers.
type StringNumberer struct {
	labelNumbers map[string]uint
	labels       []string
}

// NewStringStringNumberer creates a new StringNumberer that is empty (it
// has no mappings yet).
func NewStringStringNumberer() *StringNumberer {
	return &StringNumberer{make(map[string]uint), make([]string, 0)}
}

// Number returns the (unique) number for for a label (string).
func (l *StringNumberer) Number(label string) uint {
	idx, ok := l.labelNumbers[label]

	if !ok {
		idx = uint(len(l.labelNumbers))
		l.labelNumbers[label] = idx
		l.labels = append(l.labels, label)
	}

	return idx
}

// Label returns the label (string) for a number.
func (l *StringNumberer) Label(number uint) string {
	return l.labels[number]
}

// Size returns the number of labels known in the bijection.
func (l *StringNumberer) Size() int {
	return len(l.labels)
}

// Read a label <-> number bijection from a Reader.
func (l *StringNumberer) Read(reader io.Reader) error {
	var labels []string
	bufReader := bufio.NewReader(reader)

	eof := false
	for !eof {
		line, err := bufReader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}

			eof = true
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		labels = append(labels, line)
	}

	numbers := make(map[string]uint)
	for idx, label := range labels {
		numbers[label] = uint(idx)
	}

	l.labels = labels
	l.labelNumbers = numbers

	return nil
}

// WriteStringStringNumberer writes the bijection in a StringNumberer to a file.
func (l *StringNumberer) WriteStringStringNumberer(writer io.Writer) error {
	for _, s := range l.labels {
		fmt.Fprintf(writer, "%s\n", s)
	}

	return nil
}

// GobDecode decodes a Model from a gob.
func (l *StringNumberer) GobDecode(data []byte) error {
	var en encodedStringNumberer
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	if err := decoder.Decode(&en); err != nil {
		return err
	}

	l.labelNumbers = en.LabelNumbers
	l.labels = en.Labels

	return nil
}

// GobEncode encodes a StringNumberer as a gob.
func (l *StringNumberer) GobEncode() ([]byte, error) {
	en := encodedStringNumberer{
		LabelNumbers: l.labelNumbers,
		Labels:       l.labels,
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(en); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type encodedStringNumberer struct {
	LabelNumbers map[string]uint
	Labels       []string
}
