# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

import unittest
from xltest import read_file

class TestXLTest(unittest.TestCase):

    def test_sample(self):
        self.assertEqual(1, 1)

    def test_add(self):
        #TODO(jba): OS-independent file paths
        tst = read_file('../../testdata/add.yaml')
        

        
        
