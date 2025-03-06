// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

func getSystemLibrary() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/lib/libSystem.B.dylib", nil
	case "freebsd":
		return "libc.so.7", nil
	case "linux":
		return "libc.so.6", nil
	case "netbsd":
		return "libc.so", nil
	case "windows":
		return "ucrtbase.dll", nil
	default:
		return "", fmt.Errorf("GOOS=%s is not supported", runtime.GOOS)
	}
}

func TestRegisterFunc(t *testing.T) {
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}
	var puts func(string)
	purego.RegisterLibFunc(&puts, libc, "puts")
	puts("Calling C from from Go without Cgo!")
}

func Test_qsort(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("Platform doesn't support Floats")
		return
	}
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}

	data := []int{88, 56, 100, 2, 25}
	sorted := []int{2, 25, 56, 88, 100}
	compare := func(_ purego.CDecl, a, b *int) int {
		return *a - *b
	}
	var qsort func(data []int, nitms uintptr, size uintptr, compar func(_ purego.CDecl, a, b *int) int)
	purego.RegisterLibFunc(&qsort, libc, "qsort")
	qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
	for i := range data {
		if data[i] != sorted[i] {
			t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
		}
	}
}

func TestRegisterFunc_Floats(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("Platform doesn't support Floats")
		return
	}
	library, err := getSystemLibrary()
	if err != nil {
		t.Fatalf("couldn't get system library: %s", err)
	}
	libc, err := load.OpenLibrary(library)
	if err != nil {
		t.Fatalf("failed to dlopen: %s", err)
	}
	{
		var strtof func(arg string) float32
		purego.RegisterLibFunc(&strtof, libc, "strtof")
		const (
			arg = "2"
		)
		got := strtof(arg)
		expected := float32(2)
		if got != expected {
			t.Errorf("strtof failed. got %f but wanted %f", got, expected)
		}
	}
	{
		var strtod func(arg string, ptr **byte) float64
		purego.RegisterLibFunc(&strtod, libc, "strtod")
		const (
			arg = "1"
		)
		got := strtod(arg, nil)
		expected := float64(1)
		if got != expected {
			t.Errorf("strtod failed. got %f but wanted %f", got, expected)
		}
	}
}

func TestRegisterLibFunc_Bool(t *testing.T) {
	if runtime.GOARCH != "arm64" && runtime.GOARCH != "amd64" {
		t.Skip("Platform doesn't support callbacks")
		return
	}
	// this callback recreates the state where the return register
	// contains other information but the least significant byte is false
	cbFalse := purego.NewCallback(func() uintptr {
		x := uint64(0x7F5948AE9A00)
		return uintptr(x)
	})
	var runFalse func() bool
	purego.RegisterFunc(&runFalse, cbFalse)
	expected := false
	if got := runFalse(); got != expected {
		t.Errorf("runFalse failed. got %t but wanted %t", got, expected)
	}
}
