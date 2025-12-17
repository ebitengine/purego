#!/bin/sh -e

# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 The Ebitengine Authors

# BSD tests run within QEMU on Ubuntu.
# vmactions/*-vm only supports a single "step" where it
# brings down the VM at the end of the step, so all
# the commands to run need to be put into this single block.

echo "Running tests on $(uname -a) at $PWD"

PATH=$PATH:/usr/local/go/bin/

# verify Go is available
go version

echo "=> go build"
go build -v ./...
# Compile without optimization to check potential stack overflow.
# The option '-gcflags=all=-N -l' is often used at Visual Studio Code.
# See also https://go.googlesource.com/vscode-go/+/HEAD/docs/debugging.md#launch and the issue hajimehoshi/ebiten#2120.
go build "-gcflags=all=-N -l" -v ./...

# Check cross-compiling Windows binaries.
env GOOS=windows GOARCH=386 go build -v ./...
env GOOS=windows GOARCH=amd64 go build -v ./...
env GOOS=windows GOARCH=arm64 go build -v ./...

# Check cross-compiling macOS binaries.
env GOOS=darwin GOARCH=amd64 go build -v ./...
env GOOS=darwin GOARCH=arm64 go build -v ./...

# Check cross-compiling Linux binaries.
env GOOS=linux GOARCH=amd64 go build -v ./...
env GOOS=linux GOARCH=arm64 go build -v ./...

# Check cross-compiling FreeBSD binaries.
env GOOS=freebsd GOARCH=amd64 go build -gcflags="github.com/ebitengine/purego/internal/fakecgo=-std" -v ./...
env GOOS=freebsd GOARCH=arm64 go build -gcflags="github.com/ebitengine/purego/internal/fakecgo=-std" -v ./...

# Check cross-compiling NetBSD binaries.
env GOOS=netbsd GOARCH=amd64 go build -v ./...
env GOOS=netbsd GOARCH=arm64 go build -v ./...

echo "=> go build (plugin)"
# Make sure that plugin buildmode works since we save the R15 register (#254)
go build -buildmode=plugin ./examples/libc

echo "=> go mod vendor"
mkdir /tmp/vendoring
cd /tmp/vendoring
go mod init foo
echo 'package main' > main.go
echo 'import (' >> main.go
echo '  _ "github.com/ebitengine/purego"' >> main.go
echo ')' >> main.go
echo 'func main() {}' >> main.go
go mod edit -replace github.com/ebitengine/purego=$GITHUB_WORKSPACE
go mod tidy
go mod vendor
go build -v .

cd $GITHUB_WORKSPACE
echo "=> go test CGO_ENABLED=0"
env CGO_ENABLED=0 go test -gcflags="github.com/ebitengine/purego/internal/fakecgo=-std" -shuffle=on -v -count=10 ./...

echo "=> go test CGO_ENABLED=1"
env CGO_ENABLED=1 go test -shuffle=on -v -count=10 ./...

echo "=> go test CGO_ENABLED=0 w/o optimization"
env CGO_ENABLED=0 go test "-gcflags=all=-N -l -std" -v ./...

echo "=> go test CGO_ENABLED=1 w/o optimization"
env CGO_ENABLED=1 go test "-gcflags=all=-N -l" -v ./...

if [ -z "$(go version | grep '^1.1')" ]; then
  echo "=> go test race"
  go test -race -shuffle=on -v -count=10 ./...
fi
