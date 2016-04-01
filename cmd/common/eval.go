package common

import (
	"fmt"

	"github.com/danieldk/citar/model"
	"github.com/danieldk/citar/tagger"
	"github.com/danieldk/conllx"
)

// The Evaluator type is used to keep counts on the number of
// correctly/incorrectly tagged known/unknown tokens.
type Evaluator struct {
	tagger           tagger.HMMTagger
	model            model.Model
	knownCorrect     uint
	knownIncorrect   uint
	unknownCorrect   uint
	unknownIncorrect uint
}

// NewEvaluator creates an evaluator that uses the provided tagger and
// the corresponding model. The model is used to distinguish known and
// unknown tokens.
func NewEvaluator(tagger tagger.HMMTagger, model model.Model) *Evaluator {
	return &Evaluator{
		tagger: tagger,
		model:  model,
	}
}

// Process a sentence, tagging it using the Evaluator's tagger and counting
// the number of tokens that were tagged correctly.
func (e *Evaluator) Process(sent []conllx.Token) error {
	words := make([]string, 0, len(sent))
	for _, token := range sent {
		if form, ok := token.Form(); ok {
			words = append(words, form)
		} else {
			return fmt.Errorf("Token does not have a form: %s", token)
		}
	}

	tags, _ := e.tagger.Tag(words).Tags()

	for idx, token := range sent {
		_, inLexicon := e.model.WordTagFreqs()[words[idx]]

		correctTag, ok := token.PosTag()
		if !ok {
			return fmt.Errorf("Token does not have a tag: %s", token)
		}

		if tags[idx] == correctTag {
			if inLexicon {
				e.knownCorrect++
			} else {
				e.unknownCorrect++
			}
		} else {
			if inLexicon {
				e.knownIncorrect++
			} else {
				e.unknownIncorrect++
			}
		}

	}

	return nil
}

// KnownCorrect returns the number of correctly tagged known words.
func (e *Evaluator) KnownCorrect() uint {
	return e.knownCorrect
}

// KnownIncorrect returns the number of incorrectly tagged known words.
func (e *Evaluator) KnownIncorrect() uint {
	return e.knownIncorrect
}

// UnknownCorrect returns the number of correctly tagged unknown words.
func (e *Evaluator) UnknownCorrect() uint {
	return e.unknownCorrect
}

// UnknownIncorrect returns the number of incorrectly tagged unknown words.
func (e *Evaluator) UnknownIncorrect() uint {
	return e.unknownIncorrect
}

// OverallCorrect returns the number of correctly tagged words.
func (e *Evaluator) OverallCorrect() uint {
	return e.knownCorrect + e.unknownCorrect
}

// OverallIncorrect returns the number of incorrectly tagged words.
func (e *Evaluator) OverallIncorrect() uint {
	return e.knownIncorrect + e.unknownIncorrect
}

// KnownAccuracy returns the tagging accuracy of known words.
func (e *Evaluator) KnownAccuracy() float64 {
	return float64(e.KnownCorrect()) / float64(e.KnownCorrect()+e.KnownIncorrect())
}

// Accuracy returns the tagging accuracy.
func (e *Evaluator) Accuracy() float64 {
	return float64(e.OverallCorrect()) / float64(e.OverallCorrect()+e.OverallIncorrect())
}

// UnknownAccuracy returns the tagging accuracy of unknown words.
func (e *Evaluator) UnknownAccuracy() float64 {
	return float64(e.UnknownCorrect()) / float64(e.UnknownCorrect()+e.UnknownIncorrect())
}
