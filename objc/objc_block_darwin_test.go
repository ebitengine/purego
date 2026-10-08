// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"path/filepath"
	"structs"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/testlib"
	"github.com/ebitengine/purego/objc"
)

func ExampleNewBlock() {
	_, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		panic(err)
	}

	var count = 0
	block := objc.NewBlock(
		func(block objc.Block, line objc.ID, stop *bool) {
			count++
			fmt.Printf("LINE %d: %s\n", count, objc.Send[string](line, objc.RegisterName("UTF8String")))
			*stop = count == 3
		},
	)
	defer block.Release()

	lines := objc.ID(objc.GetClass("NSString")).Send(objc.RegisterName("stringWithUTF8String:"), "Alpha\nBeta\nGamma\nDelta\nEpsilon")
	defer lines.Send(objc.RegisterName("release"))

	lines.Send(objc.RegisterName("enumerateLinesUsingBlock:"), block)
	// Output:
	// LINE 1: Alpha
	// LINE 2: Beta
	// LINE 3: Gamma
}

func ExampleInvokeBlock() {
	type vector struct {
		_       structs.HostLayout
		X, Y, Z float64
	}

	block := objc.NewBlock(
		func(block objc.Block, v1, v2 *vector) *vector {
			return &vector{
				X: v1.Y*v2.Z - v1.Z*v2.Y,
				Y: v1.Z*v2.X - v1.X*v2.Z,
				Z: v1.X*v2.Y - v1.Y*v2.X,
			}
		},
	)
	defer block.Release()

	result, err := objc.InvokeBlock[*vector](
		block,
		&vector{X: 0.1, Y: 2.3, Z: 4.5},
		&vector{X: 6.7, Y: 8.9, Z: 0.1},
	)

	fmt.Printf("{%.2f %.2f %.2f} %v\n", result.X, result.Y, result.Z, err)
	// Output: {-39.82 30.14 -14.52} <nil>
}

func TestInvoke(t *testing.T) {
	t.Run("return an error when passing an invalid number of arguments", func(t *testing.T) {
		block := objc.NewBlock(func(_ objc.Block, a int32, b int32) int32 {
			return a + b
		})
		defer block.Release()

		if _, err := objc.InvokeBlock[int32](block, int32(8)); err == nil {
			t.Fatal(err)
		}
	})

	t.Run("return an error when passing an invalid return type", func(t *testing.T) {
		block := objc.NewBlock(func(_ objc.Block, a int32, b int32) int32 {
			return a + b
		})
		defer block.Release()

		if _, err := objc.InvokeBlock[string](block, int32(8), int32(2)); err == nil {
			t.Fatal(err)
		}
	})

	t.Run("add two int32's and returns the result", func(t *testing.T) {
		block := objc.NewBlock(func(_ objc.Block, a int32, b int32) int32 {
			return a + b
		})
		defer block.Release()

		result, err := objc.InvokeBlock[int32](block, int32(8), int32(2))
		if err != nil {
			t.Fatal(err)
		}
		if result != 10 {
			t.Fatalf("expected 10, got %d", result)
		}
	})

	t.Run("add two int32's and store the result in a variable", func(t *testing.T) {
		var result int32
		block := objc.NewBlock(func(_ objc.Block, a int32, b int32) {
			result = a + b
		})
		defer block.Release()

		block.Invoke(int32(8), int32(2))
		if result != 10 {
			t.Fatalf("expected 10, got %d", result)
		}
	})
}

func TestBlockCopyAndBlockRelease(t *testing.T) {
	t.Parallel()

	var refCount int
	block := objc.NewBlock(
		func(objc.Block) {
			refCount++
		},
	)
	defer block.Release()
	refCount++

	copies := make([]objc.Block, 17)
	copies[0] = block
	for index := 1; index < len(copies); index++ {
		if refCount != index {
			t.Fatalf("refCount: %d != %d", refCount, index)
		}

		copies[index] = copies[index-1].Copy()
		if copies[index] != block {
			t.Fatalf("Block.Copy(): %v != %v", copies[index], block)
		}
		copies[index].Invoke()
	}

	for _, copy := range copies[1:] {
		copy.Release()
		refCount--
	}
	refCount--

	block.Invoke()
	if refCount != 1 {
		t.Fatalf("refCount: %d != 1", refCount)
	}
}

