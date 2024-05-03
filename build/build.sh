#!/bin/bash

echo "linux_amd64: building"
GOOS=linux \
GOARCH=amd64 \
CGO_ENABLED=1 \
GODEBUG=cgocheck=0 \
  go build -buildvcs=false -pgo=default.pgo -o=dendy_linux_amd64 ./cmd/dendy

echo "win_amd64: building"
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
CGO_ENABLED=1 \
GODEBUG=cgocheck=0 \
CGO_LDFLAGS="-static-libgcc -static -lpthread" \
  go build -buildvcs=false -pgo=default.pgo -o=dendy_win_amd64.exe ./cmd/dendy

echo "done"
