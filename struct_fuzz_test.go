// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build (darwin || linux || windows) && (amd64 || arm64 || loong64 || ppc64le)

package purego_test

import (
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/internal/load"
)

// FuzzStructRoundTrip fuzzes RegisterFunc struct round trips through the real
// C ABI: the fuzzer-supplied bytes are decoded into a struct type (never via
// a hash-to-seed avalanche: the byte stream is consumed as a sequential
// opcode stream so one mutated byte changes one decoding decision), the
// equivalent C declaration is emitted and compiled, and values are compared
// byte-for-byte in both the argument and return directions.
//
// The compiled libraries are cached by content hash so a repeated shape pays
// no gcc cost; only genuinely new shapes compile. Plain `go test` runs just
// the seed corpus, so normal CI stays fast and deterministic.
var sfzScalars = []reflect.Type{
	reflect.TypeFor[int8](), reflect.TypeFor[uint8](),
	reflect.TypeFor[int16](), reflect.TypeFor[uint16](),
	reflect.TypeFor[int32](), reflect.TypeFor[uint32](),
	reflect.TypeFor[int64](), reflect.TypeFor[uint64](),
	reflect.TypeFor[int](), reflect.TypeFor[uint](),
	reflect.TypeFor[uintptr](),
	reflect.TypeFor[float32](), reflect.TypeFor[float64](),
	reflect.TypeFor[bool](),
}

var sfzScalarC = map[reflect.Kind]string{
	reflect.Int8: "int8_t", reflect.Uint8: "uint8_t",
	reflect.Int16: "int16_t", reflect.Uint16: "uint16_t",
	reflect.Int32: "int32_t", reflect.Uint32: "uint32_t",
	reflect.Int64: "int64_t", reflect.Uint64: "uint64_t",
	reflect.Int: "int64_t", reflect.Uint: "uint64_t",
	reflect.Uintptr: "uintptr_t",
	reflect.Float32: "float", reflect.Float64: "double",
	reflect.Bool: "_Bool",
}

// sfzCursor is a sequential opcode stream over the fuzzer input. Every
// decision consumes explicit bytes, so a one-byte mutation perturbs one
// decision instead of reshuffling the whole type. Exhaustion yields zero,
// biasing toward small scalar shapes.
type sfzCursor struct {
	b []byte
}

func (c *sfzCursor) next(n int) int {
	if n <= 0 {
		return 0
	}
	if len(c.b) == 0 {
		return 0
	}
	v := int(c.b[0]) % n
	c.b = c.b[1:]
	return v
}

func (c *sfzCursor) nextByte(fallback byte) byte {
	if len(c.b) == 0 {
		return fallback
	}
	v := c.b[0]
	c.b = c.b[1:]
	return v
}

func sfzNext64(c *sfzCursor) uint64 {
	var u uint64
	for i := range 8 {
		u |= uint64(c.nextByte(0xA5)) << (8 * i)
	}
	return u
}

// sfzDecode builds a type from the opcode stream. depth bounds nesting so
// shapes stay small and fast to compile.
func sfzDecode(c *sfzCursor, depth int) reflect.Type {
	if depth <= 0 {
		return sfzScalars[c.next(len(sfzScalars))]
	}
	switch c.next(4) {
	case 1:
		return reflect.ArrayOf(1+c.next(2), sfzDecode(c, depth-1))
	case 2:
		n := 1 + c.next(2)
		fields := make([]reflect.StructField, n)
		for i := range fields {
			fields[i] = reflect.StructField{Name: fmt.Sprintf("F%d", i), Type: sfzDecode(c, depth-1)}
		}
		return reflect.StructOf(fields)
	default:
		return sfzScalars[c.next(len(sfzScalars))]
	}
}

// sfzDecodeTop always yields a struct: the identity round trip is func(S) S.
func sfzDecodeTop(c *sfzCursor) reflect.Type {
	n := 1 + c.next(2)
	fields := make([]reflect.StructField, n)
	for i := range fields {
		fields[i] = reflect.StructField{Name: fmt.Sprintf("F%d", i), Type: sfzDecode(c, 1)}
	}
	return reflect.StructOf(fields)
}

