// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package words

import (
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/danieldk/citar/model"
)

var _ WordHandler = Lexicon{}

type wordTagProbs map[string]map[model.Tag]float64

// Lexicon is an emission probability estimator for 'known words' (words
// seen in the training data).
type Lexicon struct {
	wordTagProbs wordTagProbs
	fallback     WordHandler
}

// NewLexicon constructs a new Lexicon from word/tag frequencies and unigram
// frequencies.
func NewLexicon(wtf map[string]map[model.Tag]int, uf map[model.Unigram]int) Lexicon {
	return Lexicon{
		wordTagProbs: calculateWordTagProbs(wtf, uf),
		fallback:     nil,
	}
}

// NewLexiconWithFallback construct a new Lexicon from word/tag frequencies,
// unigram frequencies, and a fallback. The fallback is used to estimate the
// emission probabilities when the word is not in the lexicon. For instance,
// this permits use of Lexicon with SuffixHandler to estimate the emission
// probability for any word.
func NewLexiconWithFallback(wtf map[string]map[model.Tag]int, uf map[model.Unigram]int, fallback WordHandler) Lexicon {
	return Lexicon{
		wordTagProbs: calculateWordTagProbs(wtf, uf),
		fallback:     fallback,
	}
}

// TagProbs returns P(w|t) for a particular word 'w'. Probabilities are only
// returned for tags with which the word occurred in the training data, except
// if the word did not occur in the training data and a fallback is used.
func (l Lexicon) TagProbs(word string) map[model.Tag]float64 {
	// Lookup word. If it is known, return P(w|t) for each tag that
	// the word was seen with in the training model.
	if probs, ok := l.wordTagProbs[word]; ok {
		return probs
	}

	// If the word could not be found, maybe its lowercase variant can
	// be found (e.g. capitalized words that start a sentence).
	runes := []rune(word)
	if unicode.IsUpper(runes[0]) {
		if probs, ok := l.wordTagProbs[strings.ToLower(word)]; ok {
			return probs
		}
	}

	// Try the fallback word handler, if it is available.
	if l.fallback != nil {
		return l.fallback.TagProbs(word)
	}

	return make(map[model.Tag]float64)
}

func calculateWordTagProbs(wtf map[string]map[model.Tag]int, uf map[model.Unigram]int) wordTagProbs {
	probs := make(wordTagProbs)

	for word, counts := range wtf {
		if _, ok := probs[word]; !ok {
			probs[word] = make(map[model.Tag]float64)
		}

		for tag, freq := range counts {
			// P(w|t) = f(w,t) / f(t)
			p := math.Log(float64(freq) / float64(uf[model.Unigram{T1: tag}]))
			probs[word][tag] = p
		}
	}

	return probs
}

type Substitution struct {
	Pattern     *regexp.Regexp
	Replacement string
}

type SubstLexicon struct {
	lexicon       Lexicon
	substitutions []Substitution
	fallback      WordHandler
}

// NewSubstLexicon construct a new Lexicon with substitution rules from a
// lexicon. If the lexicon does not return results for a word, the
// substitutions are applied and another lookup is attempted.
func NewSubstLexicon(lexicon Lexicon, substitutions []Substitution) SubstLexicon {
	return SubstLexicon{
		lexicon:       lexicon,
		substitutions: substitutions,
		fallback:      nil,
	}
}

// NewSubstLexiconWithFallback construct a new Lexicon with substitution rules
// from a lexicon and a fallback. If the lexicon does not return results for
// a word, the substitutions are applied and another lookup is attempted. If
// this fails as well, the fallback is used.
func NewSubstLexiconWithFallback(lexicon Lexicon, fallback WordHandler, substitutions []Substitution) SubstLexicon {
	return SubstLexicon{
		lexicon:       lexicon,
		substitutions: substitutions,
		fallback:      fallback,
	}
}

// TagProbs returns P(w|t) for a particular word 'w'. Probabilities are only
// returned for tags with which the word occurred in the training data, except
// if the word did not occur in the training data and a fallback is used.
func (l SubstLexicon) TagProbs(word string) map[model.Tag]float64 {
	probs := l.lexicon.TagProbs(word)
	if len(probs) != 0 {
		return probs
	}

	substWord := word
	// Attempt substitutions
	for _, subst := range l.substitutions {
		substWord = subst.Pattern.ReplaceAllString(substWord, subst.Replacement)
	}

	probs = l.lexicon.TagProbs(substWord)
	if len(probs) != 0 {
		return probs
	}

	// Try the fallback word handler, if it is available.
	if l.fallback != nil {
		return l.fallback.TagProbs(word)
	}

	return probs
}
