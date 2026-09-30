// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"path/filepath"
	"structs"
	"testing"

	"github.com/ebitengine/purego"
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

// loadBlockFixture compiles testdata/block.m, which creates blocks in Objective-C.
func loadBlockFixture(t *testing.T) uintptr {
	t.Helper()
	library := filepath.Join(t.TempDir(), "block.dylib")
	if err := buildSharedLib(t, library, filepath.Join("testdata", "block.m")); err != nil {
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

	check := func(name string, block objc.Block) {
		t.Helper()
		block.Invoke(int64(20), 3.5) // the result is discarded; this must not panic
		// the block returns base + i + int64(f), where base is 100.
		got, err := objc.InvokeBlock[int64](block, int64(20), 3.5)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != 123 {
			t.Errorf("%s: InvokeBlock = %d, want 123", name, got)
		}
	}

	t.Run("heap", func(t *testing.T) {
		var heapBlock func(base int64) objc.Block
		purego.RegisterLibFunc(&heapBlock, lib, "purego_heap_block")
		block := heapBlock(100)
		defer block.Release()
		check("InvokeBlock", block)
	})

	t.Run("stack", func(t *testing.T) {
		var withStackBlock func(base int64, cb uintptr)
		purego.RegisterLibFunc(&withStackBlock, lib, "purego_with_stack_block")
		called := false
		cb := purego.NewCallback(func(block objc.Block) {
			called = true
			check("InvokeBlock", block)
		})
		withStackBlock(100, cb)
		if !called {
			t.Fatal("callback was not called")
		}
	})
}
