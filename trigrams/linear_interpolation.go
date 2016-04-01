package trigrams

import (
	"fmt"
	"math"

	"github.com/danieldk/citar/model"
)

type unigramProbs map[model.Unigram]float64

type bigramProbs map[model.Bigram]float64

type trigramProbs map[model.Trigram]float64

// LinearInterpolationModel estimates transmission (trigram) probabilities
// using maximum likelihood estimation and linear interpolation smoothing
// (Brants, 2000).
type LinearInterpolationModel struct {
	unigramProbs unigramProbs
	bigramProbs  bigramProbs
	trigramProbs trigramProbs
}

// NewLinearInterpolationModel constructs a LinearInterpolation model from
// a data model.
func NewLinearInterpolationModel(model model.Model) LinearInterpolationModel {
	corpusSize := corpusSize(model.UnigramFreqs())
	smoothingParameters := calculateLambdas(corpusSize, model.UnigramFreqs(), model.BigramFreqs(),
		model.TrigramFreqs())

	return LinearInterpolationModel{
		unigramProbs: calcUnigramProbs(corpusSize, smoothingParameters, model.UnigramFreqs()),
		bigramProbs: calcBigramProbs(corpusSize, smoothingParameters,
			model.UnigramFreqs(), model.BigramFreqs()),
		trigramProbs: calcTrigramProbs(corpusSize, smoothingParameters,
			model.UnigramFreqs(), model.BigramFreqs(), model.TrigramFreqs()),
	}
}

// TrigramProb estimates transition probabilities using trigrams,
// p(t3|t1,t2).
func (m LinearInterpolationModel) TrigramProb(trigram model.Trigram) float64 {
	if p, ok := m.trigramProbs[trigram]; ok {
		return p
	}

	if p, ok := m.bigramProbs[model.Bigram{T1: trigram.T2, T2: trigram.T3}]; ok {
		return p
	}

	if p, ok := m.unigramProbs[model.Unigram{T1: trigram.T3}]; ok {
		return p
	}

	panic(fmt.Sprintf("Unknown tag: %v", trigram.T3))
}

func corpusSize(unigramFreqs map[model.Unigram]int) int {
	var size int

	for _, freq := range unigramFreqs {
		size += freq
	}

	return size
}

func calculateLambdas(corpusSize int, unigramFreqs map[model.Unigram]int, bigramFreqs map[model.Bigram]int,
	trigramFreqs map[model.Trigram]int) smoothingParameters {
	var l1f, l2f, l3f int

	for t1t2t3, t1t2t3Freq := range trigramFreqs {
		t1t2 := model.Bigram{T1: t1t2t3.T1, T2: t1t2t3.T2}

		var l3p float64
		if t1t2Freq, ok := bigramFreqs[t1t2]; ok {
			l3p = float64(t1t2t3Freq-1) / float64(t1t2Freq-1)
		}

		t2t3 := model.Bigram{T1: t1t2t3.T2, T2: t1t2t3.T3}
		t2 := model.Unigram{T1: t1t2t3.T2}

		var l2p float64
		if t2t3Freq, ok := bigramFreqs[t2t3]; ok {
			if t2Freq, ok := unigramFreqs[t2]; ok {
				l2p = float64(t2t3Freq-1) / float64(t2Freq-1)
			}
		}

		t3 := model.Unigram{T1: t1t2t3.T3}
		var l1p float64
		if t3Freq, ok := unigramFreqs[t3]; ok {
			l1p = float64(t3Freq-1) / float64(corpusSize-1)
		}

		if l1p > l2p && l1p > l3p {
			l1f += t1t2t3Freq
		} else if l2p > l1p && l2p > l3p {
			l2f += t1t2t3Freq
		} else {
			l3f += t1t2t3Freq
		}
	}

	totalTrigrams := l1f + l2f + l3f

	return smoothingParameters{
		L1: float64(l1f) / float64(totalTrigrams),
		L2: float64(l2f) / float64(totalTrigrams),
		L3: float64(l3f) / float64(totalTrigrams),
	}
}

type smoothingParameters struct {
	L1 float64
	L2 float64
	L3 float64
}

func calcUnigramProbs(corpusSize int, smoothingParameters smoothingParameters,
	unigramFreqs map[model.Unigram]int) unigramProbs {
	probs := make(unigramProbs)

	for unigram, freq := range unigramFreqs {
		prob := float64(freq) / float64(corpusSize)

		// Smooth and transform to log-space
		prob = math.Log(smoothingParameters.L1 * prob)

		probs[unigram] = prob
	}

	return probs
}

func calcBigramProbs(corpusSize int, smoothingParameters smoothingParameters,
	unigramFreqs map[model.Unigram]int, bigramFreqs map[model.Bigram]int) bigramProbs {
	probs := make(bigramProbs)

	for bigram, freq := range bigramFreqs {
		t2 := model.Unigram{T1: bigram.T2}

		// Unigram likelihood P(t2)
		unigramProb := float64(unigramFreqs[t2]) / float64(corpusSize)

		// Bigram likelihood P(t2|t1).
		t1 := model.Unigram{T1: bigram.T1}
		t1Freq := unigramFreqs[t1]
		bigramProb := float64(freq) / float64(t1Freq)

		// Smooth and transform to log-space
		prob := math.Log(smoothingParameters.L1*unigramProb + smoothingParameters.L2*bigramProb)

		probs[bigram] = prob
	}

	return probs
}

func calcTrigramProbs(corpusSize int, smoothingParameters smoothingParameters,
	unigramFreqs map[model.Unigram]int, bigramFreqs map[model.Bigram]int,
	trigramFreqs map[model.Trigram]int) trigramProbs {
	probs := make(trigramProbs)

	for trigram, freq := range trigramFreqs {
		// Unigram likelihood P(t3)
		t3 := model.Unigram{T1: trigram.T3}
		unigramProb := float64(unigramFreqs[t3]) / float64(corpusSize)

		// Bigram likelihood P(t3|t2).
		t2t3 := model.Bigram{T1: trigram.T2, T2: trigram.T3}
		t2 := model.Unigram{T1: trigram.T2}
		bigramProb := float64(bigramFreqs[t2t3]) / float64(unigramFreqs[t2])

		t1t2 := model.Bigram{T1: trigram.T1, T2: trigram.T2}
		trigramProb := float64(freq) / float64(bigramFreqs[t1t2])

		prob := math.Log(smoothingParameters.L1*unigramProb +
			smoothingParameters.L2*bigramProb +
			smoothingParameters.L3*trigramProb)

		probs[trigram] = prob

	}

	return probs
}