func loadBlockFixture(t testing.TB) uintptr {
	t.Helper()
	library := filepath.Join(t.TempDir(), "block.dylib")
	if err := testlib.BuildSharedLib(t, "CC", library, filepath.Join("testdata", "block.m")); err != nil {
		t.Fatal(err)
	}
	lib, err := purego.Dlopen(library, purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		t.Fatal(err)
	}
	return lib
}

func TestInvokeForeignBlock(t *testing.T) {
	lib := loadBlockFixture(t)

	check := func(t *testing.T, block objc.Block) {
		t.Helper()
		block.Invoke(int64(20), 3.5) // the result is discarded; this must not panic
		// the block returns base + i + int64(f), where base is 100.
		got, err := objc.InvokeBlock[int64](block, int64(20), 3.5)
		if err != nil {
			t.Fatal(err)
		}
		if got != 123 {
			t.Errorf("InvokeBlock = %d, want 123", got)
		}
	}

	t.Run("heap", func(t *testing.T) {
		var heapBlock func(base int64) objc.Block
		purego.RegisterLibFunc(&heapBlock, lib, "purego_heap_block")
		block := heapBlock(100)
		defer block.Release()
		check(t, block)
	})

	t.Run("stack", func(t *testing.T) {
		var withStackBlock func(base int64, cb uintptr)
		purego.RegisterLibFunc(&withStackBlock, lib, "purego_with_stack_block")
		called := false
		cb := purego.NewCallback(func(block objc.Block) {
			called = true
			check(t, block)
		})
		withStackBlock(100, cb)
		if !called {
			t.Fatal("callback was not called")
		}
	})
}

func TestInvokeForeignBlockMismatch(t *testing.T) {
	lib := loadBlockFixture(t)
	var heapBlock func(base int64) objc.Block
	purego.RegisterLibFunc(&heapBlock, lib, "purego_heap_block")
	block := heapBlock(100)
	defer block.Release()

	for range 2 { // the second call is answered from the cache
		if _, err := objc.InvokeBlock[int64](block, int64(20)); err == nil {
			t.Error("missing argument: expected an error")
		}
	}
	if _, err := objc.InvokeBlock[int64](block, int64(20), int64(3)); err == nil {
		t.Error("integer for a double argument: expected an error")
	}
	if _, err := objc.InvokeBlock[float64](block, int64(20), 3.5); err == nil {
		t.Error("wrong result type: expected an error")
	}
	if _, err := objc.InvokeBlock[any](block, int64(20), 3.5); err == nil {
		t.Error("unsupported result type: expected an error")
	}
	if _, err := objc.InvokeBlock[int32](block, int64(20), 3.5); err == nil {
		t.Error("smaller result type: expected an error")
	}
	if _, err := objc.InvokeBlock[int64](block, 20, 3.5); err != nil {
		t.Errorf("int for an int64_t argument: %v", err)
	}
	if _, err := objc.InvokeBlock[int64](block, int32(20), 3.5); err == nil {
		t.Error("int32 for an int64_t argument: expected an error")
	}
	if _, err := objc.InvokeBlock[int64](block, nil, 3.5); err == nil {
		t.Error("nil argument: expected an error")
	}
	if got, err := objc.InvokeBlock[int64](block, int64(20), 3.5); err != nil || got != 123 {
		t.Errorf("InvokeBlock = %d, %v; want 123, nil", got, err)
	}
}

