package words

import "github.com/danieldk/citar/model"

// A WordHandler returns or estimates the emission probabilities P(w|t) for
// a given words.
type WordHandler interface {
	TagProbs(word string) map[model.Tag]float64
}
