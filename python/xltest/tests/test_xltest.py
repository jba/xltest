# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

from __future__ import annotations

import os
import unittest
from xltest import read_file, read_dir

class TestXLTest(unittest.TestCase):

    testdata_dir = os.path.join('..', '..', 'testdata')
    
    def test_add(self):
        self.run_file('add', lambda args: args[0] + args[1])

    def test_env(self):
        def lookup(s: string) -> string:
            if s != '':
                return s
            return os.environ.get('XLTEST') or ''

        self.run_file('env', lookup)

    def test_validate(self):
        self.run_file('validate', lambda s: 'You say ' + s, self.assertRegex)

    def test_errors(self):
        def validate(got, want):
            if issubclass(type(got), Exception):
                self.assertIsInstance(got, ValueError)
            else:
                self.assertEqual(got, want)
                
        self.run_file('errors', int, validate)

    def run_file(self, name: string, testFunc: Any, validateFunc: Any = None):
        t = read_file(os.path.join(self.testdata_dir, name + '.yaml'))
        t.run(self, testFunc, validateFunc)

    def test_read_dir(self):
        t = read_dir(self.testdata_dir)
        self.assertEqual(t.name, "testdata")
        self.assertEqual([s.name for s in t.subtests], ['add', 'env', 'errors', 'validate'])


