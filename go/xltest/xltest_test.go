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
	tst := mustReadFile[[2]int, int](t, "add")
	tst.Run(t, func(input [2]int) (int, error) { return input[0] + input[1], nil }, nil, nil)

	tst2 := mustReadFile[string, string](t, "env")
	testfunc := func(s string) (string, error) {
		if s != "" {
			return s, nil
		}
		return os.Getenv("XLTEST"), nil
	}
	tst2.Run(t, testfunc, nil, nil)

	tst3 := mustReadFile[string, string](t, "validate")
	tst3.Run(t,
		func(s string) (string, error) { return "LLM says: " + s, nil },
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
		nil)

	tst4 := mustReadFile[string, int](t, "errors")
	tst4.Run(t, strconv.Atoi, nil, func(got error, _ int) error {
		var nerr *strconv.NumError
		if !errors.As(got, &nerr) {
			return fmt.Errorf("got error of type %T, want strconv.NumError", got)
		}
		return nil
	})
}

func mustReadFile[I, W any](t *testing.T, name string) *Test[I, W] {
	tst, err := ReadFile[I, W](filepath.Join(testdataDir, name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return tst
}

func (t *Test[I, W]) name() string { return t.Name }

func TestReadDir(t *testing.T) {
	got, err := ReadDir[any, any](testdataDir)
	if err != nil {
		t.Fatal(err)
	}

	// We can't actually run these because they require different test functions.

	checkName := func(gotName, wantName string) {
		t.Helper()
		if g, w := gotName, wantName; g != w {
			t.Errorf("got %q, want %q", g, w)
		}
	}

	checkName(got.Name, "testdata")
	if g, w := len(got.SubTests), 4; g != w {
		t.Fatalf("got %d subtests, want %d", g, w)
	}
	checkName(got.SubTests[0].Name, "add")
	checkName(got.SubTests[1].Name, "env")
	checkName(got.SubTests[2].Name, "errors")
	checkName(got.SubTests[3].Name, "validate")
}
