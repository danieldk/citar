// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package model

// Tag represents a part of speech tag. The Capital field is used to
// mark whether the corresponding word started with a capital letter.
type Tag struct {
	Tag     uint
	Capital bool
}

// Unigram stores a tag unigram.
type Unigram struct {
	T1 Tag
}

// Bigram stores a tag bigram.
type Bigram struct {
	T1 Tag
	T2 Tag
}

// Trigram stores a tag trigram.
type Trigram struct {
	T1 Tag
	T2 Tag
	T3 Tag
}
