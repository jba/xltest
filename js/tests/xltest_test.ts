// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

import { test } from 'node:test';
import path from 'path';
import * as process from 'process';
import { readFile } from '../src/xltest.js';

test('Test.run', async (t) => {
  const dir = path.join('..', 'testdata');
  const tests = {
    add: {testFunc: (args) => args[0] + args[1]},
    env: {testFunc: (s) => {
            if (s) return s;
            const e = process.env.XLTEST;
            if (e === undefined) return '';
            return e;
          }},
  };
  for (const name in tests) {
    let tst = readFile(path.join(dir, name + '.yaml'));
    await tst.run(t, tests[name].testFunc, tests[name].validate);
  }
});
