package objc

import (
	"strings"
	"unsafe"
)

func cstring(g string) *uint8 {
	if !strings.HasSuffix(g, "\x00") {
		panic("str argument missing null terminator: " + g)
	}
	return &(*(*[]byte)(unsafe.Pointer(&g)))[0]
}
