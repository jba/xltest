# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

from __future__ import annotations

import os
import unittest
from xltest import read_file

class TestXLTest(unittest.TestCase):

    dir = os.path.join('..', '..', 'testdata')
    
    def test_add(self):
        t = read_file(os.path.join(self.dir, 'add.yaml'))
        t.run(self, lambda args: args[0] + args[1])

    def test_env(self):
        t = read_file(os.path.join(self.dir, 'env.yaml'))
        t.run(self, lookup)

    def test_validate(self):
        t = read_file(os.path.join(self.dir, 'validate.yaml'))
        t.run(self, lambda s: 'You say ' + s, self.assertRegex)

def lookup(s: string) -> string:
    if s != '':
        return s
    return os.environ.get('XLTEST') or ''

