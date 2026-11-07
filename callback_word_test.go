// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || freebsd || illumos || netbsd || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x))

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

type wordCallbackLibrary struct {
	call   wordCallbackCaller
	thread func(uintptr, *uintptr) int32
	caller uintptr
}

func loadWordCallbackLibrary(t testing.TB) wordCallbackLibrary {
	t.Helper()
	path := filepath.Join(t.TempDir(), "libwordcbtest.so")
	if err := testlib.BuildSharedLib(t, "CC", path, "testdata/libcbtest/word_callback_test.c", "-pthread"); err != nil {
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
	caller, err := purego.Dlsym(library, "callWordCallback")
	if err != nil {
		t.Fatal(err)
	}
	lib := wordCallbackLibrary{caller: caller}
	purego.RegisterFunc(&lib.call, caller)
	purego.RegisterLibFunc(&lib.thread, library, "callWordCallbackOnThread")
	return lib
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

// NewCallback owns its slots for the process lifetime. Reuse test callbacks
// across tests, benchmark calibration, and repeated -count runs.
var registeredWordCallbacks = sync.OnceValue(func() []uintptr {
	var callbacks []uintptr
	for _, fn := range wordCallbacks() {
		callbacks = append(callbacks, purego.NewCallback(fn))
	}
	return callbacks
})

func TestWordCallbacks(t *testing.T) {
	call := loadWordCallbackLibrary(t).call
	for count, callback := range registeredWordCallbacks() {
		t.Run(fmt.Sprintf("%dargs", count), func(t *testing.T) {
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
	call := loadWordCallbackLibrary(t).call
	args := [9]uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9}
	const iterations = 64
	for count, callback := range registeredWordCallbacks() {
		t.Run(fmt.Sprintf("%dargs", count), func(t *testing.T) {
			want := hashWords(args[:count]...) * iterations
			var got uintptr
			baseline := testing.AllocsPerRun(10, func() { got = call(callback, int32(count), &args[0], 0) })
			if got != 0 {
				t.Fatalf("zero iterations: got %#x, want 0", got)
			}
			allocations := testing.AllocsPerRun(10, func() { got = call(callback, int32(count), &args[0], iterations) })
			if got != want {
				t.Fatalf("got %#x, want %#x", got, want)
			}
			// Exclude the outgoing RegisterFunc call's fixed allocation cost.
			if allocations != baseline {
				t.Fatalf("allocations grew from %.0f to %.0f for %d callbacks", baseline, allocations, iterations)
			}
		})
	}
}

var wordForeignThreadCallback = sync.OnceValue(func() uintptr {
	data := make([]byte, 4096)
	data[0], data[len(data)-1] = 17, 29
	return purego.NewCallback(func() uintptr {
		runtime.GC()
		return uintptr(data[0]) + uintptr(data[len(data)-1])
	})
})

func TestWordCallbackForeignThread(t *testing.T) {
	thread := loadWordCallbackLibrary(t).thread
	callback := wordForeignThreadCallback()
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

var wordNestedCallback = sync.OnceValue(func() uintptr {
	inner := purego.NewCallback(func(a uintptr) uintptr { return a + 1 })
	register := sync.OnceValue(func() uintptr {
		return purego.NewCallback(func() uintptr { return 0 })
	})
	return purego.NewCallback(func(caller, a uintptr) uintptr {
		// Registration and nested native calls must not retain cbs.lock or
		// share arguments with the outer callback. Pass the current library's
		// address so a cached callback never retains an unloaded C function.
		_ = register()
		var call wordCallbackCaller
		purego.RegisterFunc(&call, caller)
		arg := a
		return call(inner, 1, &arg, 1) + 1
	})
})

func TestWordCallbackConcurrentAndNested(t *testing.T) {
	lib := loadWordCallbackLibrary(t)
	outer := wordNestedCallback()
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Go(func() {
			args := [2]uintptr{lib.caller, uintptr(i)}
			if got := lib.call(outer, 2, &args[0], 1); got != uintptr(i)+2 {
				t.Errorf("got %d, want %d", got, i+2)
			}
		})
	}
	wg.Wait()
}

var wordRegistrationCallbacks = func() []func() uintptr {
	var callbacks []func() uintptr
	for i := range 16 {
		callbacks = append(callbacks, sync.OnceValue(func() uintptr {
			return purego.NewCallback(func(v uintptr) uintptr { return v + uintptr(i) })
		}))
	}
	return callbacks
}()

func TestWordCallbackConcurrentRegistration(t *testing.T) {
	call := loadWordCallbackLibrary(t).call
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i, register := range wordRegistrationCallbacks {
		wg.Go(func() {
			<-start
			callback := register()
			arg := uintptr(i)
			// Call newly published entries while other goroutines register and
			// invoke different callbacks in the same process-wide table.
			for range 32 {
				if got := call(callback, 1, &arg, 1); got != 2*arg {
					t.Errorf("got %d, want %d", got, 2*arg)
				}
			}
		})
	}
	close(start)
	wg.Wait()
}

type wordCallbackFallbackCase struct {
	name     string
	fn       any
	callback uintptr
	invoke   func(uintptr) uintptr
	want     uintptr
}

var wordCallbackFallbackCases = sync.OnceValue(func() []wordCallbackFallbackCase {
	// These signatures are valid NewCallback inputs and must retain generic
	// dispatch semantics: named types, CDecl, pointers, floats and narrow ints.
	type word uintptr
	type namedFunction func(uintptr) uintptr
	cases := []wordCallbackFallbackCase{
		{name: "NamedFunction", fn: namedFunction(func(v uintptr) uintptr { return v + 1 }), invoke: func(p uintptr) uintptr {
			var f func(uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(41)
		}, want: 42},
		{name: "NamedWord", fn: func(v word) word { return v + 1 }, invoke: func(p uintptr) uintptr {
			var f func(word) word
			purego.RegisterFunc(&f, p)
			return uintptr(f(41))
		}, want: 42},
		{name: "CDecl", fn: func(_ purego.CDecl, v uintptr) uintptr { return v + 1 }, invoke: func(p uintptr) uintptr {
			var f func(uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(41)
		}, want: 42},
		{name: "Pointer", fn: func(p *uintptr) uintptr { return *p }, invoke: func(p uintptr) uintptr {
			var f func(*uintptr) uintptr
			purego.RegisterFunc(&f, p)
			v := uintptr(42)
			return f(&v)
		}, want: 42},
		{name: "Float", fn: func(v float64) uintptr { return uintptr(v * 2) }, invoke: func(p uintptr) uintptr {
			var f func(float64) uintptr
			purego.RegisterFunc(&f, p)
			return f(21)
		}, want: 42},
		{name: "NarrowInteger", fn: func(v int32) int32 { return v - 1 }, invoke: func(p uintptr) uintptr {
			var f func(int32) int32
			purego.RegisterFunc(&f, p)
			return uintptr(-f(-41))
		}, want: 42},
		{name: "TenArguments", fn: func(a, b, c, d, e, f, g, h, i, j uintptr) uintptr { return a + b + c + d + e + f + g + h + i + j }, invoke: func(p uintptr) uintptr {
			var f func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) uintptr
			purego.RegisterFunc(&f, p)
			return f(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
		}, want: 55},
	}
	for i := range cases {
		cases[i].callback = purego.NewCallback(cases[i].fn)
	}
	return cases
})

func TestWordCallbackFallback(t *testing.T) {
	for _, tc := range wordCallbackFallbackCases() {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.invoke(tc.callback); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func BenchmarkWordCallback(b *testing.B) {
	call := loadWordCallbackLibrary(b).call
	args := [9]uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9}
	const iterations = 256
	for count, callback := range registeredWordCallbacks() {
		b.Run(fmt.Sprintf("%dargs", count), func(b *testing.B) {
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

func BenchmarkWordCallbackParallel(b *testing.B) {
	call := loadWordCallbackLibrary(b).call
	callback := registeredWordCallbacks()[5]
	const iterations = 256
	want := hashWords(1, 2, 3, 4, 5) * iterations
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		args := [5]uintptr{1, 2, 3, 4, 5}
		for pb.Next() {
			if got := call(callback, 5, &args[0], iterations); got != want {
				b.Errorf("got %#x, want %#x", got, want)
			}
		}
	})
	b.StopTimer()
	b.ReportMetric(iterations, "callbacks/op")
	// This is aggregate throughput across workers, not individual latency.
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/iterations, "ns/callback")
}
