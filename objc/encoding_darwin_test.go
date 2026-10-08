// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"structs"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

type encodeTypeTestStruct struct {
	_ structs.HostLayout
	A int32
	B float64
}

type encodeTypeArrayTestStruct struct {
	_      structs.HostLayout
	Values [3]int32
}

type encodeTypeHandleTestStruct struct {
	_              structs.HostLayout
	Implementation IMP
	Variable       Ivar
	Metadata       Property
}

var encodeTypeTests = []struct {
	typ   reflect.Type
	cType string
	want  string
}{
	{reflect.TypeFor[bool](), "_Bool", "B"},
	{reflect.TypeFor[int8](), "signed char", "c"},
	{reflect.TypeFor[uint8](), "unsigned char", "C"},
	{reflect.TypeFor[int16](), "short", "s"},
	{reflect.TypeFor[uint16](), "unsigned short", "S"},
	{reflect.TypeFor[int32](), "int", "i"},
	{reflect.TypeFor[uint32](), "unsigned int", "I"},
	{reflect.TypeFor[int64](), "long long", "q"},
	{reflect.TypeFor[uint64](), "unsigned long long", "Q"},
	{reflect.TypeFor[int](), "long", "q"},
	{reflect.TypeFor[uint](), "unsigned long", "Q"},
	{reflect.TypeFor[uintptr](), "uintptr_t", "Q"},
	{reflect.TypeFor[float32](), "float", "f"},
	{reflect.TypeFor[float64](), "double", "d"},
	{reflect.TypeFor[string](), "char *", "*"},
	{reflect.TypeFor[unsafe.Pointer](), "void *", "^v"},
	{reflect.TypeFor[*int32](), "int *", "^i"},
	{reflect.TypeFor[**int32](), "int **", "^^i"},
	{reflect.TypeFor[ID](), "id", "@"},
	{reflect.TypeFor[Class](), "Class", "#"},
	{reflect.TypeFor[SEL](), "SEL", ":"},
	{reflect.TypeFor[IMP](), "IMP", "^?"},
	{reflect.TypeFor[Ivar](), "Ivar", "^{objc_ivar=}"},
	{reflect.TypeFor[Property](), "objc_property_t", "^{objc_property=}"},
	{reflect.TypeFor[*Ivar](), "Ivar *", "^^{objc_ivar}"},
	{reflect.TypeFor[*Property](), "objc_property_t *", "^^{objc_property}"},
	{reflect.TypeFor[*IMP](), "IMP *", "^^?"},
	{reflect.TypeFor[encodeTypeHandleTestStruct](), "struct encodeTypeHandleTestStruct", "{encodeTypeHandleTestStruct=^?^{objc_ivar}^{objc_property}}"},
	{reflect.TypeFor[encodeTypeTestStruct](), "struct encodeTypeTestStruct", "{encodeTypeTestStruct=id}"},
	{reflect.TypeFor[[3]int32](), "int[3]", "[3i]"},
	{reflect.TypeFor[[2][3]int32](), "int[2][3]", "[2[3i]]"},
	{reflect.TypeFor[encodeTypeArrayTestStruct](), "struct encodeTypeArrayTestStruct", "{encodeTypeArrayTestStruct=[3i]}"},
	{reflect.TypeFor[[0]uint8](), "unsigned char[0]", "[0C]"},
	{reflect.TypeFor[[2]encodeTypeTestStruct](), "struct encodeTypeTestStruct[2]", "[2{encodeTypeTestStruct=id}]"},
	{reflect.TypeFor[*[3]int32](), "int (*)[3]", "^[3i]"},
}

func TestEncodeType(t *testing.T) {
	for _, tt := range encodeTypeTests {
		got, err := encodeType(tt.typ, false)
		if err != nil {
			t.Errorf("encodeType(%v) returned error: %v", tt.typ, err)
			continue
		}
		if got != tt.want {
			t.Errorf("encodeType(%v) = %q; want %q (@encode(%s))", tt.typ, got, tt.want, tt.cType)
		}
	}
}

