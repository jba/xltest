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

type Test struct {
	Name        string            `yaml:"name,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	// Can be empty if this just holds subtests
	Input    any     `yaml:"in,omitempty"`
	Want     any     `yaml:"want,omitempty"`
	SubTests []*Test `yaml:"subtests,omitempty"`
}

func (tst *Test) Init(name string) error {
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

func (tst *Test) init(prefix string, addMsg func(string)) {
	prefix = path.Join(prefix, tst.Name)
	for i, st := range tst.SubTests {
		if st.Name == "" {
			st.Name = fmt.Sprint(i)
		}
		st.init(prefix, addMsg)
	}
}

func (tst *Test) Run(t *testing.T, testFunction, validateFunction any) {
	testFunc := makeTestFunc(testFunction)
	if testFunc == nil {
		t.Fatal("bad test function: want func(_) _ or func (_) (_, error)")
	}
	var validateFunc validateFuncType
	if validateFunction != nil {
		validateFunc = makeValidateFunc(validateFunction)
		if validateFunc == nil {
			t.Fatal("bad validate function: want func(_, _) error")
		}
	} else {
		validateFunc = func(got, want any) error {
			if cmp.Equal(got, want) {
				return nil
			}
			return fmt.Errorf("got %v, want %v", got, want)
		}
	}
	tst.run(t, testFunc, validateFunc)
}

func (tst *Test) run(t *testing.T, testFunc testFuncType, validateFunc validateFuncType) {
	t.Run(tst.Name, func(t *testing.T) {
		for name, value := range tst.Env {
			t.Setenv(name, value)
		}
		if tst.Input != nil {
			got, err := testFunc(tst.Input)
			if err != nil {
				t.Fatalf("test function: %v", err)
			}
			if err := validateFunc(got, tst.Want); err != nil {
				t.Error(err)
			}
		}
		for _, test := range tst.SubTests {
			test.run(t, testFunc, validateFunc)
		}
	})
}

type testFuncType func(any) (any, error)

func makeTestFunc(f any) testFuncType {
	fv := reflect.ValueOf(f)
	ft := fv.Type()
	if ft == nil {
		return nil
	}
	if ft.Kind() != reflect.Func {
		return nil
	}
	if ft.NumIn() != 1 {
		return nil
	}
	switch ft.NumOut() {
	case 1:
		return func(x any) (any, error) {
			rs := fv.Call([]reflect.Value{reflectValue(x)})
			return rs[0].Interface(), nil
		}
	case 2:
		if ft.Out(1) != reflect.TypeFor[error]() {
			return nil
		}
		return func(x any) (any, error) {
			rs := fv.Call([]reflect.Value{reflectValue(x)})
			return rs[0].Interface(), rs[1].Interface().(error)
		}
	default:
		return nil
	}
}

type validateFuncType func(any, any) error

func makeValidateFunc(f any) validateFuncType {
	fv := reflect.ValueOf(f)
	ft := fv.Type()
	if ft == nil {
		return nil
	}
	if ft.Kind() != reflect.Func {
		return nil
	}
	if ft.NumIn() != 2 {
		return nil
	}
	if ft.NumOut() != 1 {
		return nil
	}
	if ft.Out(0) != reflect.TypeFor[error]() {
		return nil
	}
	return func(x, y any) error {
		rs := fv.Call([]reflect.Value{reflectValue(x), reflectValue(y)})
		r := rs[0].Interface()
		if r == nil {
			return nil
		}
		return r.(error)
	}
}

func ReadFile(filename string) (*Test, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tst Test
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&tst); err != nil {
		return nil, err
	}

	cname := filepath.Clean(filename)
	defaultName := strings.TrimSuffix(filepath.Base(cname), filepath.Ext(cname))
	if err := tst.Init(defaultName); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}
	return &tst, nil
}

func ReadDir(dir string) (*Test, error) {
	filenames, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	var subTests []*Test
	for _, fn := range filenames {
		st, err := ReadFile(fn)
		if err != nil {
			return nil, err
		}
		subTests = append(subTests, st)
	}
	return &Test{
		Name:        filepath.Base(filepath.Clean(dir)),
		Description: fmt.Sprintf("test files from %s", dir),
		SubTests:    subTests,
	}, nil
}

func reflectValue(x any) reflect.Value {
	if x == nil {
		// The reflect value of nil is the zero reflect.Value, which can't be
		// passed to reflect.Call. But this works.
		return reflect.ValueOf(&x).Elem()
	}
	return reflect.ValueOf(x)
}
