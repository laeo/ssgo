.PHONY: test release clean

SRC=cmd/ssgo/main.go
BIN=bin/ssgo
VERSION=`git rev-parse HEAD | cut -c1-6`
BUILD=`date -u +%FT%T%z`

test:
	@go build --ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" -o ${BIN} ${SRC}

release:
	@GOOS=linux GOARCH=amd64 go build --ldflags "-X main.version=${VERSION} -X main.build=${BUILD}" -o ${BIN} ${SRC}

clean:
	@if [ -f ${BIN}]; then rm -f ${BIN}; fi