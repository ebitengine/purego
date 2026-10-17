// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || s390x))

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
)

// FuzzCallbackRoundTrip fuzzes NewCallback round trips through the real C
// ABI: the fuzzer-supplied bytes are decoded into a callback signature (never
// via a hash-to-seed avalanche: the byte stream is consumed as a sequential
// opcode stream so one mutated byte changes one decoding decision), a C
// forwarder calling the Go callback is emitted and compiled, and arguments
// plus return values are compared byte-for-byte in both directions.
//
// The compiled libraries are cached by content hash so a repeated shape pays
// no gcc cost; only genuinely new shapes compile. Plain `go test` runs just
// the seed corpus, so normal CI stays fast and deterministic.
//
// NOTE: every NewCallback slot is process-global and never released (at most
// 2000). Each fuzz exec registers exactly one callback, so a long fuzz run
// can exhaust the table; that exhaustion is reported by skipping, not as a
// failure.
var cfzScalars = []reflect.Type{
	reflect.TypeFor[int8](), reflect.TypeFor[uint8](),
	reflect.TypeFor[int16](), reflect.TypeFor[uint16](),
	reflect.TypeFor[int32](), reflect.TypeFor[uint32](),
	reflect.TypeFor[int64](), reflect.TypeFor[uint64](),
	reflect.TypeFor[int](), reflect.TypeFor[uint](),
	reflect.TypeFor[uintptr](),
	reflect.TypeFor[float32](), reflect.TypeFor[float64](),
	reflect.TypeFor[bool](),
}

// cfzRetScalars excludes float returns, which compileCallback rejects.
var cfzRetScalars = []reflect.Type{
	reflect.TypeFor[int8](), reflect.TypeFor[uint8](),
	reflect.TypeFor[int16](), reflect.TypeFor[uint16](),
	reflect.TypeFor[int32](), reflect.TypeFor[uint32](),
	reflect.TypeFor[int64](), reflect.TypeFor[uint64](),
	reflect.TypeFor[int](), reflect.TypeFor[uint](),
	reflect.TypeFor[uintptr](),
	reflect.TypeFor[bool](),
}

var cfzScalarC = map[reflect.Kind]string{
	reflect.Int8: "int8_t", reflect.Uint8: "uint8_t",
	reflect.Int16: "int16_t", reflect.Uint16: "uint16_t",
	reflect.Int32: "int32_t", reflect.Uint32: "uint32_t",
	reflect.Int64: "int64_t", reflect.Uint64: "uint64_t",
	reflect.Int: "int64_t", reflect.Uint: "uint64_t",
	reflect.Uintptr: "uintptr_t",
	reflect.Float32: "float", reflect.Float64: "double",
	reflect.Bool: "_Bool",
}

// cfzCursor is a sequential opcode stream over the fuzzer input. Every
// decision consumes explicit bytes, so a one-byte mutation perturbs one
// decision instead of reshuffling the whole signature. Exhaustion yields
// zero, biasing toward small scalar shapes.
type cfzCursor struct {
	b []byte
}

func (c *cfzCursor) next(n int) int {
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

func (c *cfzCursor) nextByte(fallback byte) byte {
	if len(c.b) == 0 {
		return fallback
	}
	v := c.b[0]
	c.b = c.b[1:]
	return v
}

func cfzNext64(c *cfzCursor) uint64 {
	var u uint64
	for i := range 8 {
		u |= uint64(c.nextByte(0xA5)) << (8 * i)
	}
	return u
}

// cfzStructOK mirrors ensureCallbackStructSupported: structs in callbacks
// are only supported on amd64/arm64 on some OSes.
func cfzStructOK() bool {
	if runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64" {
		return false
	}
	switch runtime.GOOS {
	case "android", "darwin", "ios", "linux", "windows":
		return true
	default:
		return false
	}
}

// cfzDecode builds a scalar or nested-struct type from the opcode stream.
// Arrays are excluded: top-level arrays decay in C and callback argument
// generation rejects them.
func cfzDecode(c *cfzCursor, depth int) reflect.Type {
	if depth <= 0 {
		return cfzScalars[c.next(len(cfzScalars))]
	}
	if !cfzStructOK() {
		return cfzScalars[c.next(len(cfzScalars))]
	}
	switch c.next(3) {
	case 1:
		n := 1 + c.next(2)
		fields := make([]reflect.StructField, n)
		for i := range fields {
			fields[i] = reflect.StructField{Name: fmt.Sprintf("F%d", i), Type: cfzDecode(c, depth-1)}
		}
		return reflect.StructOf(fields)
	default:
		return cfzScalars[c.next(len(cfzScalars))]
	}
}