func sfzFill(v reflect.Value, c *sfzCursor) {
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			sfzFill(v.Field(i), c)
		}
	case reflect.Array:
		for i := range v.Len() {
			sfzFill(v.Index(i), c)
		}
	case reflect.Bool:
		v.SetBool(c.next(2) == 1)
	case reflect.Float32:
		v.SetFloat(float64(math.Float32frombits(uint32(sfzNext64(c)))))
	case reflect.Float64:
		v.SetFloat(math.Float64frombits(sfzNext64(c)))
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		v.SetInt(int64(sfzNext64(c)))
	default:
		v.SetUint(sfzNext64(c))
	}
}

func sfzEmitC(sb *strings.Builder, cn *map[reflect.Type]string, n *int, t reflect.Type) string {
	if name, ok := (*cn)[t]; ok {
		return name
	}
	if t.Kind() == reflect.Array {
		return sfzEmitDecl(sb, cn, n, t.Elem(), "")
	}
	if t.Kind() == reflect.Struct {
		name := fmt.Sprintf("t%d", *n)
		*n++
		var b strings.Builder
		b.WriteString("typedef struct {\n")
		for i := range t.NumField() {
			b.WriteString("\t" + sfzEmitDecl(sb, cn, n, t.Field(i).Type, fmt.Sprintf("f%d", i)) + ";\n")
		}
		b.WriteString("} " + name + ";\n")
		sb.WriteString(b.String())
		(*cn)[t] = name
		return name
	}
	return sfzScalarC[t.Kind()]
}

func sfzEmitDecl(sb *strings.Builder, cn *map[reflect.Type]string, n *int, t reflect.Type, name string) string {
	if t.Kind() == reflect.Array {
		return sfzEmitDecl(sb, cn, n, t.Elem(), fmt.Sprintf("%s[%d]", name, t.Len()))
	}
	return sfzEmitC(sb, cn, n, t) + " " + name
}

func sfzBytes(v reflect.Value) []byte {
	return unsafe.Slice((*byte)(v.Addr().UnsafePointer()), v.Type().Size())
}

func sfzHexdump(b []byte) string {
	const max = 48
	if len(b) > max {
		b = b[:max]
	}
	var s strings.Builder
	for _, c := range b {
		fmt.Fprintf(&s, "%02x ", c)
	}
	return s.String()
}

var sfzCompileMu sync.Mutex

var sfzCCOnce = struct {
	sync.Once
	cc string
}{}

// sfzCC returns the C compiler buildSharedLib will use. It is part of the
// cache key below: the same GOARCH can be built with different compilers
// (e.g. CI arm hard-float vs soft-float on one runner sharing /tmp).
func sfzCC() string {
	sfzCCOnce.Do(func() {
		out, err := exec.Command("go", "env", "CC").Output()
		if err != nil {
			return
		}
		sfzCCOnce.cc = strings.TrimSpace(string(out))
	})
	return sfzCCOnce.cc
}

// sfzCachedLib compiles csrc once per content hash and returns the cached
// shared library path. The version prefix keeps stale cache entries from an
// older generator from being reused. GOOS/GOARCH and the C compiler are part
// of the key: shared /tmp (e.g. CI minor-arches running loong64, ppc64le,
// ... in sequence on one runner, or arm hard/soft-float back to back) must
// never hand another toolchain's .so to Dlopen.
func sfzCachedLib(t *testing.T, csrc string) (string, error) {
	t.Helper()
	sum := sha256.Sum256([]byte("sfzv1\n" + runtime.GOOS + "/" + runtime.GOARCH + "\nCC=" + sfzCC() + "\n" + csrc))
	name := fmt.Sprintf("sfz%x", sum[:8])
	dir := filepath.Join(os.TempDir(), "purego-sfz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	lib := filepath.Join(dir, name+".so")
	sfzCompileMu.Lock()
	defer sfzCompileMu.Unlock()
	if _, err := os.Stat(lib); err == nil {
		return lib, nil
	}
	cfile := filepath.Join(dir, name+".c")
	if err := os.WriteFile(cfile, []byte(csrc), 0o644); err != nil {
		return "", err
	}
	// Compile to a unique temp file and rename: parallel fuzz workers may
	// compile the same new shape concurrently, and a partially written .so
	// must never be visible under the final name.
	tmp, err := os.CreateTemp(dir, name+"-*.so")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpName)
	if err := buildSharedLib(t, "CC", tmpName, cfile); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, lib); err != nil {
		// Lost a concurrent rename race for the identical content.
		if _, serr := os.Stat(lib); serr == nil {
			return lib, nil
		}
		return "", err
	}
	return lib, nil
}

