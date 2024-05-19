// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

package xltest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
			func(s string) string { return "LLM says: " + s },
			func(got, wantRegexp string) error {
				matched, err := regexp.MatchString(wantRegexp, got)
				if err != nil {
					return err
				}
				if !matched {
					return fmt.Errorf("got %q, wanted match for %q", got, wantRegexp)
				}
				return nil
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
