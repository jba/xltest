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
	dir := filepath.Join("..", "..", "testdata")

	for _, test := range []struct {
		file         string
		testFunc     any
		validateFunc any
	}{
		{
			"add",
			func(input []any) int { return input[0].(int) + input[1].(int) },
			nil,
		},
		{
			"env",
			func(s string) string {
				if s != "" {
					return s
				}
				return os.Getenv("XLTEST")
			},
			nil,
		},
		{
			"validate",
			func(int) int { return rand.IntN(3) },
			func(got int, _ any) string {
				if got < 3 {
					return ""
				}
				return fmt.Sprintf("got %v, want a number less than 3", got)
			},
		},
	} {
		tst, err := ReadFile(filepath.Join(dir, test.file+".yaml"))
		if err != nil {
			t.Fatal(err)
		}
		tst.Run(t, test.testFunc, test.validateFunc)
	}
}
