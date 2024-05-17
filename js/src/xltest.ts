// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.

import fs from 'fs';
import path from 'path';
import { test, TestContext } from 'node:test';
import assert from 'node:assert';

export type Call = any[];

// TODO(jba): type is string to function
type funcMap = { [index: string]: any };

export class Test {
  name: string;
  description: string;
  functions: { [index: string]: string };
  env: { [index: string]: string };
  call: Call;
  want: any;
  subtests: Test[];

  static from(obj: any) {
    return Object.assign(new Test(), obj);
  }

  init(name: string) {
    if (!this.name) {
      this.name = name;
    }
    // TODO(jba): various checks, as in the Go equivalent.
    for (const i in this.subtests) {
      let st = Test.from(this.subtests[i]);
      this.subtests[i] = st;
      st.init(`${i}`);
    }
  }

  async run(t: TestContext, funcs: funcMap) {
    if (this.call) {
      const got = invoke(this.call, funcs);
      assert.equal(got, this.want);
    }
    if (this.subtests) {
      await Promise.all(this.subtests.map((st) =>st.run(t, funcs)));
    }
  }
}

function invoke(c: Call, funcs: funcMap): any {
  if (!c) {
    throw new Error('empty Call');
  }
  const f = funcs[c[0]];
  if (!f) {
    throw new Error(`missing function named "${c[0]}"`);
  }
  return f.apply(null, c.slice(1));
}

export function readFile(filePath: string): Test {
  const data = fs.readFileSync(filePath, 'utf8');
  let tst = Test.from(JSON.parse(data));
  const nname = path.normalize(filePath);
  const defaultName = path.basename(nname, path.extname(nname));
  tst.init(defaultName);
  return tst;
}

export function readDir(dir: string): Test {
  const files = fs.readdirSync(dir);
  let t = new Test();
  t.name = path.basename(path.normalize(dir));
  t.description = `files from ${dir}`;
  t.subtests = files
    .filter((f) => f.endsWith('.json'))
    .map((f) => readFile(path.join(dir, f)));
  return t;
}
