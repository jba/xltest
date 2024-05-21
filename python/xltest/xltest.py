# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

from __future__ import annotations

import os
import unittest
import yaml

class Test:
    name: string = ''
    description: string = ''
    env: dict[string, string] = {}
    input: Any = None
    want: Any = None
    # TODO: an enum type for onError
    onError: string = ''
    subtests: list[Test] = []

    def __init__(self, **entries): 
        self.__dict__.update(entries)
        self.input = entries.get('in')
        self.subtests = [Test(**d) for d in self.subtests]

    def init(self, name: string, prepend: bool = False):
        if not self.name:
            self.name = name
        if self.input is None and self.want is not None:
            raise ValueError(f"test {self.name} has 'want' but not 'in'")
        if self.input is None and len(self.subtests) == 0:
            raise ValueError(f"test {self.name} has no input and no subtests")
        for i, st in enumerate(self.subtests):
            n = f"{i}"
            if prepend:
                n = self.name + "/" + n
            st.init(n, True)

    def run(self, tc: unittest.TestCase, testFunc: Any, validateFunc: Any = None):
        if not validateFunc:
            validateFunc = tc.assertEqual
        self._run(tc, testFunc, validateFunc, 'fail')

    def _run(self, tc: unittest.TestCase, testFunc: Any, validateFunc: Any, onError: string):
        # Override onError if set in the test.
        if self.onError != '':
            onError = self.onError
        # Set environment variables, remembering their values.
        oldenv = {}
        for name, val in self.env.items():
            oldenv[name] = os.environ.get(name)
            os.environ[name] = val
        try:
            if self.input is not None:
                match onError:
                    case 'fail':
                        got = testFunc(self.input)
                        validateFunc(got, self.want)

                    case 'succeed':
                        tc.assertRaises(Exception, testFunc, self.input)

                    case 'validate':
                        try:
                            got = testFunc(self.input)
                            validateFunc(got, self.want)
                        except Exception as e:
                            validateFunc(e, self.want)

                    case _:
                        tc.fail(f"unknown onError value: {onError}")

            for st in self.subtests:
                with tc.subTest(st.name):
                    st._run(tc, testFunc, validateFunc, onError)
        finally:
            # Restore environment variables.
            for name, val in oldenv.items():
                if val is None:
                    del os.environ[name]
                else:
                    os.environ[name] = val
            
        
def read_file(filename: string) -> Test:
    with open(filename, 'r') as file:
        d = yaml.safe_load(file)
        t = Test(**d)
        t.init(os.path.basename(os.path.normpath(filename)).removesuffix('.yaml'))
        return t
    
def read_dir(dir: string) -> Test:    
    test = Test()
    for filename in sorted(os.listdir(dir)):
        if filename.endswith('.yaml'):
            test.subtests.append(read_file(os.path.join(dir, filename)))

    test.init(os.path.basename(os.path.normpath(dir)))
    return test
