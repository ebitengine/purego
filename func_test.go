// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// This is an internal OS-dependent function for getting the handle to a library
//
//go:linkname openLibrary openLibrary
func openLibrary(name string) (uintptr, error)

func getSystemLibrary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib", nil
	case "linux":
		return "libc.so.6", nil
	case "freebsd":
		return "libc.so.7", nil
	case "windows":
		return "ucrtbase.dll", nil
	default:
		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}
}

func TestRegisterFunc(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}

func ExampleNewCallback() {
	cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int {
		fmt.Println(a1, a2, a3, a4, a5, a6, a7, a8, a9)
		return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
	})

	var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int
	purego.RegisterFunc(&fn, cb)

	ret := fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
	fmt.Println(ret)

	// Output: 1 2 3 4 5 6 7 8 9
	// 45
}

func Test_qsort(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	sorted := []int{2, 25, 56, 88, 100}
	compare := func(a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(a, b *int) int)
	purego.RegisterLibFunc(&qsort, libc, "qsort")
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	for i := range data {
		if data[i] != sorted[i] {
			t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
		}
	}
}

// BenchmarkRegisterFuncQsort-16    	  558045	      2064 ns/op	     264 B/op	       6 allocs/op
func BenchmarkRegisterFuncQsort(b *testing.B) {
	library, err := getSystemLibrary()
	if err != nil {
		b.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		b.Errorf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	compare := func(a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(a, b *int) int)
	purego.RegisterLibFunc(&qsort, libc, "qsort")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	}
}

// BenchmarkRegisterFuncStrlen-16    	 2411634	       490.4 ns/op	     120 B/op	       6 allocs/op
func BenchmarkRegisterFuncStrlen(b *testing.B) {
	library, err := getSystemLibrary()
	if err != nil {
		b.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		b.Errorf("failed to dlopen: %s", err)
	}
	var strlen func(string) int
	purego.RegisterLibFunc(&strlen, libc, "strlen")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		strlen("abcdefghijklmnopqrstuvwxyz")
	}
}

// v2

func Test2RegisterFuncPuts(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc2(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo! 2")
	puts("Calling C from from Go without Cgo! 3")
	puts("Calling C from from Go without Cgo! 4")
}

func Test2_strlen(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}
	var strlen func(string) int
	purego.RegisterLibFunc2(&strlen, libc, "strlen")
	count := strlen("abcdefghijklmnopqrstuvwxyz")
	if count != 26 {
		t.Errorf("strlen(0): expected 26 but got %d", count)
	}
	count = strlen("abcdefghijklmnopqrstuvwxyz")
	if count != 26 {
		t.Errorf("strlen(1): expected 26 but got %d", count)
	}
}

func Test2_qsort(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		t.Errorf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	sorted := []int{2, 25, 56, 88, 100}
	compare := func(a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(a, b *int) int)
	purego.RegisterLibFunc2(&qsort, libc, "qsort")
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	for i := range data {
		if data[i] != sorted[i] {
			t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
		}
	}
}

// Benchmark2RegisterFuncQsort-16    	  558032	      2057 ns/op	     264 B/op	       6 allocs/op
func Benchmark2RegisterFuncQsort(b *testing.B) {
	library, err := getSystemLibrary()
	if err != nil {
		b.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		b.Errorf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	compare := func(a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(a, b *int) int)
	purego.RegisterLibFunc2(&qsort, libc, "qsort")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	}
}

// Benchmark2RegisterFuncStrlen-16    	 2502175	       461.1 ns/op	     190 B/op	       5 allocs/op
func Benchmark2RegisterFuncStrlen(b *testing.B) {
	library, err := getSystemLibrary()
	if err != nil {
		b.Errorf("couldn't get system library: %s", err)
	}
	libc, err := openLibrary(library)
	if err != nil {
		b.Errorf("failed to dlopen: %s", err)
	}
	var strlen func(string) int
	purego.RegisterLibFunc2(&strlen, libc, "strlen")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		strlen("abcdefghijklmnopqrstuvwxyz")
	}
}