// TestEncodeTypeMatchesClang uses the C compiler as an oracle: it compiles a
// program that prints @encode for each C type in encodeTypeTests and checks
// that the expected encodings are the ones the compiler actually produces.
func TestEncodeTypeMatchesClang(t *testing.T) {
	out, err := exec.Command("go", "env", "CC").Output()
	if err != nil {
		t.Fatalf("go env CC: %v", err)
	}
	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		t.Skip("no C compiler to use as an @encode oracle")
	}
	if _, err := exec.LookPath(compiler); err != nil {
		t.Skipf("no C compiler to use as an @encode oracle: %v", err)
	}

	var src strings.Builder
	src.WriteString("#include <stdio.h>\n#include <stdint.h>\n#include <objc/runtime.h>\n")
	src.WriteString("struct encodeTypeTestStruct { int a; double b; };\n")
	src.WriteString("struct encodeTypeArrayTestStruct { int values[3]; };\n")
	src.WriteString("struct encodeTypeHandleTestStruct { IMP implementation; Ivar variable; objc_property_t metadata; };\n")
	src.WriteString("int main(void) {\n")
	for _, tt := range encodeTypeTests {
		fmt.Fprintf(&src, "\tprintf(\"%%s\\n\", @encode(%s));\n", tt.cType)
	}
	src.WriteString("\treturn 0;\n}\n")

	dir := t.TempDir()
	srcFile := filepath.Join(dir, "encode.m")
	if err := os.WriteFile(srcFile, []byte(src.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// Built for the host architecture rather than GOARCH, unlike the shared
	// libraries in the root package's tests: every darwin target Go supports
	// is LP64, so these encodings do not vary by architecture, and building
	// for the host avoids needing Rosetta to run an amd64 binary on arm64.
	exeFile := filepath.Join(dir, "encode")
	cmd := exec.Command(compiler, "-Wall", "-Werror", "-o", exeFile, srcFile)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile oracle: %v\n%q\n%s", err, cmd, out)
	}

	out, err = exec.Command(exeFile).Output()
	if err != nil {
		t.Fatalf("run oracle: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(lines) != len(encodeTypeTests) {
		t.Fatalf("oracle printed %d encodings; want %d", len(lines), len(encodeTypeTests))
	}
	for i, tt := range encodeTypeTests {
		if lines[i] != tt.want {
			t.Errorf("@encode(%s) = %q; want %q", tt.cType, lines[i], tt.want)
		}
	}
}

// TestEncodeFunc checks how encodeFunc assembles a method signature: the
// return type, the implicit self and _cmd parameters, and the argument order.
// encodeType is checked per type by TestEncodeType, so the cases here vary the
// shape of the signature rather than the types in it. The first case is the
// signature from issue #493, which was recorded as v@:LLLLLLLLS - declaring
// 64-bit arguments as 32-bit made NSInvocation and other consumers of the type
// encoding truncate values above 2^32.
func TestEncodeFunc(t *testing.T) {
	tests := []struct {
		name string
		fn   any
		want string
	}{
		{
			name: "void return, 64-bit args",
			fn:   func(_ ID, _ SEL, a1, a2, a3, a4, a5, a6, a7, a8 uint64, a9 uint16) {},
			want: "v@:QQQQQQQQS",
		},
		{
			name: "value return, mixed integer kinds",
			fn:   func(_ ID, _ SEL, a int, b int64, c uint, d uint64) int { return 0 },
			want: "q@:qqQQ",
		},
		{
			name: "uintptr preserves adjacent argument",
			fn:   func(_ ID, _ SEL, p uintptr, n int32) uintptr { return 0 },
			want: "Q@:Qi",
		},
		{
			name: "no arguments",
			fn:   func(_ ID, _ SEL) {},
			want: "v@:",
		},
		{
			name: "argument order is preserved",
			fn:   func(_ ID, _ SEL, a int8, b float64, c ID) Class { return 0 },
			want: "#@:cd@",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeFunc(tt.fn)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("encodeFunc = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestEncodeFuncErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   any
	}{
		{"not a func", 0},
		{"array argument", func(_ ID, _ SEL, v [3]int32) {}},
		{"array return", func(_ ID, _ SEL) [3]int32 { return [3]int32{} }},
		{"too many return values", func(_ ID, _ SEL) (int, int) { return 0, 0 }},
		{"missing self and _cmd", func() {}},
		{"missing _cmd", func(_ ID) {}},
		{"unencodable argument", func(_ ID, _ SEL, c chan int) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := encodeFunc(tt.fn); err == nil {
				t.Errorf("encodeFunc = %q; want an error", got)
			}
		})
	}
}

func TestEncodeArrayElementError(t *testing.T) {
	if _, err := encodeType(reflect.TypeFor[[2]func()](), false); err == nil {
		t.Error("want an error for an unencodable array element")
	}
}

func TestBlockEncodeArrayErrors(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[func(Block, [3]int32)](), reflect.TypeFor[func(Block) [3]int32]()} {
		t.Run(typ.String(), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("want a panic for a top-level array")
				}
			}()
			new(blockCache).encode(typ)
		})
	}
}