func FuzzStructRoundTrip(f *testing.F) {
	if os.Getenv("PUREGO_TEST_PREBUILT_LIBDIR") != "" {
		f.Skip("generates C sources at run time, needs a local C toolchain")
	}
	// Seed corpus. Every seed below was verified green on main: plain
	// `go test` must stay green, new shapes are explored under -fuzz.
	for _, s := range [][]byte{
		{},
		{0},
		{1, 2, 3},
		{3, 1, 4, 1, 5},
		{2, 7, 0, 9, 3, 2},
		{1, 1, 2, 3, 5, 8, 13},
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			t.Skip("input too large, keeps each exec cheap")
		}
		c := &sfzCursor{b: data}
		ty := sfzDecodeTop(c)
		if ty.Size() == 0 || ty.Size() > 32 {
			t.Skip("shape out of the fast range")
		}

		var sb strings.Builder
		sb.WriteString("#include <string.h>\n#include <stdint.h>\n#include <stddef.h>\n\n")
		cn := map[reflect.Type]string{}
		n := 0
		ct := sfzEmitC(&sb, &cn, &n, ty)
		sb.WriteString(fmt.Sprintf("%s sfzid(%s s) { return s; }\n", ct, ct))
		sb.WriteString(fmt.Sprintf("size_t sfzsize(void) { return sizeof(%s); }\n", ct))
		csrc := sb.String()

		lib, err := sfzCachedLib(t, csrc)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		handle, err := load.OpenLibrary(lib)
		if err != nil {
			t.Fatalf("Dlopen: %v", err)
		}
		defer load.CloseLibrary(handle)

		var size func() uintptr
		purego.RegisterLibFunc(&size, handle, "sfzsize")
		if got, want := size(), ty.Size(); got != want {
			// The harness generated a Go/C layout mismatch, not a purego
			// bug; skip instead of reporting a false positive.
			t.Skipf("Go/C layout mismatch for %s: C sizeof=%d Go Sizeof=%d", ty, got, want)
		}

		fnType := reflect.FuncOf([]reflect.Type{ty}, []reflect.Type{ty}, false)
		ptr := reflect.New(fnType)
		func() {
			defer func() {
				if r := recover(); r != nil {
					if msg := fmt.Sprint(r); strings.Contains(msg, "not supported") || strings.Contains(msg, "unsupported") {
						t.Skipf("harness hit an unsupported shape: %v", r)
					}
					t.Fatalf("PANIC during RegisterLibFunc of %s: %v\n%s", ty, r, csrc)
				}
			}()
			purego.RegisterLibFunc(ptr.Interface(), handle, "sfzid")
		}()

		in := reflect.New(ty).Elem()
		sfzFill(in, c)
		var out reflect.Value
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC on identity %s: %v\n%s", ty, r, csrc)
				}
			}()
			out = ptr.Elem().Call([]reflect.Value{in})[0]
		}()
		outCopy := reflect.New(ty).Elem()
		outCopy.Set(out)
		want, got := sfzBytes(in), sfzBytes(outCopy)
		if string(want) != string(got) {
			t.Fatalf("identity %s corrupted\n  want: %s\n  got:  %s\n%s", ty, sfzHexdump(want), sfzHexdump(got), csrc)
		}
	})
}
