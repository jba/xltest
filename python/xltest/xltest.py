# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

from __future__ import annotations

import unittest
import yaml

class Test:
    name: string = ''
    description: string = ''
    input: Any = None
    want: Any = None

    def __init__(self, **entries): 
        self.__dict__.update(entries)

    def run(self, tc: unittest.TestCase, testFunc: Any, validateFunc: Any):
        print('run', self.name)
        if self.input is not None:
            got = testFunc(self.input)
            self.assertEqual(got, self.want)
        print('todo: subtests')
        
def read_file(filename: string) -> Test:
    with open(filename, 'r') as file:
        d = yaml.safe_load(file)
        return Test(**d)
        
    

