# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

from __future__ import annotations

import os.path as path
import unittest
import yaml

class Test:
    name: string = ''
    description: string = ''
    input: Any = None
    want: Any = None
    subtests: list[Test] = []

    def __init__(self, **entries): 
        self.__dict__.update(entries)
        self.input = entries.get('in')

    def init(self, name: string):
        if not self.name:
            self.name = name
        for i, stdict in enumerate(self.subtests):
            st = Test(**stdict)
            self.subtests[i] = st 
            st.init(f"{i}")

    def run(self, tc: unittest.TestCase, testFunc: Any, validateFunc: Any = None):
        if self.input is not None:
            got = testFunc(self.input)
            tc.assertEqual(got, self.want)
        for st in self.subtests:
            with tc.subTest(st.name):
                st.run(tc, testFunc, validateFunc)
            
        
def read_file(filename: string) -> Test:
    with open(filename, 'r') as file:
        d = yaml.safe_load(file)
        t = Test(**d)
        t.init(path.basename(path.normpath(filename)).removesuffix('.yaml'))
        return t
    
    

