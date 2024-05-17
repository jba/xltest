'use strict';
// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that can be found in the LICENSE
// file.
Object.defineProperty(exports, '__esModule', { value: true });
exports.readDir = exports.readFile = exports.Test = void 0;
var fs_1 = require('fs');
var path_1 = require('path');
var Test = /** @class */ (function () {
  function Test() {}
  Test.fromJSON = function (json) {
    return Object.assign(new Test(), json);
  };
  Test.prototype.init = function (name) {
    if (!this.name) {
      this.name = name;
    }
    // TODO(jba): various checks, as in the Go equivalent.
    for (var i in this.subtests) {
      var st = Test.fromJSON(this.subtests[i]);
      this.subtests[i] = st;
      st.init(''.concat(i));
    }
  };
  return Test;
})();
exports.Test = Test;
function readFile(filePath) {
  var data = fs_1.default.readFileSync(filePath, 'utf8');
  var tst = Test.fromJSON(data);
  var nname = path_1.default.normalize(filePath);
  var defaultName = path_1.default.basename(
    nname,
    path_1.default.extname(nname)
  );
  tst.init(defaultName);
  return tst;
}
exports.readFile = readFile;
function readDir(dir) {
  var files = fs_1.default.readdirSync(dir);
  var t = new Test();
  t.name = path_1.default.basename(path_1.default.normalize(dir));
  t.description = 'files from '.concat(dir);
  t.subtests = files
    .filter(function (f) {
      return f.endsWith('.json');
    })
    .map(readFile);
  return t;
}
exports.readDir = readDir;
