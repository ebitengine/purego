// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || netbsd || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x))

package purego_test

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/testlib"
)

type wordCallbackCaller func(uintptr, int32, *uintptr, uintptr) uintptr

func loadWordCallbackLibrary(t testing.TB) (wordCallbackCaller, func(uintptr, *uintptr) int32) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "libcbtest.so")
	if err := testlib.BuildSharedLib(t, "CC", path, "testdata/libcbtest/callback_test.c", "-pthread"); err != nil {
		t.Fatal(err)
	}
	library, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := purego.Dlclose(library); err != nil {
			t.Error(err)
		}
	})
	var call wordCallbackCaller
	purego.RegisterLibFunc(&call, library, "callWordCallback")
	var thread func(uintptr, *uintptr) int32
	purego.RegisterLibFunc(&thread, library, "callWordCallbackOnThread")
	return call, thread
}

func hashWords(values ...uintptr) uintptr {
	hash := uintptr(0x9e37)
	for i, value := range values {
		hash = (hash^value)*33 + uintptr(i)
	}
	return hash
}

func wordCallbacks() []any {
	return []any{
		func() uintptr { return hashWords() },
		func(a uintptr) uintptr { return hashWords(a) },
		func(a, b uintptr) uintptr { return hashWords(a, b) },
		func(a, b, c uintptr) uintptr { return hashWords(a, b, c) },
		func(a, b, c, d uintptr) uintptr { return hashWords(a, b, c, d) },
		func(a, b, c, d, e uintptr) uintptr { return hashWords(a, b, c, d, e) },
		func(a, b, c, d, e, f uintptr) uintptr { return hashWords(a, b, c, d, e, f) },
		func(a, b, c, d, e, f, g uintptr) uintptr { return hashWords(a, b, c, d, e, f, g) },
		func(a, b, c, d, e, f, g, h uintptr) uintptr { return hashWords(a, b, c, d, e, f, g, h) },
		func(a, b, c, d, e, f, g, h, i uintptr) uintptr { return hashWords(a, b, c, d, e, f, g, h, i) },
	}
}

func TestWordCallbacks(t *testing.T) {
	call, _ := loadWordCallbackLibrary(t)
	for count, fn := range wordCallbacks() {
		t.Run(fmt.Sprintf("%dargs", count), func(t *testing.T) {
			callback := purego.NewCallback(fn)
			for _, highBits := range []uintptr{0, ^uintptr(0)} {
				var args [9]uintptr
				for i := range args {
					args[i] = highBits ^ uintptr(1<<i)
				}
				want := hashWords(args[:count]...) * 3
				if got := call(callback, int32(count), &args[0], 3); got != want {
					t.Fatalf("got %#x, want %#x", got, want)
				}
			}
		})
	}
}

func TestWordCallbackAllocations(t *testing.T) {
	call, _ := loadWordCallbackLibrary(t)
	args := [9]uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9}
	const iterations = 64
	for count, fn := range wordCallbacks() {
		t.Run(fmt.Sprintf("%dargs", count), func(t *testing.T) {
			callback := purego.NewCallback(fn)
			want := hashWords(args[:count]...) * iterations
			var got uintptr
			allocations := testing.AllocsPerRun(10, func() { got = call(callback, int32(count), &args[0], iterations) })
			if got != want {
				t.Fatalf("got %#x, want %#x", got, want)
			}
			// The single outgoing RegisterFunc call may allocate. The callbacks
			// themselves must not allocate once for each C-to-Go transition.
			if allocations >= iterations {
				t.Fatalf("%.0f allocations for %d callbacks", allocations, iterations)
			}
		})
	}
}

func TestWordCallbackForeignThread(t *testing.T) {
	_, thread := loadWordCallbackLibrary(t)
	data := make([]byte, 4096)
	data[0], data[len(data)-1] = 17, 29
	callback := purego.NewCallback(func() uintptr {
		runtime.GC()
		return uintptr(data[0]) + uintptr(data[len(data)-1])
	})
	for range 5 {
		var result uintptr
		if rc := thread(callback, &result); rc != 0 {
			t.Fatalf("pthread error %d", rc)
		}
		if result != 46 {
			t.Fatal(result)
		}
	}
}

func TestWordCallbackConcurrentAndNested(t *testing.T) {
	call, _ := loadWordCallbackLibrary(t)
	inner := purego.NewCallback(func(a uintptr) uintptr { return a + 1 })
	outer := purego.NewCallback(func(a uintptr) uintptr {
		// Registration and nested native calls must not retain cbs.lock or
		// share scratch argument storage with the outer callback.
		_ = purego.NewCallback(func() uintptr { return 0 })
		arg := a
		return call(inner, 1, &arg, 1) + 1
	})
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			arg := uintptr(i)
			if got := call(outer, 1, &arg, 1); got != arg+2 {
				t.Errorf("got %d, want %d", got, arg+2)
			}
		})
	}
	wg.Wait()
}

func TestWordCallbackFallback(t *testing.T) {
	// These signatures are valid NewCallback inputs and must retain generic
	// dispatch semantics: named types, CDecl, pointers, floats and narrow ints.
	type word uintptr
	type namedFunction func(uintptr) uintptr
	cases := []struct {
		name   string
		fn     any
		invoke func(uintptr) uintptr
		want   uintptr
	}{
		{"NamedFunction", namedFunction(func(v uintptr) uintptr { return v + 1 }), func(p uintptr) uintptr {
			var f func(uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(41)
		}, 42},
		{"NamedWord", func(v word) word { return v + 1 }, func(p uintptr) uintptr {
			var f func(word) word
			purego.RegisterFunc(&f, p)
			return uintptr(f(41))
		}, 42},
		{"CDecl", func(_ purego.CDecl, v uintptr) uintptr { return v + 1 }, func(p uintptr) uintptr {
			var f func(uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(41)
		}, 42},
		{"Pointer", func(p *uintptr) uintptr { return *p }, func(p uintptr) uintptr {
			var f func(*uintptr) uintptr
			purego.RegisterFunc(&f, p)
			v := uintptr(42)
			return f(&v)
		}, 42},
		{"Float", func(v float64) uintptr { return uintptr(v * 2) }, func(p uintptr) uintptr {
			var f func(float64) uintptr
			purego.RegisterFunc(&f, p)
			return f(21)
		}, 42},
		{"NarrowInteger", func(v int32) int32 { return v - 1 }, func(p uintptr) uintptr {
			var f func(int32) int32
			purego.RegisterFunc(&f, p)
			return uintptr(-f(-41))
		}, 42},
		{"TenArguments", func(a, b, c, d, e, f, g, h, i, j uintptr) uintptr { return a + b + c + d + e + f + g + h + i + j }, func(p uintptr) uintptr {
			var f func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
		}, 55},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.invoke(purego.NewCallback(tc.fn)); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func BenchmarkWordCallback(b *testing.B) {
	call, _ := loadWordCallbackLibrary(b)
	args := [9]uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9}
	const iterations = 256
	for count, fn := range wordCallbacks() {
		b.Run(fmt.Sprintf("%dargs", count), func(b *testing.B) {
			callback := purego.NewCallback(fn)
			want := hashWords(args[:count]...) * iterations
			var got uintptr
			b.ReportAllocs()
			for b.Loop() {
				got = call(callback, int32(count), &args[0], iterations)
			}
			if got != want {
				b.Fatalf("got %#x, want %#x", got, want)
			}
			b.ReportMetric(float64(iterations), "callbacks/op")
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/iterations, "ns/callback")
		})
	}
}
