// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin || linux

package purego_test

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

func getSystemLibrary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib", nil
	case "linux":
		return "libc.so.6", nil
	default:
		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}
}

func TestRegisterFunc(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err := purego.Dlerror(); err != "" {
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

	//Output: 1 2 3 4 5 6 7 8 9
	// 45
}

func Test_qsort(t *testing.T) {
	if runtime.GOOS == "linux" {
		// TODO: enable once callbacks are working properly on Linux
		t.SkipNow()
	}

	library, err := getSystemLibrary()
	if err != nil {
		t.Errorf("couldn't get system library: %s", err)
	}
	libc := purego.Dlopen(library, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err := purego.Dlerror(); err != "" {
		t.Errorf("failed to dlopen: %s", err)
	}

	var data = []int{88, 56, 100, 2, 25}
	var sorted = []int{2, 25, 56, 88, 100}
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
