package words

import "github.com/danieldk/citar/model"

type wordSuffixTree struct {
	unigramFreqs map[model.Unigram]int
	root         *treeNode
	maxLength    int
	theta        float64
}

func newWordSuffixTree(unigramFreqs map[model.Unigram]int,
	skip map[uint]interface{}, theta float64,
	maxLength int) *wordSuffixTree {
	root := newTreeNode()

	// Populate root node tag frequencies
	for unigram, freq := range unigramFreqs {
		if _, ok := skip[unigram.T1.Tag]; ok {
			continue
		}

		root.tagFreqs[unigram.T1] = freq
		root.tagFreq += freq
	}

	return &wordSuffixTree{
		unigramFreqs: unigramFreqs,
		root:         root,
		maxLength:    maxLength,
		theta:        theta,
	}
}

func (t wordSuffixTree) addWord(word string, tf map[model.Tag]int) {
	runes := []rune(word)
	reverse(runes)
	if len(runes) > t.maxLength {
		runes = runes[0:t.maxLength]
	}

	t.root.addSuffix(runes, tf)
}

func (t wordSuffixTree) suffixTagProbs(word string) map[model.Tag]float64 {
	runes := []rune(word)
	reverse(runes)
	if len(runes) > t.maxLength {
		runes = runes[0:t.maxLength]
	}

	tagProbs := make(map[model.Tag]float64)
	for tag := range t.root.tagFreqs {
		tagProbs[tag] = 0
	}

	return t.root.suffixTagProbs(t.theta, t.unigramFreqs, runes, tagProbs)
}

func reverse(runes []rune) {
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
}

type treeNode struct {
	children map[rune]*treeNode
	tagFreqs map[model.Tag]int
	tagFreq  int
}

func newTreeNode() *treeNode {
	return &treeNode{
		children: make(map[rune]*treeNode),
		tagFreqs: make(map[model.Tag]int),
	}
}

func (n *treeNode) addSuffix(revSuffix []rune, tf map[model.Tag]int) {
	// Add the tag frequencies to the current node.
	for tag, freq := range tf {
		if _, ok := n.tagFreqs[tag]; ok {
			n.tagFreqs[tag] += freq
		} else {
			n.tagFreqs[tag] = freq
		}

		n.tagFreq += freq
	}

	// If the suffix is fully processed, we reached the final
	// state for this suffix.
	if len(revSuffix) == 0 {
		return
	}

	// Add transition.
	child, ok := n.children[revSuffix[0]]
	if !ok {
		child = newTreeNode()
		n.children[revSuffix[0]] = child
	}

	child.addSuffix(revSuffix[1:], tf)
}

func (n *treeNode) suffixTagProbs(theta float64, uf map[model.Unigram]int, revSuffix []rune,
	tp map[model.Tag]float64) map[model.Tag]float64 {
	for tag, prob := range tp {
		var nodeProb float64
		if f, ok := n.tagFreqs[tag]; ok {
			nodeProb = float64(f) / float64(n.tagFreq)
		}

		// Add weighted probability of the shorter suffixes.
		nodeProb += theta * prob

		// Normalize
		nodeProb /= theta + 1

		tp[tag] = nodeProb
	}

	// If the remaining suffix length is zero, we reached the final
	// state for this suffix.
	if len(revSuffix) == 0 {
		bayesianInversion(uf, tp)
		return tp
	}

	// Transition on the next suffix character.
	if child, ok := n.children[revSuffix[0]]; ok {
		return child.suffixTagProbs(theta, uf, revSuffix[1:], tp)
	}

	bayesianInversion(uf, tp)
	return tp
}

func bayesianInversion(uf map[model.Unigram]int, tp map[model.Tag]float64) {
	for tag, prob := range tp {
		tp[tag] = prob / float64(uf[model.Unigram{T1: tag}])
	}
}
