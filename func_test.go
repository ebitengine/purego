// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/strings"
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

// NewCallBack

func Test_NewCallBack(t *testing.T) {
	// Original
	t.Run("RegisterFunc(original)", func(t *testing.T) {
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
	})
	// New
	t.Run("RegisterFunc9_1", func(t *testing.T) {
		cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int {
			fmt.Println(a1, a2, a3, a4, a5, a6, a7, a8, a9)
			return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
		})

		var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int
		purego.RegisterFunc9_1(&fn, cb)

		ret := fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
		fmt.Println(ret)

		// Output: 1 2 3 4 5 6 7 8 9
		// 45
	})
}

func Benchmark_NewCallBack(b *testing.B) {
	// Original
	b.Run("RegisterFunc(original)", func(b *testing.B) {
		// 1000000, 1111 ns/op, 328 B/op, 12 allocs/op
		cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int {
			return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
		})

		var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int
		purego.RegisterFunc(&fn, cb)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
		}
	})
	// New
	b.Run("RegisterFunc9_1(new)", func(b *testing.B) {
		// 3153188, 383.6 ns/op, 144 B/op, 1 allocs/op
		cb := purego.NewCallback(func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int {
			return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9
		})

		var fn func(a1, a2, a3, a4, a5, a6, a7, a8, a9 int) int
		purego.RegisterFunc9_1(&fn, cb)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = fn(1, 2, 3, 4, 5, 6, 7, 8, 9)
		}
	})
}

// qsort

func Test_qsort(t *testing.T) {
	// Original
	t.Run("RegisterFunc(original)", func(t *testing.T) {
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
	})
	// New
	t.Run("RegisterFunc4_0(new)", func(t *testing.T) {
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
		symbol := purego.Symbol(libc, "qsort")
		purego.RegisterFunc4_0(&qsort, symbol)
		qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
		for i := range data {
			if data[i] != sorted[i] {
				t.Errorf("got %d wanted %d at %d", data[i], sorted[i], i)
			}
		}
	})
}

func Benchmark_qsort(b *testing.B) {
	// Original
	b.Run("RegisterFunc(original)", func(b *testing.B) {
		// 558027, 2067 ns/op, 264 B/op, 6 allocs/op
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
	})
	// New
	b.Run("RegisterFunc4_0(new)", func(b *testing.B) {
		// 648578, 1806 ns/op, 296 B/op, 4 allocs/op
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
		symbol := purego.Symbol(libc, "qsort")
		purego.RegisterFunc4_0(&qsort, symbol)
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			qsort(data, uintptr(len(data)), unsafe.Sizeof(int(0)), compare)
		}
	})
}

// puts

func Test_puts(t *testing.T) {
	// Original
	t.Run("RegisterFunc(original)", func(t *testing.T) {
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
		puts("Calling C from from Go without Cgo! (original)")
	})
	// New
	t.Run("RegisterFunc1_0(new)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var puts func(string)
		symbol := purego.Symbol(libc, "puts")
		purego.RegisterFunc1_0(&puts, symbol)
		puts("Calling C from from Go without Cgo! (new)")
	})
}

// strlen

func Test_strlen(t *testing.T) {
	t.Run("RegisterFunc1_1(new)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var strlen func(string) int
		symbol := purego.Symbol(libc, "strlen")
		purego.RegisterFunc1_1(&strlen, symbol)
		count := strlen("abcdefghijklmnopqrstuvwxyz")
		if count != 26 {
			t.Errorf("strlen(0): expected 26 but got %d", count)
		}
		count = strlen("abcdefghijklmnopqrstuvwxyz")
		if count != 26 {
			t.Errorf("strlen(1): expected 26 but got %d", count)
		}
	})
}

func Benchmark_strlen(b *testing.B) {
	// Current
	b.Run("RegisterFunc(original)", func(b *testing.B) {
		// 2411634 - 490.4 ns/op - 120 B/op - 6 allocs/op
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
	})
	// New
	b.Run("RegisterFunc1_1(new)", func(b *testing.B) {
		// 7690965 - 157.0 ns/op - 176 B/op - 2 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		var strlen func(string) int
		symbol := purego.Symbol(libc, "strlen")
		purego.RegisterFunc1_1(&strlen, symbol)
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			strlen("abcdefghijklmnopqrstuvwxyz")
		}
	})
	// Direct
	b.Run("SyscallN", func(b *testing.B) {
		// 8449221, 142.2 ns/op, 112 B/op, 2 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		symbol := purego.Symbol(libc, "strlen")
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			ptr := strings.CString("abcdefghijklmnopqrstuvwxyz")
			sysargs := [9]uintptr{
				uintptr(unsafe.Pointer(ptr)),
			}
			_, _, _ = purego.SyscallN(symbol, sysargs[:]...)
		}
	})
}

