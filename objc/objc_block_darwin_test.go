// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc_test

import (
	"fmt"
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

// foreignBlock returns a block allocated by the runtime (not by NewBlock) that wraps inner.
func foreignBlock(t *testing.T, inner objc.Block) objc.Block {
	t.Helper()
	lib, err := purego.Dlopen("/usr/lib/libSystem.B.dylib", purego.RTLD_LAZY|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatal(err)
	}
	create, err := purego.Dlsym(lib, "dispatch_block_create")
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ := purego.SyscallN(create, 0, uintptr(inner))
	if r == 0 {
		t.Fatal("dispatch_block_create returned nil")
	}
	return objc.Block(r)
}

func TestInvokeForeignBlock(t *testing.T) {
	ran := false
	ours := objc.NewBlock(func(_ objc.Block) { ran = true })
	defer ours.Release()

	foreign := foreignBlock(t, ours)
	defer foreign.Release()
	if foreign == ours {
		t.Fatal("expected a distinct block")
	}

	foreign.Invoke()
	if !ran {
		t.Fatal("inner block was not run")
	}
}

// TestInvokeFrameworkCompletionHandler invokes a block created by Foundation.
// NSItemProvider passes a completion handler (a stack block) to the load handler
// given to registerItemForTypeIdentifier:loadHandler:.
func TestInvokeFrameworkCompletionHandler(t *testing.T) {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		t.Fatal(err)
	}
	str := func(s string) objc.ID {
		return objc.Send[objc.ID](objc.ID(objc.GetClass("NSString")), objc.RegisterName("stringWithUTF8String:"), s)
	}
	typeID := str("public.plain-text")
	provider := objc.Send[objc.ID](objc.ID(objc.GetClass("NSItemProvider")), objc.RegisterName("new"))
	defer objc.Send[objc.ID](provider, objc.RegisterName("release"))

	want := str("hello")
	load := objc.NewBlock(func(_ objc.Block, completion objc.Block, _ objc.Class, _ objc.ID) {
		completion.Invoke(want, objc.ID(0))
	})
	defer load.Release()
	objc.Send[objc.ID](provider, objc.RegisterName("registerItemForTypeIdentifier:loadHandler:"), typeID, load)

	done := make(chan objc.ID, 1)
	handler := objc.NewBlock(func(_ objc.Block, item, _ objc.ID) { done <- item })
	defer handler.Release()
	objc.Send[objc.ID](provider, objc.RegisterName("loadItemForTypeIdentifier:options:completionHandler:"), typeID, objc.ID(0), handler)

	// the load handler is dispatched to the main queue, so the main run loop must run.
	runLoop := objc.Send[objc.ID](objc.ID(objc.GetClass("NSRunLoop")), objc.RegisterName("mainRunLoop"))
	date := objc.Send[objc.ID](objc.ID(objc.GetClass("NSDate")), objc.RegisterName("dateWithTimeIntervalSinceNow:"), 5.0)
	for {
		select {
		case item := <-done:
			if item != want {
				t.Fatalf("got item %#x, want %#x", item, want)
			}
			return
		default:
		}
		mode := str("kCFRunLoopDefaultMode")
		soon := objc.Send[objc.ID](objc.ID(objc.GetClass("NSDate")), objc.RegisterName("dateWithTimeIntervalSinceNow:"), 0.05)
		objc.Send[bool](runLoop, objc.RegisterName("runMode:beforeDate:"), mode, soon)
		if objc.Send[float64](date, objc.RegisterName("timeIntervalSinceNow")) < 0 {
			t.Fatal("timed out waiting for the completion handler")
		}
	}
}

func TestInvokeBlockForeign(t *testing.T) {
	ran := false
	ours := objc.NewBlock(func(_ objc.Block) { ran = true })
	defer ours.Release()

	foreign := foreignBlock(t, ours)
	defer foreign.Release()

	// the wrapper returns nothing meaningful; just check the call goes through.
	if _, err := objc.InvokeBlock[int32](foreign); err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("inner block was not run")
	}
}