func cfzFill(v reflect.Value, c *cfzCursor) {
	switch v.Kind() {
	case reflect.Struct:
		for i := range v.NumField() {
			cfzFill(v.Field(i), c)
		}
	case reflect.Array:
		for i := range v.Len() {
			cfzFill(v.Index(i), c)
		}
	case reflect.Bool:
		v.SetBool(c.next(2) == 1)
	case reflect.Float32:
		v.SetFloat(float64(math.Float32frombits(uint32(cfzNext64(c)))))
	case reflect.Float64:
		v.SetFloat(math.Float64frombits(cfzNext64(c)))
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		v.SetInt(int64(cfzNext64(c)))
	default:
		v.SetUint(cfzNext64(c))
	}
}

func cfzEmitC(sb *strings.Builder, cn *map[reflect.Type]string, n *int, t reflect.Type) string {
	if name, ok := (*cn)[t]; ok {
		return name
	}
	if t.Kind() == reflect.Struct {
		name := fmt.Sprintf("t%d", *n)
		*n++
		var b strings.Builder
		b.WriteString("typedef struct {\n")
		for i := range t.NumField() {
			b.WriteString("\t" + cfzEmitC(sb, cn, n, t.Field(i).Type) + fmt.Sprintf(" f%d;\n", i))
		}
		b.WriteString("} " + name + ";\n")
		sb.WriteString(b.String())
		(*cn)[t] = name
		return name
	}
	return cfzScalarC[t.Kind()]
}

func cfzBytes(v reflect.Value) []byte {
	return unsafe.Slice((*byte)(v.Addr().UnsafePointer()), v.Type().Size())
}

func cfzAlignUp(off, align uintptr) uintptr {
	return (off + align - 1) &^ (align - 1)
}

var cfzCompileMu sync.Mutex

var cfzCCOnce = struct {
	sync.Once
	cc string
}{}

// cfzCC returns the C compiler buildSharedLib will use. It is part of the
// cache key below: the same GOARCH can be built with different compilers
// (e.g. CI arm hard-float vs soft-float on one runner sharing /tmp).
func cfzCC() string {
	cfzCCOnce.Do(func() {
		out, err := exec.Command("go", "env", "CC").Output()
		if err != nil {
			return
		}
		cfzCCOnce.cc = strings.TrimSpace(string(out))
	})
	return cfzCCOnce.cc
}

// cfzCachedLib compiles csrc once per content hash and returns the cached
// shared library path, compiling to a temp file first so parallel fuzz
// workers never observe a partially written .so. GOOS/GOARCH and the C
// compiler are part of the key: shared /tmp (e.g. CI minor-arches running
// several QEMU targets in sequence on one runner, or arm hard/soft-float
// back to back) must never hand another toolchain's .so to Dlopen.
func cfzCachedLib(t *testing.T, csrc string) (string, error) {
	t.Helper()
	sum := sha256.Sum256([]byte("cfzv1\n" + runtime.GOOS + "/" + runtime.GOARCH + "\nCC=" + cfzCC() + "\n" + csrc))
	name := fmt.Sprintf("cfz%x", sum[:8])
	dir := filepath.Join(os.TempDir(), "purego-cfz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	lib := filepath.Join(dir, name+".so")
	cfzCompileMu.Lock()
	defer cfzCompileMu.Unlock()
	if _, err := os.Stat(lib); err == nil {
		return lib, nil
	}
	cfile := filepath.Join(dir, name+".c")
	if err := os.WriteFile(cfile, []byte(csrc), 0o644); err != nil {
		return "", err
	}
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
		if _, serr := os.Stat(lib); serr == nil {
			return lib, nil
		}
		return "", err
	}
	return lib, nil
}

