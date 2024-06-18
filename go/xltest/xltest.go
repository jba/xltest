// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that
// can be found in the LICENSE file.

package xltest

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

// A Test validates the result of a function on some input.
// See the [JSON schema] for documentation.
//
// [JSON schema]: https://github.com/jba/xltest/blob/main/test-schema.yaml
type Test[Input, Want any] struct {
	Name        string               `yaml:"name,omitempty"`
	Description string               `yaml:"description,omitempty"`
	Env         map[string]string    `yaml:"env,omitempty"`
	Input       *Input               `yaml:"in,omitempty"`
	Want        *Want                `yaml:"want,omitempty"`
	OnError     string               `yaml:"onError,omitempty"`
	SubTests    []*Test[Input, Want] `yaml:"subtests,omitempty"`
}

// Init initializes and validates a test and its subtests.
// It is only necessary for tests that have been constructed in memory; [ReadFile]
// and [ReadDir] call it themselves.
func (tst *Test[I, W]) Init(name string) error {
	if tst.Name == "" {
		if name == "" {
			return errors.New("no name for top-level test")
		}
		tst.Name = name
	}
	var errs []error
	tst.init("", func(msg string) {
		errs = append(errs, errors.New(msg))
	})
	return errors.Join(errs...)
}

func (tst *Test[I, W]) init(prefix string, addMsg func(string)) {
	prefix = path.Join(prefix, tst.Name)

	if tst.Input == nil && tst.Want != nil {
		addMsg(fmt.Sprintf("%s: test has 'want' but not 'in'", prefix))
	}
	if tst.Input == nil && len(tst.SubTests) == 0 {
		addMsg(fmt.Sprintf("%s: test has no input and no subtests", prefix))
	}
	for i, st := range tst.SubTests {
		if st.Name == "" {
			st.Name = fmt.Sprint(i)
		}
		st.init(prefix, addMsg)
	}
}

const ( // onError values
	fail     = "fail"
	succeed  = "succeed"
	validate = "validate"
)

// Run runs the test with the given functions.
//
// testFunc is the function under test. It takes one argument whose type matches
// the type of the inputs declared in the test. Its first return value is validated
// against the "want" field of each test (if any) by the validation function.
// If testFunc returns a non-nil error, the test fails immediately.
//
// validateFunc validates the result of testFunc. If non-nil, it takes two arguments:
// the first is the value returned by testFunc, and the second is the value of the
// test's "want" field (or the zero value of the Want type, if the want field is
// omitted). It should return a non-nil error if the test should fail.
// If validateFunc is nil, the actual and expected values will be compared for (deep)
// equality using [github.com/google/go-cmp/cmp.Equal].
//
// The third function validates errors when the onError field of the test is "validate".
// It is an error to pass nil for it in that case.
func (tst *Test[I, W]) Run(
	t *testing.T,
	testFunc func(I) (W, error),
	validateFunc func(W, W) error,
	errValidateFunc func(error, W) error,
) {
	if validateFunc == nil {
		validateFunc = func(got, want W) error {
			if cmp.Equal(got, want) {
				return nil
			}
			return fmt.Errorf("got %v, want %v", got, want)
		}
	}
	tst.run(t, testFunc, validateFunc, errValidateFunc, fail)
}

func (tst *Test[I, W]) run(
	t *testing.T,
	testFunc func(I) (W, error),
	validateFunc func(W, W) error,
	errValidateFunc func(error, W) error,
	onError string,
) {
	t.Run(tst.Name, func(t *testing.T) {
		for name, value := range tst.Env {
			t.Setenv(name, value)
		}
		// Override the inherited onError value with the one specified by this test.
		if tst.OnError != "" {
			onError = tst.OnError
		}
		if tst.Input != nil {
			got, err := testFunc(*tst.Input)
			var want W
			if tst.Want != nil {
				want = *tst.Want
			}
			switch onError {
			case fail:
				if err != nil {
					t.Errorf("test function: %v", err)
				} else if err := validateFunc(got, want); err != nil {
					t.Error(err)
				}
			case succeed:
				if err == nil {
					t.Error("test function returned nil, wanted error")
				}
			case validate:
				if errValidateFunc == nil {
					t.Fatal("test onError=validate, but error validation function is nil")
				}
				if err := errValidateFunc(err, want); err != nil {
					t.Error(err)
				}
			default:
				panic("bad onError value")
			}
		}
		for _, test := range tst.SubTests {
			test.run(t, testFunc, validateFunc, errValidateFunc, onError)
		}
	})
}

// ReadFile reads a Test from a YAML file.
// If the test doesn't have a name, it is named after the
// last component of the filename, excluding the extension.
func ReadFile[I, W any](filename string) (*Test[I, W], error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tst Test[I, W]
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&tst); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	cname := filepath.Clean(filename)
	defaultName := strings.TrimSuffix(filepath.Base(cname), filepath.Ext(cname))
	if err := tst.Init(defaultName); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return &tst, nil
}

// ReadDir calls ReadFile on all the .yaml files in dir.
// The resulting Tests become sub-tests of the returned Test,
// whose name is the last component of dir.
func ReadDir[I, W any](dir string) (*Test[I, W], error) {
	filenames, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var subTests []*Test[I, W]
	for _, fn := range filenames {
		st, err := ReadFile[I, W](fn)
		if err != nil {
			return nil, err
		}
		subTests = append(subTests, st)
	}
	return &Test[I, W]{
		Name:        filepath.Base(filepath.Clean(dir)),
		Description: fmt.Sprintf("test files from %s", dir),
		SubTests:    subTests,
	}, nil
}
