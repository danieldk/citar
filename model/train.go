package model

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/danieldk/conllx"
)

// A FrequencyCollector collects frequencies from the training corpus that
// are relevant to a trigram HMM tagger.
type FrequencyCollector struct {
	numberer *StringNumberer
	lexicon  map[string]map[Tag]int
	unigrams map[Unigram]int
	bigrams  map[Bigram]int
	trigrams map[Trigram]int
}

// NewFrequencyCollector constructs a FrequencyCollector instance.
func NewFrequencyCollector() FrequencyCollector {
	return FrequencyCollector{
		numberer: NewStringStringNumberer(),
		lexicon:  make(map[string]map[Tag]int),
		unigrams: make(map[Unigram]int),
		bigrams:  make(map[Bigram]int),
		trigrams: make(map[Trigram]int),
	}
}

// Model returns the collected frequencies as a model.
func (c FrequencyCollector) Model() Model {
	return newModel(c.numberer, c.lexicon, c.unigrams, c.bigrams, c.trigrams)
}

// Process a sentence.
func (c FrequencyCollector) Process(sentence []conllx.Token) error {
	sentence = c.addMarkers(sentence)

	wordTags, err := c.sentenceToWordTags(sentence)
	if err != nil {
		return err
	}

	for i := 0; i < len(wordTags); i++ {
		if err := c.addLexiconEntry(wordTags[i]); err != nil {
			return err
		}

		c.addUnigram(wordTags[i])
		if i > 0 {
			c.addBigram(wordTags[i-1], wordTags[i])
		}
		if i > 1 {
			c.addTrigram(wordTags[i-2], wordTags[i-1], wordTags[i])
		}
	}

	return nil
}

type wordTag struct {
	word    string
	tag     uint
	isUpper bool
}

func (c FrequencyCollector) sentenceToWordTags(sentence []conllx.Token) ([]wordTag, error) {
	wordTags := make([]wordTag, 0, len(sentence))

	for _, token := range sentence {
		form, ok := token.Form()
		if !ok {
			return nil, fmt.Errorf("token does not contain a form: %s", token)
		}

		pos, ok := token.PosTag()
		if !ok {
			return nil, fmt.Errorf("token does not contain a part-of-speech: %s", token)
		}

		first, _ := utf8.DecodeRuneInString(form)
		if first == utf8.RuneError {
			return nil, fmt.Errorf("invalid UTF-8 character in form: %s", form)
		}

		wordTags = append(wordTags, wordTag{form, c.numberer.Number(pos), unicode.IsUpper(first)})
	}

	return wordTags, nil
}

func (c FrequencyCollector) addLexiconEntry(wordTag wordTag) error {
	tagFreqs, ok := c.lexicon[wordTag.word]
	if !ok {
		tagFreqs = make(map[Tag]int)
		c.lexicon[wordTag.word] = tagFreqs
	}

	tagFreqs[Tag{wordTag.tag, wordTag.isUpper}]++

	return nil
}

func (c FrequencyCollector) addBigram(wordTag wordTag, wordTag2 wordTag) {
	c.bigrams[Bigram{
		T1: Tag{wordTag.tag, wordTag.isUpper},
		T2: Tag{wordTag2.tag, wordTag2.isUpper},
	}]++
}

func (c FrequencyCollector) addTrigram(wordTag wordTag, wordTag2 wordTag, wordTag3 wordTag) {
	c.trigrams[Trigram{
		T1: Tag{wordTag.tag, wordTag.isUpper},
		T2: Tag{wordTag2.tag, wordTag2.isUpper},
		T3: Tag{wordTag3.tag, wordTag3.isUpper},
	}]++
}

func (c FrequencyCollector) addUnigram(wordTag wordTag) {
	c.unigrams[Unigram{
		T1: Tag{wordTag.tag, wordTag.isUpper},
	}]++
}

func (c FrequencyCollector) addMarkers(sentence []conllx.Token) []conllx.Token {
	startToken := conllx.NewToken()
	startToken.SetForm(StartToken)
	startToken.SetPosTag(StartToken)

	endToken := conllx.NewToken()
	endToken.SetForm(EndToken)
	endToken.SetPosTag(EndToken)

	start := []conllx.Token{*startToken, *startToken}
	sentence = append(start, sentence...)
	sentence = append(sentence, *endToken)

	return sentence
}

func (c FrequencyCollector) addCapitalTags(sentence []conllx.Token) error {
	for i := 0; i < len(sentence); i++ {
		form, ok := sentence[i].Form()
		pos, pok := sentence[i].PosTag()

		if ok && pok {
			first, _ := utf8.DecodeRuneInString(form)
			if first == utf8.RuneError {
				return fmt.Errorf("invalid UTF-8 character in form: %s", form)
			}

			if unicode.IsUpper(first) {
				sentence[i].SetPosTag(fmt.Sprintf("c-%s", pos))
			} else {
				sentence[i].SetPosTag(fmt.Sprintf("n-%s", pos))
			}
		} else {
			return fmt.Errorf("token does not contain form or part-of-speech: %s", sentence[i])
		}
	}

	return nil
}
