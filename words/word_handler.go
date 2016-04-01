// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package words

import "github.com/danieldk/citar/model"

// A WordHandler returns or estimates the emission probabilities P(w|t) for
// a given words.
type WordHandler interface {
	TagProbs(word string) map[model.Tag]float64
}
