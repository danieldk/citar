package words

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/danieldk/citar/model"
)

var _ WordHandler = SuffixHandler{}

// SuffixHandler is an emission probability estimator that uses word suffices.
// It is normally used for words that were not seen in the training model.
//
// Internally, this estimator uses four different distributions based properties
// of the token: (1) Tokens that start with an uppercase letter; (2) tokens that
// contain a dash (currently only '-'); (3) tokens that are recognized as
// cardinals; and (4) remaining tokens (typically lowercase words).
type SuffixHandler struct {
	upperTree    *wordSuffixTree
	lowerTree    *wordSuffixTree
	dashTree     *wordSuffixTree
	cardinalTree *wordSuffixTree
	maxTags      int
}

// SuffixHandlerConfig stores the configuration for a SuffixHandler. It
// allows specification of the length of the suffix to be considered,
// maximum frequencies of tokens in order to be used as training data,
// and the maximum number of tags that a SuffixHandler should return p(w|t)
// for.
//
// Tweaking this parameters can have a profound effect on the quality if
// the estimator. For instance, the typical length of inflectional suffixes
// is highly language-dependent. Good values for the maximum frequencies for
// the various types of tokens depends on the size of the training corpus -
// the distribution of unknown words is typically closer to that of
// low-frequency words than high-frequency words.
type SuffixHandlerConfig struct {
	MaxSuffixLen    int
	UpperMaxFreq    int
	LowerMaxFreq    int
	DashMaxFreq     int
	CardinalMaxFreq int
	MaxTags         int
}

// DefaultSuffixHandlerConfig returns a SuffixHandlerConfig that works reasonably
// well on German and English with approximately 50,000 to 100,000 sentences.
func DefaultSuffixHandlerConfig() SuffixHandlerConfig {
	return SuffixHandlerConfig{
		MaxSuffixLen:    2,
		UpperMaxFreq:    2,
		LowerMaxFreq:    8,
		DashMaxFreq:     4,
		MaxTags:         10,
		CardinalMaxFreq: 10,
	}
}

// NewSuffixHandler constructs a new SuffixHandler from the given configuration
// and model.
func NewSuffixHandler(config SuffixHandlerConfig, m model.Model) SuffixHandler {
	skip := make(map[uint]interface{})

	skip[m.TagNumberer().Number(model.StartToken)] = nil
	skip[m.TagNumberer().Number(model.EndToken)] = nil

	theta := calcTheta(m.UnigramFreqs(), skip)

	upperTree := newWordSuffixTree(m.UnigramFreqs(), skip, theta, config.MaxSuffixLen)
	lowerTree := newWordSuffixTree(m.UnigramFreqs(), skip, theta, config.MaxSuffixLen)
	dashTree := newWordSuffixTree(m.UnigramFreqs(), skip, theta, config.MaxSuffixLen)
	cardinalTree := newWordSuffixTree(m.UnigramFreqs(), skip, theta, config.MaxSuffixLen)

	sh := SuffixHandler{
		upperTree:    upperTree,
		lowerTree:    lowerTree,
		dashTree:     dashTree,
		cardinalTree: cardinalTree,
		maxTags:      config.MaxTags,
	}

	for word, tagFreqs := range m.WordTagFreqs() {
		if word == model.StartToken || word == model.EndToken {
			continue
		}

		// Incorrect lexicon entry.
		if len(word) == 0 {
			continue
		}

		var wordFreq int
		for _, tagFreq := range tagFreqs {
			wordFreq += tagFreq
		}

		if t := sh.selectSuffixTreeWithCutoffs(config, word, wordFreq); t != nil {
			t.addWord(word, tagFreqs)
		}
	}

	return sh
}

type tagProb struct {
	tag  model.Tag
	prob float64
}

// TagProbs estimates P(w|t) for a particular word 'w'.
func (h SuffixHandler) TagProbs(word string) map[model.Tag]float64 {
	t := h.selectSuffixTree(word)

	return bestNLogSpace(t.suffixTagProbs(word), h.maxTags)
}

func (h SuffixHandler) selectSuffixTree(word string) *wordSuffixTree {
	runes := []rune(word)

	var t *wordSuffixTree
	if unicode.IsUpper(runes[0]) {
		t = h.upperTree
	} else if cardinalPattern.MatchString(word) {
		t = h.cardinalTree
	} else if strings.ContainsRune(word, '-') {
		t = h.dashTree
	} else {
		t = h.lowerTree
	}

	return t
}

func (h SuffixHandler) selectSuffixTreeWithCutoffs(config SuffixHandlerConfig, word string,
	wordFreq int) *wordSuffixTree {
	runes := []rune(word)

	var t *wordSuffixTree
	if unicode.IsUpper(runes[0]) {
		if wordFreq <= config.UpperMaxFreq {
			t = h.upperTree
		}
	} else if cardinalPattern.MatchString(word) {
		if wordFreq <= config.CardinalMaxFreq {
			t = h.cardinalTree
		}
	} else if strings.ContainsRune(word, '-') {
		if wordFreq <= config.DashMaxFreq {
			t = h.dashTree
		}
	} else {
		if wordFreq <= config.LowerMaxFreq {
			t = h.lowerTree
		}
	}

	return t
}

