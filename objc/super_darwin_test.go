// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"structs"
	"sync/atomic"
	"testing"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

var superTestClassID atomic.Uint64

// TestSendSuperDispatch checks argument and receiver forwarding, return values,
// and generic super calls from direct and inherited methods.
func TestSendSuperDispatch(t *testing.T) {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		t.Fatal(err)
	}
	prefix := fmt.Sprintf("PuregoSuperTest%d", superTestClassID.Add(1))
	sel := objc.RegisterName("probe:")
	var baseCalls, child1Calls int
	var receiver objc.ID
	base, err := objc.RegisterClass(prefix+"Base", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
		{
			Cmd: sel,
			Fn: func(self objc.ID, cmd objc.SEL, value int) int {
				baseCalls++
				receiver = self
				return value + 3
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var child1 objc.Class
	child1, err = objc.RegisterClass(prefix+"Child1", base, nil, nil, []objc.MethodDef{
		{
			Cmd: sel,
			Fn: func(self objc.ID, cmd objc.SEL, value int) int {
				child1Calls++
				// Bound recursion so an incorrect class anchor fails without overflowing the stack.
				if child1Calls > 1 {
					return -1
				}

				return objc.SendSuper2[int](self, child1, cmd, value) + 5
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	child2, err := objc.RegisterClass(prefix+"Child2", child1, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, class := range []objc.Class{child1, child2} {
		baseCalls, child1Calls, receiver = 0, 0, 0
		object := objc.ID(class).Send(objc.RegisterName("new"))
		defer object.Send(objc.RegisterName("release"))
		if got := objc.Send[int](object, sel, 34); got != 42 {
			t.Errorf("class %v: result = %d, want 42", class, got)
		}
		if baseCalls != 1 || child1Calls != 1 {
			t.Errorf("class %v: base calls = %d, child1 calls = %d; want 1 each", class, baseCalls, child1Calls)
		}
		if receiver != object {
			t.Errorf("base receiver = %v, want %v", receiver, object)
		}
		// Check forwarding through the ID method too.
		receiver = 0
		if got := object.SendSuper2(child1, sel, 34); got != 37 {
			t.Errorf("class %v: SendSuper2 = %d, want 37", class, got)
		}
		if receiver != object {
			t.Errorf("SendSuper2 receiver = %v, want %v", receiver, object)
		}
	}
	// The legacy helpers retain their direct-instance behavior.
	object := objc.ID(child1).Send(objc.RegisterName("new"))
	defer object.Send(objc.RegisterName("release"))
	if got := object.SendSuper(sel, 34); got != 37 {
		t.Errorf("SendSuper = %d, want 37", got)
	}
	if got := objc.SendSuper[int](object, sel, 34); got != 37 {
		t.Errorf("SendSuper[int] = %d, want 37", got)
	}
}

func TestSendSuperStruct(t *testing.T) {
	arch := "arm64"
	if runtime.GOARCH == "amd64" {
		arch = "x86_64"
	}
	library := filepath.Join(t.TempDir(), "super.dylib")
	cmd := exec.Command("clang", "-dynamiclib", "-arch", arch, "-framework", "Foundation", "-o", library, "testdata/super.m")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile super fixture: %v\n%s", err, out)
	}
	// Objective-C retains the registered classes and their implementations.
	if _, err := purego.Dlopen(library, purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		t.Fatal(err)
	}
	type result struct {
		_       structs.HostLayout
		A, B, C int64
	}
	want := result{
		A: 12,
		B: 34,
		C: 56,
	}
	child1 := objc.GetClass("PuregoSuperStructChild1")
	for _, class := range []objc.Class{child1, objc.GetClass("PuregoSuperStructChild2")} {
		object := objc.ID(class).Send(objc.RegisterName("new"))
		defer object.Send(objc.RegisterName("release"))
		if got := objc.SendSuper2[result](object, child1, objc.RegisterName("result")); got != want {
			t.Errorf("class %v: result = %+v, want %+v", class, got, want)
		}
	}
}
