package trigrams

import "github.com/danieldk/citar/model"

// A TrigramModel estimates transition probabilities using trigrams,
// p(t3|t1,t2).
type TrigramModel interface {
	TrigramProb(trigram model.Trigram) float64
}
