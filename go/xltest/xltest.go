// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

package xltest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Test struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// Functions describes how the test's functions should behave.
	Functions map[string]string `json:"functions"`
	Env       map[string]string `json:"env,omitempty"`
	SetUp     []Call            `json:"setup,omitempty"`
	TearDown  []Call            `json:"teardown,omitempty"`
	// Can be empty if this just holds subtests
	Call Call `json:"call,omitempty"`
	Want any  `json:"want,omitempty"`
	// Name of evaluation function.
	// It must take (got, want) and return a string.
	Eval     string  `json:"eval,omitempty"`
	SubTests []*Test `json:"subtests,omitempty"`
}

// A Call represents a function call as a slice.
// Call[0] is a string that names the function.
// The remaining elements are the arguments.
type Call []any

func ReadFile(filename string) (*Test, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tst Test
	if err := json.NewDecoder(f).Decode(&tst); err != nil {
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
	filenames, err := filepath.Glob(filepath.Join(dir, "*.json"))
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

func (tst *Test) Init(name string) error {
	if tst.Name == "" {
		if name == "" {
			return errors.New("no name for top-level test")
		}
		tst.Name = name
	}
	var errs []error
	tst.init("", map[string]string{}, func(msg string) {
		errs = append(errs, errors.New(msg))
	})
	return errors.Join(errs...)
}

func (tst *Test) init(prefix string, functions map[string]string, addMsg func(string)) {
	prefix = path.Join(prefix, tst.Name)
	addf := func(format string, args ...any) {
		addMsg(prefix + ":" + fmt.Sprintf(format, args...))
	}

	for name, desc := range tst.Functions {
		functions[name] = desc
	}
	foundEmpty := false
	var calls []Call
	calls = append(calls, tst.SetUp...)
	calls = append(calls, tst.TearDown...)
	if tst.Call != nil {
		calls = append(calls, tst.Call)
	} else if tst.Want != nil {
		addf("call is empty but 'want' is not")
	}
	for _, call := range calls {
		if len(call) == 0 && !foundEmpty {
			addf("contains empty call")
			foundEmpty = true
			continue
		}
		funcName, ok := call[0].(string)
		if !ok {
			addf("call %v: first element must be a string", call)
			continue
		}
		if _, ok := functions[funcName]; !ok {
			addf("Test.Functions missing %q", funcName)
		}
	}
	for i, st := range tst.SubTests {
		if st.Name == "" {
			st.Name = fmt.Sprint(i)
		}
		st.init(prefix, functions, addMsg)
	}
}

type userFunc func([]any) (any, error)

func toUserFunc(name string, x any) (userFunc, error) {
	if name == "" {
		return nil, errors.New("empty function name")
	}
	fv := reflect.ValueOf(x)
	tv := fv.Type()
	if fv.Kind() != reflect.Func {
		return nil, fmt.Errorf("%s: not a function", name)
	}
	switch tv.NumOut() {
	case 0:
		return func(args []any) (any, error) {
			fv.Call(toReflectValues(args))
			return nil, nil
		}, nil
	case 1:
		return func(args []any) (any, error) {
			rs := fv.Call(toReflectValues(args))
			return rs[0].Interface(), nil
		}, nil
	case 2:
		if tv.Out(1) != reflect.TypeFor[error]() {
			return nil, fmt.Errorf("%s: second return value is not error", name)
		}
		return func(args []any) (any, error) {
			rs := fv.Call(toReflectValues(args))
			return rs[0].Interface(), rs[1].Interface().(error)
		}, nil
	default:
		return nil, fmt.Errorf("%s: more than two result values", name)
	}
}

func toReflectValues(xs []any) []reflect.Value {
	vs := make([]reflect.Value, len(xs))
	for i, x := range xs {
		v := reflect.ValueOf(x)
		if !v.IsValid() {
			// Can't pass a zero reflect.Value to Call.
			// TODO(jba): do something more principled.
			v = reflect.ValueOf((*int)(nil))
		}
		vs[i] = v
	}
	return vs
}

func (tst *Test) Run(t *testing.T, funcMap map[string]any) {
	funcs := map[string]userFunc{}
	for name, fn := range funcMap {
		uf, err := toUserFunc(name, fn)
		if err != nil {
			t.Fatal(err)
		}
		funcs[name] = uf
	}
	tst.run(t, funcs, func(got, want any) string {
		if cmp.Equal(got, want) {
			return ""
		}
		return fmt.Sprintf("got %v, want %v", got, want)
	})
}

func (tst *Test) run(t *testing.T, funcs map[string]userFunc, eval func(any, any) string) {
	if tst.Eval != "" {
		uf, ok := funcs[tst.Eval]
		if !ok {
			t.Fatalf("missing eval function %s", tst.Eval)
		}
		eval = func(a, b any) string {
			t.Helper()
			r, err := uf([]any{a, b})
			if err != nil {
				t.Fatal(err)
			}
			return r.(string)
		}
	}

	t.Run(tst.Name, func(t *testing.T) {
		for name, value := range tst.Env {
			t.Setenv(name, value)
		}
		if _, err := invokeCalls(tst.SetUp, funcs); err != nil {
			t.Fatalf("during setup: %v", err)
		}
		t.Cleanup(func() {
			if _, err := invokeCalls(tst.TearDown, funcs); err != nil {
				t.Fatalf("during teardown: %v", err)
			}
		})

		if tst.Call != nil {
			got, err := invoke(tst.Call, funcs)
			if err != nil {
				t.Fatalf("during test calls: %v", err)
			}
			if s := eval(got, tst.Want); s != "" {
				t.Error(s)
			}
		}
		for _, test := range tst.SubTests {
			test.run(t, funcs, eval)
		}
	})
}

func invokeCalls(cs []Call, funcs map[string]userFunc) (result any, err error) {
	for _, c := range cs {
		result, err = invoke(c, funcs)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func invoke(c Call, funcs map[string]userFunc) (any, error) {
	if len(c) == 0 {
		panic("empty Call")
	}
	name := c[0].(string)
	if name == "" {
		panic("empty function name in Call")
	}
	f := funcs[name]
	if f == nil {
		panic(fmt.Sprintf("missing function named %s", name))
	}
	return f(c[1:])
}

// func ReadJSON(r io.Reader) (*Test, error) {
// 	var t
// 	for _, fname := range s.functionNames() {
// 		if _, ok := s.Functions[fname]; !ok {
// 			return nil, fmt.Errorf("suite.functions missing %q", fname)
// 		}
// 	}
// 	return &s, nil
// }

// func (s *Suite) functionNames() []string {
// 	var ns []string
// 	if s.SetUp != nil {
// 		ns = append(ns, s.SetUp.Func)
// 	}
// 	if s.TearDown != nil {
// 		ns = append(ns, s.TearDown.Func)
// 	}
// 	if s.Compare != "" {
// 		ns = append(ns, s.Compare)
// 	}
// 	for _, t := range s.Tests {
// 		ns = append(ns, t.Call.Func)
// 	}
// 	return ns
// }
