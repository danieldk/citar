// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package words provides methods to estimate (word) emission probabilities.
//
// The parameters in Hidden Markov Models (HMM) come in two forms: transition and
// emission probabilities. In a trigram HMM tagger, the transition probabilities
// are P(t3|t1,t2) and the emission probabilities P(w|t), where 'w' is a word and
// 't' a tag.
//
// This package concerns itself with estimating emission probabilities. Generally,
// the emission probabilities are estimated as follows: (1) for words seen in the
// training data, probability is the (smoothed) maximum likelihood estimation;
// (2) for words that are not seen in the training data the probabilies are usually
// estimated based on inflectional properties.
//
// The `Lexicon` type implements (1), while the SuffixHandler type is a possible
// implementation of (2) based on Brants, 2000. Both types implement the
// WordHandler interface.
package words
