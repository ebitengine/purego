// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"testing"

	"github.com/ebitengine/purego"
)

// BenchmarkCallingMethods compares RegisterFunc, SyscallN, and Callback methods
func BenchmarkCallingMethods(b *testing.B) {
	testCases := []struct {
		n             int
		goFn          any
		goFnPtr       uintptr
		cFnPtr        uintptr
		cFnName       string
		cCallbackPtr  uintptr
		cCallbackName string
		args          []int64
		expectedSum   int64
	}{
		{1, goSum1, 0, 0, "sum1_c", 0, "call_callback1", []int64{1}, 1},
		{2, goSum2, 0, 0, "sum2_c", 0, "call_callback2", []int64{1, 2}, 3},
		{3, goSum3, 0, 0, "sum3_c", 0, "call_callback3", []int64{1, 2, 3}, 6},
		{5, goSum5, 0, 0, "sum5_c", 0, "call_callback5", []int64{1, 2, 3, 4, 5}, 15},
		{10, goSum10, 0, 0, "sum10_c", 0, "call_callback10", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 55},
	}

	libHandle := openBenchmarkLibrary(b)

	// Create callbacks and load C functions
	for i := range testCases {
		testCases[i].goFnPtr = purego.NewCallback(testCases[i].goFn)

		testCases[i].cFnPtr = openBenchmarkSymbol(b, libHandle, testCases[i].cFnName)
		testCases[i].cCallbackPtr = openBenchmarkSymbol(b, libHandle, testCases[i].cCallbackName)
	}

	b.Run("RegisterFunc/Callback", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				registerFn := makeRegisterFunc(tc.n)
				purego.RegisterFunc(registerFn, tc.goFnPtr)

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
				args := int64sToUintptrs(tc.args)
				var result uintptr
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _, _ = purego.SyscallN(tc.goFnPtr, args...)
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
				args := int64sToUintptrs(tc.args)
				var result uintptr
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, _, _ = purego.SyscallN(tc.cFnPtr, args...)
				}
				b.StopTimer()
				if int64(result) != tc.expectedSum {
					b.Fatalf("SyscallN/CFunc: expected sum %d, got %d", tc.expectedSum, result)
				}
			})
		}
	})

	// Benchmark round-trip: Go → C → Go callback (realistic use case)
	b.Run("RoundTrip/GoC", func(b *testing.B) {
		for _, tc := range testCases {
			b.Run(fmt.Sprintf("%dargs", tc.n), func(b *testing.B) {
				b.ReportAllocs()
				// Build args: first arg is callback pointer, rest are the arguments
				args := int64sToUintptrs(tc.args)
				callbackArgs := make([]uintptr, len(args)+1)
				callbackArgs[0] = tc.goFnPtr
				copy(callbackArgs[1:], args)

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
	default:
		panic(fmt.Sprintf("unsupported arg count: %d", n))
	}
}

// callRegisterFunc calls the registered function with the appropriate number of arguments
func callRegisterFunc(registerFn any, n int, args []int64, iterations int) int64 {
	var result int64
	switch n {
	case 1:
		f := registerFn.(*func(int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(args[0])
		}
	case 2:
		f := registerFn.(*func(int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(args[0], args[1])
		}
	case 3:
		f := registerFn.(*func(int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(args[0], args[1], args[2])
		}
	case 5:
		f := registerFn.(*func(int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(args[0], args[1], args[2], args[3], args[4])
		}
	case 10:
		f := registerFn.(*func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64)
		for i := 0; i < iterations; i++ {
			result = (*f)(args[0], args[1], args[2], args[3], args[4],
				args[5], args[6], args[7], args[8], args[9])
		}
	default:
		panic(fmt.Sprintf("unsupported arg count: %d", n))
	}
	return result
}

// int64sToUintptrs converts []int64 to []uintptr for SyscallN
func int64sToUintptrs(args []int64) []uintptr {
	result := make([]uintptr, len(args))
	for i, v := range args {
		result[i] = uintptr(v)
	}
	return result
}

func goSum1(a1 int64) int64 { return a1 }

func goSum2(a1, a2 int64) int64 { return a1 + a2 }

func goSum3(a1, a2, a3 int64) int64 { return a1 + a2 + a3 }

func goSum5(a1, a2, a3, a4, a5 int64) int64 { return a1 + a2 + a3 + a4 + a5 }

func goSum10(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 int64) int64 {
	return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10
}

func BenchmarkInterleavedFloat32(b *testing.B) {
	libHandle := openBenchmarkLibrary(b)

	b.Run("5args_float_at_3", func(b *testing.B) {
		b.ReportAllocs()
		sym := openBenchmarkSymbol(b, libHandle, "rmsnorm_shape_c")
		var fn func(uintptr, uintptr, uintptr, float32, uintptr) int32
		purego.RegisterFunc(&fn, sym)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fn(10, 20, 30, 1.0, 40)
		}
	})

	b.Run("9args_float_at_4", func(b *testing.B) {
		b.ReportAllocs()
		sym := openBenchmarkSymbol(b, libHandle, "sdpa_shape_c")
		var fn func(uintptr, uintptr, uintptr, uintptr, float32, uintptr, uintptr, uintptr, uintptr) int32
		purego.RegisterFunc(&fn, sym)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			fn(1, 2, 3, 4, 1.0, 5, 6, 7, 8)
		}
	})
}
