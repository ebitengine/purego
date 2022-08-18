package fakecgo

import _ "unsafe" // for go:linkname

//go:linkname _iscgo runtime.iscgo
var _iscgo bool = true
