package fakecgo

import _ "unsafe"

// let's pretend we have cgo:
//go:linkname _iscgo runtime.iscgo
var _iscgo = true

//go:linkname x_cgo_thread_start _cgo_thread_start

//go:nosplit
func x_cgo_thread_start() {

}
