// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

// BenchmarkCallingMethods compares RegisterFunc, SyscallN, and Callback methods
func BenchmarkCallingMethods(b *testing.B) {
	testCases := []struct {
		n             int
		fn            any
		fnPtr         uintptr
		cFnPtr        uintptr
		cFnName       string
		cCallbackPtr  uintptr
		cCallbackName string
		args          []uintptr
		expectedSum   int64
	}{
		{1, sum1, 0, 0, "sum1_c", 0, "call_callback1", []uintptr{1}, 1},
		{2, sum2, 0, 0, "sum2_c", 0, "call_callback2", []uintptr{1, 2}, 3},
		{3, sum3, 0, 0, "sum3_c", 0, "call_callback3", []uintptr{1, 2, 3}, 6},
		{5, sum5, 0, 0, "sum5_c", 0, "call_callback5", []uintptr{1, 2, 3, 4, 5}, 15},
		{10, sum10, 0, 0, "sum10_c", 0, "call_callback10", []uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 55},
		{14, sum15, 0, 0, "sum14_c", 0, "call_callback14", []uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}, 105},
		{15, sum15, 0, 0, "sum15_c", 0, "call_callback15", []uintptr{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, 120},
	}

	// Build C library for benchmarking
	libFileName := filepath.Join(b.TempDir(), "libbenchmark.so")
	if err := buildSharedLib("CC", libFileName, filepath.Join("testdata", "benchmarktest", "benchmark.c")); err != nil {
		b.Skipf("Failed to build C library: %v", err)
	}
	b.Cleanup(func() {
		os.Remove(libFileName)
	})

	libHandle, err := load.OpenLibrary(libFileName)
	if err != nil {
		b.Fatalf("Failed to load C library: %v", err)
	}
	defer func() {
		if err := load.CloseLibrary(libHandle); err != nil {
			b.Fatalf("Failed to close library: %s", err)
		}
	}()

	// Create callbacks and load C functions
	for i := range testCases {
		testCases[i].fnPtr = purego.NewCallback(testCases[i].fn)

		cFn, err := load.OpenSymbol(libHandle, testCases[i].cFnName)
		if err != nil {
			b.Fatalf("Failed to load C function %s: %v", testCases[i].cFnName, err)
		}
		testCases[i].cFnPtr = cFn

		cCallbackFn, err := load.OpenSymbol(libHandle, testCases[i].cCallbackName)
		if err != nil {
			b.Fatalf("Failed to load C callback wrapper %s: %v", testCases[i].cCallbackName, err)
		}
		testCases[i].cCallbackPtr = cCallbackFn
	}

	b.Run("RegisterFunc/Callback", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				registerFn := makeRegisterFunc(tc.n)
				purego.RegisterFunc(registerFn, tc.fnPtr)

				b.ResetTimer()
				result := callRegisterFunc(registerFn, tc.n, tc.args, b.N)
				b.StopTimer()

				if result != tc.expectedSum {
					b.Fatalf("RegisterFunc/Callback: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})

	// Benchmark RegisterFunc with C functions
	b.Run("RegisterFunc/CFunc", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				registerFn := makeRegisterFunc(tc.n)
				purego.RegisterFunc(registerFn, tc.cFnPtr)

				b.ResetTimer()
				result := callRegisterFunc(registerFn, tc.n, tc.args, b.N)
				b.StopTimer()

				if result != tc.expectedSum {
					b.Fatalf("RegisterFunc/CFunc: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})

	// Benchmark SyscallN with Go callbacks
	b.Run("SyscallN/Callback", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				var result uintptr
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _, _ = purego.SyscallN(tc.fnPtr, tc.args...)
				}
				b.StopTimer()
				if int64(result) != tc.expectedSum {
					b.Fatalf("SyscallN/Callback: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})

	// Benchmark SyscallN with C functions
	b.Run("SyscallN/CFunc", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				var result uintptr
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _, _ = purego.SyscallN(tc.cFnPtr, tc.args...)
				}
				b.StopTimer()
				if int64(result) != tc.expectedSum {
					b.Fatalf("SyscallN/CFunc: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})

	// Benchmark round-trip: Go → C → Go callback (realistic use case)
	b.Run("RoundTrip", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				// Build args: first arg is callback pointer, rest are the arguments
				callbackArgs := make([]uintptr, len(tc.args)+1)
				callbackArgs[0] = tc.fnPtr
				copy(callbackArgs[1:], tc.args)

				// Skip if total args (callback + args) exceeds or meets limit
				// SyscallN has issues with exactly 15 or more arguments
				if len(callbackArgs) >= 15 {
					b.Skipf("Round-trip with %d args + callback (%d total) exceeds/meets SyscallN limit", tc.n, len(callbackArgs))
				}

				var result uintptr
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _, _ = purego.SyscallN(tc.cCallbackPtr, callbackArgs...)
				}
				b.StopTimer()
				if int64(result) != tc.expectedSum {
					b.Fatalf("RoundTrip: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})
}

// makeRegisterFunc creates a function pointer of the appropriate signature
func makeRegisterFunc(n int) any {
	switch n {
	case 1:
		return new(func(int64) int64)
	case 2:
		return new(func(int64, int64) int64)
	case 3:
		return new(func(int64, int64, int64) int64)
	case 5:
		return new(func(int64, int64, int64, int64, int64) int64)
	case 10:
		return new(func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
	case 14:
		return new(func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
	case 15:
		return new(func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
	default:
		return nil
	}
}

// callRegisterFunc calls the registered function with the appropriate number of arguments
func callRegisterFunc(registerFn any, n int, args []uintptr, iterations int) int64 {
	var result int64
	switch n {
	case 1:
		f := registerFn.(*func(int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]))
		}
	case 2:
		f := registerFn.(*func(int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]))
		}
	case 3:
		f := registerFn.(*func(int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]), int64(args[2]))
		}
	case 5:
		f := registerFn.(*func(int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]), int64(args[2]), int64(args[3]), int64(args[4]))
		}
	case 10:
		f := registerFn.(*func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]), int64(args[2]), int64(args[3]), int64(args[4]),
				int64(args[5]), int64(args[6]), int64(args[7]), int64(args[8]), int64(args[9]))
		}
	case 14:
		f := registerFn.(*func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]), int64(args[2]), int64(args[3]), int64(args[4]),
				int64(args[5]), int64(args[6]), int64(args[7]), int64(args[8]), int64(args[9]),
				int64(args[10]), int64(args[11]), int64(args[12]), int64(args[13]))
		}
	case 15:
		f := registerFn.(*func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(int64(args[0]), int64(args[1]), int64(args[2]), int64(args[3]), int64(args[4]),
				int64(args[5]), int64(args[6]), int64(args[7]), int64(args[8]), int64(args[9]),
				int64(args[10]), int64(args[11]), int64(args[12]), int64(args[13]), int64(args[14]))
		}
	}
	return result
}

//go:noinline
func sum1(a1 int64) int64 { return a1 }

//go:noinline
func sum2(a1, a2 int64) int64 { return a1 + a2 }

//go:noinline
func sum3(a1, a2, a3 int64) int64 { return a1 + a2 + a3 }

//go:noinline
func sum5(a1, a2, a3, a4, a5 int64) int64 { return a1 + a2 + a3 + a4 + a5 }

//go:noinline
func sum10(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 int64) int64 {
	return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10
}

//go:noinline
func sum15(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15 int64) int64 {
	return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 + a14 + a15
}
