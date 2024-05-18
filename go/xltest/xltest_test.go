// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

package xltest

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	tst, err := ReadDir(filepath.Join("..", "..", "testdata"))
	if err != nil {
		t.Fatal(err)
	}
	tst.Run(t, map[string]any{
		"add": func(a, b int) int { return a + b },
		"lookup": func(s string) string {
			if s != "" {
				return s
			}
			return os.Getenv("XLTEST")
		},
		"rand3": func() int { return rand.IntN(3) },
		"lessThan3": func(got, want any) string {
			if got.(int) < 3 {
				return ""
			}
			return fmt.Sprintf("got %v, want a number less than 3", got)
		},
	})
}
