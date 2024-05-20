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
    subtests: list[Test] = []

    def __init__(self, **entries): 
        self.__dict__.update(entries)
        self.input = entries.get('in')
        self.subtests = [Test(**d) for d in self.subtests]

    def init(self, name: string, prepend: bool = False):
        if not self.name:
            self.name = name
        for i, st in enumerate(self.subtests):
            n = f"{i}"
            if prepend:
                n = self.name + "/" + n
            st.init(n, True)

    def run(self, tc: unittest.TestCase, testFunc: Any, validateFunc: Any = None):
        oldenv = {}
        for name, val in self.env.items():
            oldenv[name] = os.environ.get(name)
            os.environ[name] = val
        try:
            if self.input is not None:
                got = testFunc(self.input)
                if validateFunc:
                    validateFunc(got, self.want)
                else:
                    tc.assertEqual(got, self.want)
            for st in self.subtests:
                with tc.subTest(st.name):
                    st.run(tc, testFunc, validateFunc)
        finally:
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
    for filename in os.listdir(dir):
        if filename.endswith('.yaml'):
            test.subtests.push(read_file(os.path.join(dir, filename)))
    test.init(os.path.basename(os.path.normalize(dir)))
    return test