func TestIMPFailures(t *testing.T) {
	var nilFunction func(ID, SEL)
	tests := []struct {
		name         string
		fn           any
		message      string
		runtimeError bool
	}{
		{name: "nil", fn: nil, message: "runtime error: invalid memory address or nil pointer dereference", runtimeError: true},
		{name: "not a function", fn: 42, message: "objc: not a function"},
		{name: "missing arguments", fn: func() {}, message: "objc: NewIMP must take a (id, SEL) as its first two arguments; got func()"},
		{name: "wrong self", fn: func(int, SEL) {}, message: "objc: NewIMP must take a (id, SEL) as its first two arguments; got func(int, objc.SEL)"},
		{name: "wrong selector", fn: func(ID, int) {}, message: "objc: NewIMP must take a (id, SEL) as its first two arguments; got func(objc.ID, int)"},
		{name: "nil function", fn: nilFunction, message: "purego: function must not be nil"},
		{name: "unsupported result", fn: func(ID, SEL) [3]int32 { return [3]int32{} }, message: "purego: unsupported return type: func(objc.ID, objc.SEL) [3]int32"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value any
			func() {
				defer func() { value = recover() }()
				NewIMP(tt.fn)
			}()
			if value == nil {
				t.Error("NewIMP did not panic")
			} else {
				if got := fmt.Sprint(value); got != tt.message {
					t.Errorf("panic message = %q; want %q", got, tt.message)
				}
				if tt.runtimeError {
					if _, ok := value.(runtime.Error); !ok {
						t.Errorf("panic type = %T; want runtime.Error", value)
					}
					if got := reflect.TypeOf(value).String(); got != "runtime.errorString" {
						t.Errorf("panic type = %q; want runtime.errorString", got)
					}
				} else if reflect.TypeOf(value) != reflect.TypeFor[string]() {
					t.Errorf("panic type = %T; want string", value)
				}
			}
			imp, err := newIMP(tt.fn)
			if imp != 0 || err == nil {
				t.Errorf("newIMP = %v, %v; want zero and error", imp, err)
			} else if got, want := err.Error(), "objc: failed to create IMP: "+tt.message; got != want {
				t.Errorf("newIMP error = %q; want %q", got, want)
			}
		})
	}
}

func TestIMPInvocation(t *testing.T) {
	for _, tt := range []struct {
		name   string
		create func(any) (IMP, error)
	}{
		{name: "public", create: func(fn any) (IMP, error) { return NewIMP(fn), nil }},
		{name: "internal", create: newIMP},
	} {
		t.Run(tt.name, func(t *testing.T) {
			imp, err := tt.create(func(_ ID, _ SEL, value int32) int32 { return value + 1 })
			if err != nil {
				t.Fatal(err)
			}
			var call func(ID, SEL, int32) int32
			purego.RegisterFunc(&call, uintptr(imp))
			if got := call(0, 0, 41); got != 42 {
				t.Errorf("IMP result = %d; want 42", got)
			}
		})
	}
}
