// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

type encodeTypeTestStruct struct {
	A int32
	B float64
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
	{reflect.TypeFor[float32](), "float", "f"},
	{reflect.TypeFor[float64](), "double", "d"},
	{reflect.TypeFor[string](), "char *", "*"},
	{reflect.TypeFor[unsafe.Pointer](), "void *", "^v"},
	{reflect.TypeFor[*int32](), "int *", "^i"},
	{reflect.TypeFor[**int32](), "int **", "^^i"},
	{reflect.TypeFor[ID](), "id", "@"},
	{reflect.TypeFor[Class](), "Class", "#"},
	{reflect.TypeFor[SEL](), "SEL", ":"},
	{reflect.TypeFor[encodeTypeTestStruct](), "struct encodeTypeTestStruct", "{encodeTypeTestStruct=id}"},
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
	src.WriteString("#include <stdio.h>\n#include <objc/objc.h>\n")
	src.WriteString("struct encodeTypeTestStruct { int a; double b; };\n")
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
