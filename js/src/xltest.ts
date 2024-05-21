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

enum OnError {
  fail = 'fail',
  succeed = 'succeed',
  validate = 'validate',
}

export class Test {
  name: string = '';
  description: string = '';
  env: Env = {};
  in: any;
  want: any;
  onError: OnError | undefined;
  subtests: Test[] = [];

  static from(obj: any) {
    return Object.assign(new Test(), obj);
  }

  init(name: string) {
    if (!this.name) {
      this.name = name;
    }
    if (this.in === undefined && this.want !== undefined)
      throw new Error(`test ${this.name} has 'want' but not 'in'`);
    if (this.in === undefined && this.subtests.length == 0)
      throw new Error(`test ${this.name} has no 'in' and no subtests`);
    for (const i in this.subtests) {
      let st = Test.from(this.subtests[i]);
      this.subtests[i] = st;
      st.init(`${i}`);
    }
  }

  async run(t: TestContext, testFunc: any, validateFunc: any) {
    if (!validateFunc) validateFunc = assert.deepStrictEqual;
    return this._run(t, testFunc, validateFunc, OnError.fail);
  }

  async _run(
    t: TestContext,
    testFunc: any,
    validateFunc: any,
    onError: OnError
  ) {
    if (this.onError !== undefined) onError = this.onError;
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
          let got: any;
          switch (onError) {
            case OnError.fail:
              assert.doesNotThrow(() => {
                got = testFunc(this.in);
              });
              validateFunc(got, this.want);
              break;

            case OnError.succeed:
              assert.throws(() => testFunc(this.in));
              break;

            case OnError.validate:
              try {
                got = testFunc(this.in);
              } catch (e) {
                got = e;
              }
              validateFunc(got, this.want);
              break;
          }
        }
        for (const st of this.subtests) {
          await st._run(t, testFunc, validateFunc, onError);
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
  t.description = `files from ${dir}`;
  t.subtests = files
    .filter((f) => f.endsWith('.yaml'))
    .map((f) => readFile(path.join(dir, f)));
  t.init(path.basename(path.normalize(dir)));
  return t;
}