func TestInvokeForeignBlockStruct(t *testing.T) {
	lib := loadBlockFixture(t)
	var bigBlock func() objc.Block
	purego.RegisterLibFunc(&bigBlock, lib, "purego_big_block")
	block := bigBlock()
	defer block.Release()

	type big struct {
		_          structs.HostLayout
		a, b, c, d int64
	}
	got, err := objc.InvokeBlock[big](block, int64(1))
	if err != nil {
		t.Fatal(err)
	}
	if want := (big{a: 1, b: 2, c: 3, d: 4}); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}

	type small struct {
		_    structs.HostLayout
		a, b int64
	}
	if _, err := objc.InvokeBlock[small](block, int64(1)); err == nil {
		t.Error("smaller struct: expected an error")
	}
	type reordered struct {
		_       structs.HostLayout
		a, b, c int64
		d       float64
	}
	if _, err := objc.InvokeBlock[reordered](block, int64(1)); err == nil {
		t.Error("struct with a different field: expected an error")
	}

	defer func() {
		if recover() == nil {
			t.Error("Invoke on a block returning a struct: expected a panic")
		}
	}()
	block.Invoke(int64(1))
}

func TestInvokeForeignBlockStructPadding(t *testing.T) {
	lib := loadBlockFixture(t)
	var boolFloatBlock func() objc.Block
	purego.RegisterLibFunc(&boolFloatBlock, lib, "purego_boolfloat_block")
	block := boolFloatBlock()
	defer block.Release()

	type boolFloat struct {
		_ structs.HostLayout
		b bool
		_ [3]byte
		f float32
	}
	got, err := objc.InvokeBlock[boolFloat](block, float32(1.5))
	if err != nil {
		t.Fatal(err)
	}
	if !got.b || got.f != 3 {
		t.Errorf("got {%v %v}, want {true 3}", got.b, got.f)
	}
}

func TestInvokeForeignBlockFuncArgument(t *testing.T) {
	lib := loadBlockFixture(t)
	var fnptrBlock func() objc.Block
	purego.RegisterLibFunc(&fnptrBlock, lib, "purego_fnptr_block")
	block := fnptrBlock()
	defer block.Release()

	if _, err := objc.InvokeBlock[objc.ID](block, func() {}); err == nil {
		t.Error("expected an error for a func argument")
	}

	func() {
		defer func() {
			if recover() == nil {
				t.Error("Invoke with a func argument: expected a panic")
			}
		}()
		block.Invoke(func() {})
	}()

	called := 0
	cb := purego.NewCallback(func() { called++ })
	for range 5000 {
		block.Invoke(cb)
	}
	if called != 5000 {
		t.Errorf("called = %d, want 5000", called)
	}
}

func TestInvokeForeignBlockBlockArgument(t *testing.T) {
	lib := loadBlockFixture(t)
	var blockArgBlock func() objc.Block
	purego.RegisterLibFunc(&blockArgBlock, lib, "purego_blockarg_block")
	block := blockArgBlock()
	defer block.Release()

	var got int64
	handler := objc.NewBlock(func(_ objc.Block, x int64) { got = x })
	defer handler.Release()

	block.Invoke(handler, int64(41))
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func BenchmarkInvokeForeignBlock(b *testing.B) {
	lib := loadBlockFixture(b)
	var heapBlock func(base int64) objc.Block
	purego.RegisterLibFunc(&heapBlock, lib, "purego_heap_block")
	block := heapBlock(100)
	defer block.Release()

	b.Run("Invoke", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			block.Invoke(int64(20), 3.5)
		}
	})
	b.Run("InvokeBlock", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := objc.InvokeBlock[int64](block, int64(20), 3.5); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestInvokeForeignBlockManyArguments(t *testing.T) {
	lib := loadBlockFixture(t)
	var manyBlock func() objc.Block
	purego.RegisterLibFunc(&manyBlock, lib, "purego_many_block")
	block := manyBlock()
	defer block.Release()

	for range 2 { // the second call is answered from the cache
		got, err := objc.InvokeBlock[int64](block, int64(1), int64(2), int64(3), int64(4), int64(5), int64(6), int64(7), int64(8), 9.5, int64(10))
		if err != nil {
			t.Fatal(err)
		}
		if got != 55 {
			t.Errorf("got %d, want 55", got)
		}
		if _, err := objc.InvokeBlock[int64](block, int64(1), int64(2), int64(3), int64(4), int64(5), int64(6), int64(7), int64(8), 9.5, int32(10)); err == nil {
			t.Error("int32 for the last int64_t argument: expected an error")
		}
	}
}
