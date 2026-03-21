// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package purego_test

import (
	"testing"

	"github.com/ebitengine/purego"
)

func BenchmarkAllocIntOnly5Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, int64, int64, int64, int64) int64
	purego.RegisterLibFunc(&fn, h, "sum5_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 4, 5)
	}
}

func BenchmarkAllocIntOnly10Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, int64, int64, int64, int64, int64, int64, int64, int64, int64) int64
	purego.RegisterLibFunc(&fn, h, "sum10_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	}
}

func BenchmarkAllocUintptrOnly10Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr) int32
	purego.RegisterLibFunc(&fn, h, "qmm_shape_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	}
}

func BenchmarkAllocTrailingFloat5Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, int64, int64, int64, int64, float32) int64
	purego.RegisterLibFunc(&fn, h, "weighted_sum5f_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 4, 5, 1.0)
	}
}

func BenchmarkAllocTrailingFloat3Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, int64, int64, float32) int64
	purego.RegisterLibFunc(&fn, h, "weighted_sum3f_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 1.0)
	}
}

func BenchmarkAllocInterleavedFloat4Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, float32, int64, int64) int64
	purego.RegisterLibFunc(&fn, h, "interleaved_if_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2.0, 3, 4)
	}
}

func BenchmarkAllocInterleavedFloat5Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(int64, float32, int64, float32, int64) int64
	purego.RegisterLibFunc(&fn, h, "interleaved_2f_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2.0, 3, 4.0, 5)
	}
}

func BenchmarkAllocSDPA9Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(uintptr, uintptr, uintptr, uintptr, float32, uintptr, uintptr, uintptr, uintptr) int32
	purego.RegisterLibFunc(&fn, h, "sdpa_shape_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 4, 1.0, 5, 6, 7, 8)
	}
}

func BenchmarkAllocRMSNorm5Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(uintptr, uintptr, uintptr, float32, uintptr) int32
	purego.RegisterLibFunc(&fn, h, "rmsnorm_shape_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 3, 1e-5, 4)
	}
}

type OptionalInt struct {
	Value    int32
	HasValue bool
}

type OptionalFloat struct {
	Value    float32
	HasValue int8
}

func BenchmarkAllocRoPE9Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(uintptr, uintptr, int32, bool, OptionalFloat, float32, int32, uintptr, uintptr) int32
	purego.RegisterLibFunc(&fn, h, "rope_shape_c")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(1, 2, 128, false, OptionalFloat{10000.0, 1}, 1.0, 0, 3, 4)
	}
}

func BenchmarkAllocGatherQMM13Args(b *testing.B) {
	h := openBenchmarkLibrary(b)
	var fn func(*int32, uintptr, uintptr, uintptr, uintptr, uintptr, uintptr, bool, OptionalInt, OptionalInt, *byte, bool, uintptr) int32
	purego.RegisterLibFunc(&fn, h, "gather_qmm_shape_c")
	var res int32
	mode := []byte("default\x00")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(&res, 1, 2, 3, 0, 0, 0, false,
			OptionalInt{64, true}, OptionalInt{4, true},
			&mode[0], false, 8)
	}
}
