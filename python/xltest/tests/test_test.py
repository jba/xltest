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
        t = read_file(os.path.join(self.testdata_dir, 'add.yaml'))
        t.run(self, lambda args: args[0] + args[1])

    def test_env(self):
        t = read_file(os.path.join(self.testdata_dir, 'env.yaml'))
        t.run(self, lookup)

    def test_validate(self):
        t = read_file(os.path.join(self.testdata_dir, 'validate.yaml'))
        t.run(self, lambda s: 'You say ' + s, self.assertRegex)

    def test_read_dir(self):
        t = read_dir(self.testdata_dir)
        self.assertEqual(t.name, "testdata")
        self.assertEqual([s.name for s in t.subtests], ['add', 'env', 'validate'])

def lookup(s: string) -> string:
    if s != '':
        return s
    return os.environ.get('XLTEST') or ''

