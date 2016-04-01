// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package trigrams

import "github.com/danieldk/citar/model"

// A TrigramModel estimates transition probabilities using trigrams,
// p(t3|t1,t2).
type TrigramModel interface {
	TrigramProb(trigram model.Trigram) float64
}
