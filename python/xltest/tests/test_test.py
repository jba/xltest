# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

import os
import unittest
from xltest import read_file

class TestXLTest(unittest.TestCase):

    dir = os.path.join('..', '..', 'testdata')
    
    def test_add(self):
        #TODO(jba): OS-independent file paths
        tst = read_file(os.path.join(self.dir, 'add.yaml'))
        tst.run(self, lambda args: args[0] + args[1])

        
