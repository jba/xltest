// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

package xltest

import (
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
		"add": func(a, b float64) float64 { return a + b },
		"lookup": func(s string) string {
			if s != "" {
				return s
			}
			return os.Getenv("XLTEST")
		},
	})
}
