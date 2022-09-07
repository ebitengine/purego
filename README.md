# purego
[![Go Reference](https://pkg.go.dev/badge/github.com/ebitengine/purego?GOOS=darwin.svg)](https://pkg.go.dev/github.com/ebitengine/purego?GOOS=darwin)

A library for calling C functions from Go without Cgo.

### External Code

Purego uses code that originates from the Go runtime. These files are under the BSD-3
License that can be found [in the Go Source](https://github.com/golang/go/blob/master/LICENSE).
This is a list of the copied files:

* `zcallback_darwin_*.s` from package `runtime`
* `internal/abi/abi_*.h` from package `runtime/cgo`
* `internal/fakecgo/callbacks.go` from package `runtime/cgo`
* `internal/fakecgo/iscgo.go` from package `runtime/cgo`
* `internal/fakecgo/setenv.go` from package `runtime/cgo`