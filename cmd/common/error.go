// Copyright 2016 The Citar Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

import (
	"fmt"
	"os"
)

// ExitIfError exits a program with a fatal error message, if
// // the supplied error is not nil.
func ExitIfError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", prefix, err.Error())
		os.Exit(1)
	}
}
