# purego
[![Go Reference](https://pkg.go.dev/badge/github.com/ebitengine/purego?GOOS=darwin.svg)](https://pkg.go.dev/github.com/ebitengine/purego?GOOS=darwin)

A library for calling C functions from Go without Cgo.

## Example

```go
package main

import (
	"fmt"
	"runtime"

	"github.com/ebitengine/purego"
)

func getSystemLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib"
	case "linux":
		return "libc.so.6"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func main() {
	libc := purego.Dlopen(getSystemLibrary(), purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err := purego.Dlerror(); err != "" {
		panic(err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}
```

Then to run: `CGO_ENABLED=0 go run main.go`

### External Code

Purego uses code that originates from the Go runtime. These files are under the BSD-3
License that can be found [in the Go Source](https://github.com/golang/go/blob/master/LICENSE).
This is a list of the copied files:

* `zcallback_darwin_*.s` from package `runtime`
* `internal/abi/abi_*.h` from package `runtime/cgo`
* `internal/fakecgo/asm_GOARCH.s` from package `runtime/cgo`
* `internal/fakecgo/callbacks.go` from package `runtime/cgo`
* `internal/fakecgo/go_GOOS_GOARCH.go` from package `runtime/cgo`
* `internal/fakecgo/iscgo.go` from package `runtime/cgo`
* `internal/fakecgo/setenv.go` from package `runtime/cgo`
