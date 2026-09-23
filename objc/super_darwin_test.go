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

func TestSendSuperDispatch(t *testing.T) {
	if _, err := purego.Dlopen("/System/Library/Frameworks/Foundation.framework/Foundation", purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
		t.Fatal(err)
	}
	for _, generic := range []bool{false, true} {
		t.Run(fmt.Sprintf("generic=%t", generic), func(t *testing.T) {
			prefix := fmt.Sprintf("PuregoSuperTest%d", superTestClassID.Add(1))
			sel := objc.RegisterName("probe:")
			var baseCalls, overrideCalls int
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
			var override objc.Class
			override, err = objc.RegisterClass(prefix+"Override", base, nil, nil, []objc.MethodDef{
				{
					Cmd: sel,
					Fn: func(self objc.ID, cmd objc.SEL, value int) int {
						overrideCalls++
						// Bound recursion so an incorrect class anchor fails without overflowing the stack.
						if overrideCalls > 1 {
							return -1
						}

						if generic {
							return objc.SendSuper2[int](self, override, cmd, value) + 5
						}
						return int(self.SendSuper2(override, cmd, value)) + 5
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			child, err := objc.RegisterClass(prefix+"Child", override, nil, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			for _, class := range []objc.Class{override, child} {
				baseCalls, overrideCalls, receiver = 0, 0, 0
				object := objc.ID(class).Send(objc.RegisterName("new"))
				defer object.Send(objc.RegisterName("release"))
				if got := objc.Send[int](object, sel, 34); got != 42 {
					t.Errorf("class %v: result = %d, want 42", class, got)
				}
				if baseCalls != 1 || overrideCalls != 1 {
					t.Errorf("class %v: base calls = %d, override calls = %d; want 1 each", class, baseCalls, overrideCalls)
				}
				if receiver != object {
					t.Errorf("base receiver = %v, want %v", receiver, object)
				}
			}
			// The legacy helpers retain their direct-instance behavior.
			object := objc.ID(override).Send(objc.RegisterName("new"))
			defer object.Send(objc.RegisterName("release"))
			if got := object.SendSuper(sel, 34); got != 37 {
				t.Errorf("SendSuper = %d, want 37", got)
			}
			if got := objc.SendSuper[int](object, sel, 34); got != 37 {
				t.Errorf("SendSuper[int] = %d, want 37", got)
			}
		})
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
