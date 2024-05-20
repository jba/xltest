// Copyright 2024 Jonathan Amsterdam. All rights reserved.
// Use of this source code is governed by a license that
// can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"testing"
)

func TestValidateTestdata(t *testing.T) {
	files, err := filepath.Glob(filepath.FromSlash("../../../../testdata/*.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateFiles(files); err != nil {
		t.Fatal(err)
	}
}