func calcTheta(uf map[model.Unigram]int, skip map[uint]interface{}) float64 {
	pAvg := 1. / float64(len(uf))

	freqSum := 0
	for unigram, freq := range uf {
		if _, ok := skip[unigram.T1.Tag]; ok {
			continue
		}

		freqSum += freq
	}

	var stddevSum float64
	for unigram, freq := range uf {
		if _, ok := skip[unigram.T1.Tag]; ok {
			continue
		}

		// P(t)
		p := float64(freq) / float64(freqSum)
		stddevSum += math.Pow(p-pAvg, 2.0)
	}

	return math.Sqrt(stddevSum / float64(len(uf)-1))
}

func insertWithLimit(slice []tagProb, limit, index int, value tagProb) []tagProb {
	if len(slice) < limit {
		slice = append(slice, tagProb{})
	}

	copy(slice[index+1:], slice[index:len(slice)-1])
	slice[index] = value
	return slice
}

var _ WordHandler = LookupSuffixHandler{}

type LookupSuffixHandler struct {
	upperProbs    map[string]map[model.Tag]float64
	lowerProbs    map[string]map[model.Tag]float64
	dashProbs     map[string]map[model.Tag]float64
	cardinalProbs map[string]map[model.Tag]float64
	maxLength     int
}

func NewLookupSuffixHandler(sh SuffixHandler) LookupSuffixHandler {
	return LookupSuffixHandler{
		upperProbs:    convertTree(sh.upperTree, sh.maxTags),
		lowerProbs:    convertTree(sh.lowerTree, sh.maxTags),
		dashProbs:     convertTree(sh.dashTree, sh.maxTags),
		cardinalProbs: convertTree(sh.cardinalTree, sh.maxTags),
		maxLength:     sh.lowerTree.maxLength,
	}
}

func (h LookupSuffixHandler) TagProbs(word string) map[model.Tag]float64 {
	m := h.selectMap(word)

	runes := []rune(word)
	reverse(runes)
	if len(runes) > h.maxLength {
		runes = runes[0:h.maxLength]
	}

	for len(runes) > 0 {
		if probs, ok := m[string(runes)]; ok {
			return probs
		}

		runes = runes[0 : len(runes)-1]
	}

	return m[string(runes)]
}

func (h LookupSuffixHandler) selectMap(word string) map[string]map[model.Tag]float64 {
	runes := []rune(word)

	if unicode.IsUpper(runes[0]) {
		return h.upperProbs
	} else if cardinalPattern.MatchString(word) {
		return h.cardinalProbs
	} else if strings.ContainsRune(word, '-') {
		return h.dashProbs
	}

	return h.lowerProbs
}

func convertTree(t *wordSuffixTree, maxTags int) map[string]map[model.Tag]float64 {
	probs := make(map[string]map[model.Tag]float64)

	tagProbs := make(map[model.Tag]float64)
	for tag := range t.root.tagFreqs {
		tagProbs[tag] = 0
	}

	convertTreeRecursive(t.root, maxTags, t.theta, t.unigramFreqs, "", tagProbs, probs)
	return probs
}

func convertTreeRecursive(n *treeNode, maxTags int, theta float64,
	uf map[model.Unigram]int, suffix string, tagProbs map[model.Tag]float64,
	probs map[string]map[model.Tag]float64) {

	for tag, prob := range tagProbs {
		var nodeProb float64
		if f, ok := n.tagFreqs[tag]; ok {
			nodeProb = float64(f) / float64(n.tagFreq)
		}

		// Add weighted probability of the shorter suffixes.
		nodeProb += theta * prob

		// Normalize
		nodeProb /= theta + 1

		tagProbs[tag] = nodeProb
	}

	for r, cn := range n.children {
		convertTreeRecursive(cn, maxTags, theta, uf, suffix+string(r),
			copyTagProbs(tagProbs), probs)
	}

	bayesianInversion(uf, tagProbs)
	probs[suffix] = bestNLogSpace(tagProbs, maxTags)
}

func bestNLogSpace(tp map[model.Tag]float64, n int) map[model.Tag]float64 {
	sorted := make([]tagProb, 0, n)

	for tag, prob := range tp {
		ip := sort.Search(len(sorted), func(i int) bool {
			return sorted[i].prob <= prob
		})

		if ip < n {
			sorted = insertWithLimit(sorted, n, ip, tagProb{tag, prob})
		}
	}

	results := make(map[model.Tag]float64)

	for _, tagProb := range sorted {
		results[tagProb.tag] = math.Log(tagProb.prob)
	}

	return results
}

func copyTagProbs(tp map[model.Tag]float64) map[model.Tag]float64 {
	tpc := make(map[model.Tag]float64)

	for k, v := range tp {
		tpc[k] = v
	}

	return tpc
}

var cardinalPattern = regexp.MustCompile(`^([0-9]+)|([0-9]+\.)|([0-9.,:-]+[0-9]+)|([0-9]+[a-zA-Z]{1,3})$`)
