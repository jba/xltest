# Copyright 2024 Jonathan Amsterdam. All rights reserved.
# Use of this source code is governed by a license that
# can be found in the LICENSE file.

test:
	cd go/xltest && go test
	cd js && npm test

format:
	cd go/xltest && gofmt -w *.go
	cd js && npm run format
	
	

