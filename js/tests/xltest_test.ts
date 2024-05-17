// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

import { test } from 'node:test';
import assert from 'node:assert';
import { readFile, readDir } from '../src/xltest.js';

test('read JSON', () => {
  let tst = readFile('../testdata/add.json');
});

test('run', async (t) => {
  let tst = readDir('../testdata');
  await tst.run(t, { add: (x, y) => x + y });
});
