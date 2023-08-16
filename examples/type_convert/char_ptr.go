package main

import (
	"reflect"
	"unsafe"
)

/*
process.h

ABCXXX char* Process(int* output_len);
*/

// Process takes an int ptr to set output length
// and returns a string pointer which declared as char* in C.
// Sign output as *string is not worked if char isn't
// terminated with '\0' (see https://github.com/ebitengine/purego/issues/150).
var Process func(outputLen *int) (output unsafe.Pointer)

func main() {
	outputLen := 0
	output := Process(&outputLen)

	// manually construct a string struct and set data ptr and length
	s := *(*string)(unsafe.Pointer(&reflect.StringHeader{Data: uintptr(output), Len: outputLen}))
	println(s)

	// to byte slice
	bs := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(output), Len: outputLen, Cap: outputLen}))
	println(bs)
}
