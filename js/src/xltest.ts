// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that
// can be found in the LICENSE file.

import fs from 'fs';
import path from 'path';
import * as process from 'process';
import { parse } from 'yaml';
import { test, TestContext } from 'node:test';
import assert from 'node:assert';

type Env = { [index: string]: string | undefined };

export class Test {
  name: string = '';
  description: string = '';
  env: Env = {};
  in: any;
  want: any;
  subtests: Test[] = [];

  static from(obj: any) {
    return Object.assign(new Test(), obj);
  }

  init(name: string) {
    if (!this.name) {
      this.name = name;
    }
    for (const i in this.subtests) {
      let st = Test.from(this.subtests[i]);
      this.subtests[i] = st;
      st.init(`${i}`);
    }
  }

  async run(t: TestContext, testFunc: any, validateFunc: any) {
    await t.test(this.name, async (t) => {
      let oldenv = {};
      if (this.env) {
        // Set environment variables, remembering their previous values.
        for (const name in this.env) {
          oldenv[name] = process.env[name];
        }
        setEnv(process.env, this.env);
      }
      try {
        if (this.in !== undefined) {
          const got = testFunc(this.in);
          if (validateFunc) {
            validateFunc(got, this.want);
          } else {
            assert.deepStrictEqual(got, this.want);
          }
        }
        for (const st of this.subtests) {
          await st.run(t, testFunc, validateFunc);
        }
      } finally {
        // Restore environment variables to their previous values.
        setEnv(process.env, oldenv);
      }
    });
  }
}

function setEnv(dest: Env, src: Env) {
  for (const name in src) {
    const val = src[name];
    if (val === undefined) {
      delete dest[name];
    } else {
      dest[name] = val;
    }
  }
}

export function readFile(filePath: string): Test {
  const data = fs.readFileSync(filePath, 'utf8');
  let tst = Test.from(parse(data));
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
    .filter((f) => f.endsWith('.yaml'))
    .map((f) => readFile(path.join(dir, f)));
  return t;
}
