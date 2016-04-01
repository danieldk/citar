// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tagger

import (
	"fmt"
	"math"

	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/trigrams"
	"github.com/danieldk/citar/words"
)

// A Trellis is used during HMM tagging to store possible analyses.
type Trellis struct {
	lastColumn []*trellisState
	model      model.Model
}

// Tags returns the most likely part-of-speech tag sequence in the
// Trellis.
func (t Trellis) Tags() ([]string, float64) {
	tagNumbers, prob := t.highestProbabilitySequence()
	tagNumberer := t.model.TagNumberer()

	tags := make([]string, 0, len(tagNumbers))

	for i := 2; i < len(tagNumbers)-1; i++ {
		tag := tagNumberer.Label(tagNumbers[i])
		tags = append(tags, tag)
	}

	return tags, prob
}

func (t Trellis) highestProbabilitySequence() ([]uint, float64) {
	highestProb := math.Inf(-1)
	var tail *trellisState
	var beforeTail *trellisState

	// Find the most probable state in the last column.
	for _, state := range t.lastColumn {
		for previousState, bp := range state.backpointers {
			if bp.prob > highestProb {
				highestProb = bp.prob
				tail = state
				beforeTail = previousState
			}
		}
	}

	if tail == nil {
		panic("nil tail while extracting highest probability sequence")
	}

	var tagSequence []uint
	for {
		tagSequence = append(tagSequence, tail.tag.Tag)

		if beforeTail == nil {
			break
		}

		tail, beforeTail = beforeTail, tail.backpointers[beforeTail].state
	}

	reverse(tagSequence)

	return tagSequence, highestProb
}

type trellisState struct {
	tag          model.Tag
	backpointers map[*trellisState]backpointer
}

type backpointer struct {
	state *trellisState
	prob  float64
}

func newTrellisState(tag model.Tag) *trellisState {
	return &trellisState{
		tag:          tag,
		backpointers: make(map[*trellisState]backpointer),
	}
}

// HMMTagger implement a Hidden Markov Model (HMM) part-of-speech tagger.
type HMMTagger struct {
	model        model.Model
	wordHandler  words.WordHandler
	trigramModel trigrams.TrigramModel
	beamFactor   float64
}

// NewHMMTagger constructs a new tagger from the given data model, word
// handler and trigram model. The beam factor specifies how aggressively
// the search space should be pruned. For instance, a beam factor of 1000
// will exclude all paths that are 1000 times less probable than the most
// probable path.
func NewHMMTagger(model model.Model, wordHandler words.WordHandler,
	trigramModel trigrams.TrigramModel, beamFactor float64) HMMTagger {
	return HMMTagger{
		model:        model,
		wordHandler:  wordHandler,
		trigramModel: trigramModel,
		beamFactor:   math.Log(beamFactor),
	}
}

// Tag tags a sentence.
func (t HMMTagger) Tag(sentence []string) Trellis {
	tokens := make([]string, len(sentence)+3)
	tokens[0] = model.StartToken
	tokens[1] = model.StartToken
	copy(tokens[2:], sentence)
	tokens[len(tokens)-1] = model.EndToken

	return Trellis{
		lastColumn: t.viterbi(tokens),
		model:      t.model,
	}
}

func (t HMMTagger) viterbi(sentence []string) []*trellisState {
	var trellis []*trellisState
	var nextTrellis []*trellisState

	// Prepare initial trellis states.
	startTag := t.model.TagNumberer().Number(sentence[0])
	state1 := newTrellisState(model.Tag{Tag: startTag, Capital: false})
	state2 := newTrellisState(model.Tag{Tag: startTag, Capital: false})
	state2.backpointers[state1] = backpointer{nil, 0.0}
	trellis = append(trellis, state2)

	var beam float64

	// Loop through the tokens.
	for i := 2; i < len(sentence); i++ {
		columnHighestProb := math.Inf(-1)

		tagProbs := t.wordHandler.TagProbs(sentence[i])
		if len(tagProbs) == 0 {
			panic(fmt.Sprintf("No tag probabilities for: %s", sentence[i]))
		}

		for tag, tagProb := range tagProbs {
			state := newTrellisState(tag)

			// Loop over all possible trigrams
			for _, t2 := range trellis {
				highestProb := math.Inf(-1)
				var highestProbBP *trellisState

				for t1, t1bp := range t2.backpointers {
					if t1bp.prob < beam {
						continue
					}

					curTriGram := model.Trigram{T1: t1.tag, T2: t2.tag, T3: tag}
					trigramProb := t.trigramModel.TrigramProb(curTriGram)
					prob := trigramProb + tagProb + t1bp.prob

					if prob > highestProb {
						highestProb = prob
						highestProbBP = t1
					}
				}

				state.backpointers[t2] = backpointer{highestProbBP, highestProb}

				if highestProb > columnHighestProb {
					columnHighestProb = highestProb
				}
			}

			nextTrellis = append(nextTrellis, state)
		}

		// Swap 'trelli', recycling the old trellis by setting the lenght to 0.
		trellis, nextTrellis = nextTrellis, trellis[:0]
		beam = columnHighestProb - t.beamFactor
	}

	return trellis
}

func reverse(data []uint) {
	n := len(data)
	for i := 0; i < n/2; i++ {
		data[i], data[n-1-i] = data[n-1-i], data[i]
	}
}