func FuzzCallbackRoundTrip(f *testing.F) {
	if runtime.GOOS == "windows" {
		f.Skip("windows callbacks use stdcall with restricted argument types")
	}
	if os.Getenv("PUREGO_TEST_PREBUILT_LIBDIR") != "" {
		f.Skip("generates C sources at run time, needs a local C toolchain")
	}
	// Seed corpus. Every seed below was verified green on main: plain
	// `go test` must stay green, new shapes are explored under -fuzz.
	for _, s := range [][]byte{
		{},
		{0},
		{1, 2, 3},
		{2, 1, 0},
		{1, 3, 2, 1},
		{2, 0, 1, 2, 3},
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 48 {
			t.Skip("input too large, keeps each exec cheap")
		}
		c := &cfzCursor{b: data}

		// Return: 0 = void, 1 = scalar, 2 = struct (if supported).
		var retType reflect.Type
		switch c.next(3) {
		case 1:
			retType = cfzRetScalars[c.next(len(cfzRetScalars))]
		case 2:
			if cfzStructOK() {
				rt := cfzDecode(c, 1)
				if rt.Kind() == reflect.Struct {
					retType = rt
				}
			}
		}

		nParams := c.next(3)
		params := make([]reflect.Type, nParams)
		for i := range params {
			params[i] = cfzDecode(c, 1)
		}

		var sb strings.Builder
		sb.WriteString("#include <stdint.h>\n\n")
		cn := map[reflect.Type]string{}
		n := 0
		cParamTypes := make([]string, len(params))
		cParams := make([]string, len(params))
		for i, ty := range params {
			cParamTypes[i] = cfzEmitC(&sb, &cn, &n, ty)
			cParams[i] = fmt.Sprintf("%s a%d", cfzEmitC(&sb, &cn, &n, ty), i+1)
		}
		retC := "void"
		if retType != nil {
			retC = cfzEmitC(&sb, &cn, &n, retType)
		}
		sb.WriteString(fmt.Sprintf("typedef %s (*cfzcbt)(%s);\n", retC, strings.Join(cParamTypes, ", ")))
		args := make([]string, len(params))
		for i := range params {
			args[i] = fmt.Sprintf("a%d", i+1)
		}
		fwdParams := ""
		if len(cParams) > 0 {
			fwdParams = ", " + strings.Join(cParams, ", ")
		}
		if retType == nil {
			sb.WriteString(fmt.Sprintf("void cfzfwd(cfzcbt cb%s) { cb(%s); }\n", fwdParams, strings.Join(args, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("%s cfzfwd(cfzcbt cb%s) { return cb(%s); }\n", retC, fwdParams, strings.Join(args, ", ")))
		}
		csrc := sb.String()

		lib, err := cfzCachedLib(t, csrc)
		if err != nil {
			t.Fatalf("compile: %v", err)
		}
		handle, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			t.Fatalf("Dlopen: %v", err)
		}
		defer purego.Dlclose(handle)

		goParams := append([]reflect.Type{reflect.TypeFor[uintptr]()}, params...)
		var goResults []reflect.Type
		if retType != nil {
			goResults = []reflect.Type{retType}
		}
		fnType := reflect.FuncOf(goParams, goResults, false)
		cbType := reflect.FuncOf(params, goResults, false)

		total := uintptr(0)
		for _, ty := range params {
			total = cfzAlignUp(total, uintptr(ty.Align())) + uintptr(ty.Size())
		}
		buf := make([]byte, total+16)
		offsets := make([]uintptr, len(params))
		off := uintptr(0)
		for i, ty := range params {
			off = cfzAlignUp(off, uintptr(ty.Align()))
			offsets[i] = off
			off += uintptr(ty.Size())
		}

		retVal := reflect.New(reflect.TypeFor[struct{}]()).Elem()
		if retType != nil {
			retVal = reflect.New(retType).Elem()
			cfzFill(retVal, c)
		}
		want := make([]byte, total)
		recv := func(args []reflect.Value) []reflect.Value {
			for i, o := range offsets {
				v := args[i]
				if !v.CanAddr() {
					tmp := reflect.New(v.Type()).Elem()
					tmp.Set(v)
					v = tmp
				}
				copy(buf[o:], cfzBytes(v))
			}
			if retType == nil {
				return nil
			}
			return []reflect.Value{retVal}
		}
		var cb uintptr
		func() {
			defer func() {
				if r := recover(); r != nil {
					if strings.Contains(fmt.Sprint(r), "maximum number of callbacks") {
						t.Skipf("callback table exhausted, cf #521: %v", r)
					}
					t.Fatalf("PANIC in NewCallback for %s: %v\n%s", cbType, r, csrc)
				}
			}()
			cb = purego.NewCallback(reflect.MakeFunc(cbType, recv).Interface())
		}()

		fwd := reflect.New(fnType)
		func() {
			defer func() {
				if r := recover(); r != nil {
					if msg := fmt.Sprint(r); strings.Contains(msg, "not supported") || strings.Contains(msg, "unsupported") {
						t.Skipf("harness hit an unsupported shape: %v", r)
					}
					t.Fatalf("PANIC during RegisterLibFunc of %s: %v\n%s", fnType, r, csrc)
				}
			}()
			purego.RegisterLibFunc(fwd.Interface(), handle, "cfzfwd")
		}()

		callArgs := []reflect.Value{reflect.ValueOf(cb)}
		for i, o := range offsets {
			v := reflect.New(fnType.In(i + 1)).Elem()
			cfzFill(v, c)
			callArgs = append(callArgs, v)
			copy(want[o:], cfzBytes(v))
		}

		var result reflect.Value
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC on call %s: %v\n%s", fnType, r, csrc)
				}
			}()
			results := fwd.Elem().Call(callArgs)
			if retType != nil {
				result = reflect.New(retType).Elem()
				result.Set(results[0])
			}
		}()
		if string(buf[:total]) != string(want) {
			diff := -1
			for i := range int(total) {
				if buf[i] != want[i] {
					diff = i
					break
				}
			}
			t.Fatalf("callback received corrupted arguments at byte %d\n  want: % x\n  got:  % x\n%s", diff, want, buf[:total], csrc)
		}
		if retType != nil && string(cfzBytes(result)) != string(cfzBytes(retVal)) {
			t.Fatalf("callback return corrupted\n  want: % x\n  got:  % x\n%s", cfzBytes(retVal), cfzBytes(result), csrc)
		}
	})
}
