// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

package xltest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

var testdataDir = filepath.FromSlash("../../testdata")

func TestRun(t *testing.T) {

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
		{
			"errors",
			strconv.Atoi,
			func(got, want any) error {
				ok := false
				switch got := got.(type) {
				case int:
					ok = got == want
				case error:
					var nerr *strconv.NumError
					ok = errors.As(got, &nerr)
				}
				if !ok {
					return fmt.Errorf("got %v, want %v", got, want)
				}
				return nil
			},
		},
	} {
		tst, err := ReadFile(filepath.Join(testdataDir, test.file+".yaml"))
		if err != nil {
			t.Fatal(err)
		}
		tst.Run(t, test.testFunc, test.validateFunc)
	}
}

func TestReadDir(t *testing.T) {
	got, err := ReadDir(testdataDir)
	if err != nil {
		t.Fatal(err)
	}

	// We can't actually run these because they require different test functions.

	checkName := func(tst *Test, wantName string) {
		t.Helper()
		if g, w := tst.Name, wantName; g != w {
			t.Errorf("got %q, want %q", g, w)
		}
	}

	checkName(got, "testdata")
	if g, w := len(got.SubTests), 4; g != w {
		t.Fatalf("got %d subtests, want %d", g, w)
	}
	checkName(got.SubTests[0], "add")
	checkName(got.SubTests[1], "env")
	checkName(got.SubTests[2], "errors")
	checkName(got.SubTests[3], "validate")
}