// cos

func Test_cos(t *testing.T) {
	// Original
	t.Run("RegisterFunc(original)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var cos func(float64) float64
		purego.RegisterLibFunc(&cos, libc, "cos")
		// 0.05428962282295477
		const v = 1.51648
		expected := math.Cos(v)
		actual := cos(v)
		if expected != actual {
			t.Errorf("cos(%.8f): expected %.8f but got %.8f", v, expected, actual)
		}
	})
	// New
	t.Run("RegisterFunc1_1(new)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var cos func(float64) float64
		symbol := purego.Symbol(libc, "cos")
		purego.RegisterFunc1_1(&cos, symbol)
		// 0.05428962282295477
		const v = 1.51648
		expected := math.Cos(v)
		actual := cos(v)
		if expected != actual {
			t.Errorf("cos(%.8f): expected %.8f but got %.8f", v, expected, actual)
		}
	})
}

func Benchmark_cos(b *testing.B) {
	// Original
	b.Run("RegisterFunc(original)", func(b *testing.B) {
		// 3337392, 362.0 ns/op, 64 B/op, 4 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		var cos func(float64) float64
		purego.RegisterLibFunc(&cos, libc, "cos")
		// 0.05428962282295477
		const v = 1.51648
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cos(v)
		}
	})
	// New
	b.Run("RegisterFunc1_1(new)", func(b *testing.B) {
		// 9300645, 129.0 ns/op, 144 B/op, 1 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		var cos func(float64) float64
		symbol := purego.Symbol(libc, "cos")
		purego.RegisterFunc1_1(&cos, symbol)
		// 0.05428962282295477
		const v = 1.51648
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cos(v)
		}
	})
	// Go
	b.Run("Go", func(b *testing.B) {
		const v = 1.51648
		for i := 0; i < b.N; i++ {
			_ = math.Cos(v)
		}
	})
}

// isupper

func Test_isupper(t *testing.T) {
	// Original
	t.Run("RegisterFunc(original)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var isupper func(c rune) bool
		purego.RegisterLibFunc(&isupper, libc, "isupper")
		actual := isupper('A')
		if !actual {
			t.Errorf("isupper('%c'): expected true but got false", 'A')
		}
		actual = isupper('a')
		if actual {
			t.Errorf("isupper('%c'): expected false but got true", 'a')
		}
	})
	// New
	t.Run("RegisterFunc1_1(new)", func(t *testing.T) {
		library, err := getSystemLibrary()
		if err != nil {
			t.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			t.Errorf("failed to dlopen: %s", err)
		}
		var isupper func(c rune) bool
		symbol := purego.Symbol(libc, "isupper")
		purego.RegisterFunc(&isupper, symbol)
		actual := isupper('A')
		if !actual {
			t.Errorf("isupper('%c'): expected true but got false", 'A')
		}
		actual = isupper('a')
		if actual {
			t.Errorf("isupper('%c'): expected false but got true", 'a')
		}
	})
}

func Benchmark_isupper(b *testing.B) {
	// Original
	b.Run("RegisterFunc(original)", func(b *testing.B) {
		// 3037436, 395.6 ns/op, 56 B/op, 4 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		var isupper func(c rune) bool
		purego.RegisterLibFunc(&isupper, libc, "isupper")
		for i := 0; i < b.N; i++ {
			_ = isupper('A')
		}
	})
	// New
	b.Run("RegisterFunc1_1(new)", func(b *testing.B) {
		// 10082194, 121.2 ns/op, 144 B/op, 1 allocs/op
		library, err := getSystemLibrary()
		if err != nil {
			b.Errorf("couldn't get system library: %s", err)
		}
		libc, err := openLibrary(library)
		if err != nil {
			b.Errorf("failed to dlopen: %s", err)
		}
		var isupper func(c rune) bool
		symbol := purego.Symbol(libc, "isupper")
		purego.RegisterFunc1_1(&isupper, symbol)
		for i := 0; i < b.N; i++ {
			_ = isupper('A')
		}
	})
}
