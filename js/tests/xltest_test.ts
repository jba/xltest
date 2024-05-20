// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

import { test } from 'node:test';
import assert from 'node:assert';
import path from 'path';
import process from 'process';
import { readFile, readDir } from '../src/xltest.js';

const testdataDir = path.join('..', 'testdata');

test('Test.run', async (t) => {
  const tests = {
    add: { testFunc: (args) => args[0] + args[1] },
    env: {
      testFunc: (s) => {
        if (s) return s;
        const e = process.env.XLTEST;
        if (e === undefined) return '';
        return e;
      },
    },
    validate: {
      testFunc: (s) => 'You said ' + s,
      validate: (got, re) => assert.match(got, new RegExp(re)),
    },
  };
  for (const name in tests) {
    const tst = readFile(path.join(testdataDir, name + '.yaml'));
    await tst.run(t, tests[name].testFunc, tests[name].validate);
  }
});

test('readDir', async (t) => {
  const tst = readDir(testdataDir);
  assert.equal(tst.name, "testdata");
  assert.deepStrictEqual(tst.subtests.map((s) => s.name), ['add', 'env', 'validate']);
})
