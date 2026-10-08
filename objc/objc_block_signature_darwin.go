// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

// encQualifiers are method type qualifiers, which do not affect the calling convention.
const encQualifiers = "rnNoORV"

// signature returns a string that refers to the block's descriptor; clone it to keep it.
func (b Block) signature() (string, bool) {
	layout := *(**blockLayout)(unsafe.Pointer(&b))
	if layout.flags&blockHasSignature == 0 {
		return "", false
	}
	// The descriptor is { reserved, size, [copy, dispose,] [signature] } where the
	// helpers are only present with blockHasCopyDispose.
	offset := 2 * unsafe.Sizeof(uintptr(0))
	if layout.flags&blockHasCopyDispose != 0 {
		offset += 2 * unsafe.Sizeof(uintptr(0))
	}
	sig := *(**byte)(unsafe.Add(unsafe.Pointer(layout.descriptor), offset))
	if sig == nil {
		return "", false
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(sig), n)) != 0 {
		n++
	}
	return unsafe.String(sig, n), true
}

// splitSignature splits a type encoding such as "v24@?0q8d16" into the result
// and argument types, dropping qualifiers and frame offsets.
func splitSignature(sig string) ([]string, error) {
	var types []string
	for sig != "" {
		sig = strings.TrimLeft(sig, encQualifiers)
		n, err := encodingLen(sig)
		if err != nil {
			return nil, err
		}
		types = append(types, sig[:n])
		_, sig = cutDigits(sig[n:])
	}
	return types, nil
}

func cutDigits(s string) (digits, rest string) {
	rest = strings.TrimLeft(s, "0123456789")
	return s[:len(s)-len(rest)], rest
}

func encodingLen(s string) (int, error) {
	if s == "" || s[0] >= '0' && s[0] <= '9' {
		// Clang encodes some types, such as vectors, as an empty string.
		return 0, errors.New("missing type encoding")
	}
	switch {
	case strings.IndexByte("{([", s[0]) >= 0:
		return bracketLen(s)
	case s[0] == '^':
		rest := strings.TrimLeft(s[1:], encQualifiers)
		n, err := encodingLen(rest)
		return len(s) - len(rest) + n, err
	case s[0] == 'b':
		digits, _ := cutDigits(s[1:])
		return 1 + len(digits), nil
	case strings.HasPrefix(s, `@"`):
		// @"NSString"
		n, err := quotedLen(s[1:])
		return 1 + n, err
	case strings.HasPrefix(s, "@?<"):
		// @?<v@?q>
		n, err := bracketLen(s[2:])
		return 2 + n, err
	case strings.HasPrefix(s, "@?"):
		return 2, nil
	}
	return 1, nil
}

func bracketLen(s string) (int, error) {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{', '(', '[', '<':
			depth++
		case '}', ')', ']', '>':
			if depth--; depth == 0 {
				return i + 1, nil
			}
		case '"':
			// names may contain brackets, as in @"<NSCopying>".
			n, err := quotedLen(s[i:])
			if err != nil {
				return 0, err
			}
			i += n - 1
		}
	}
	return 0, fmt.Errorf("unterminated type encoding %q", s)
}

func quotedLen(s string) (int, error) {
	if end := strings.IndexByte(s[1:], '"'); end >= 0 {
		return end + 2, nil
	}
	return 0, fmt.Errorf("unterminated name in type encoding %q", s)
}

type abiScalar struct {
	offset uintptr
	// kind is '1', '2', '4' or '8' for an integer of that size, or 'f' or 'd'.
	kind  byte
	blank bool
}

func (s abiScalar) size() uintptr {
	switch s.kind {
	case 'f':
		return 4
	case 'd':
		return 8
	}
	return uintptr(s.kind - '0')
}

func (s abiScalar) isFloat() bool {
	return s.kind == 'f' || s.kind == 'd'
}

// abiLayout flattens nested structs and arrays so that a type encoding can be compared with a Go type.
type abiLayout struct {
	size, align uintptr
	scalars     []abiScalar
	// blank holds the offset and size of each _ field. Callers use those for
	// padding, which a type encoding does not have, and for members they ignore.
	blank [][2]uintptr
}

func (a *abiLayout) append(l abiLayout, offset uintptr) {
	for _, s := range l.scalars {
		s.offset += offset
		a.scalars = append(a.scalars, s)
	}
	for _, r := range l.blank {
		a.blank = append(a.blank, [2]uintptr{offset + r[0], r[1]})
	}
}

func (l abiLayout) repeat(count uintptr) abiLayout {
	array := abiLayout{size: count * l.size, align: l.align}
	for i := range count {
		array.append(l, i*l.size)
	}
	return array
}

func alignUp(n, align uintptr) uintptr {
	return (n + align - 1) / align * align
}

// encodingLayout assumes members are at their natural alignment.
func encodingLayout(enc string) (abiLayout, error) {
	if enc == "" {
		return abiLayout{}, errors.New("missing type encoding")
	}
	var kind byte
	switch enc[0] {
	case 'v':
		return abiLayout{}, nil
	case 'c', 'C', 'B':
		kind = '1'
	case 's', 'S':
		kind = '2'
	case 'i', 'I', 'l', 'L': // long is encoded as a 32-bit quantity
		kind = '4'
	case 'q', 'Q', '^', '*', '@', '#', ':':
		kind = '8'
	case 'f', 'd':
		kind = enc[0]
	case '[':
		digits, elem := cutDigits(enc[1 : len(enc)-1])
		count, err := strconv.Atoi(digits)
		if err != nil {
			return abiLayout{}, fmt.Errorf("array %s has no count", enc)
		}
		layout, err := encodingLayout(elem)
		return layout.repeat(uintptr(count)), err
	case '{':
		_, fields, ok := strings.Cut(enc[1:len(enc)-1], "=")
		if !ok {
			return abiLayout{}, fmt.Errorf("struct %s has no fields", enc)
		}
		layout := abiLayout{align: 1}
		for fields != "" {
			if fields[0] == '"' {
				n, err := quotedLen(fields)
				if err != nil {
					return abiLayout{}, err
				}
				fields = fields[n:]
				continue
			}
			n, err := encodingLen(fields)
			if err != nil {
				return abiLayout{}, err
			}
			field, err := encodingLayout(fields[:n])
			if err != nil {
				return abiLayout{}, err
			}
			fields = fields[n:]
			if field.size == 0 {
				continue
			}
			layout.size = alignUp(layout.size, field.align)
			layout.append(field, layout.size)
			layout.size += field.size
			layout.align = max(layout.align, field.align)
		}
		layout.size = alignUp(layout.size, layout.align)
		return layout, nil
	default:
		return abiLayout{}, fmt.Errorf("unsupported type encoding %s", enc)
	}
	s := abiScalar{kind: kind}
	return abiLayout{size: s.size(), align: s.size(), scalars: []abiScalar{s}}, nil
}

func goLayout(typ reflect.Type) (abiLayout, error) {
	switch typ.Kind() {
	case reflect.Array:
		elem, err := goLayout(typ.Elem())
		return elem.repeat(uintptr(typ.Len())), err
	case reflect.Struct:
		layout := abiLayout{size: typ.Size(), align: uintptr(typ.Align())}
		for i := range typ.NumField() {
			f := typ.Field(i)
			if f.Type.Size() == 0 {
				// structs.HostLayout and other zero-sized fields have no counterpart in C.
				continue
			}
			field, err := goLayout(f.Type)
			if err != nil {
				return abiLayout{}, err
			}
			if f.Name == "_" {
				for i := range field.scalars {
					field.scalars[i].blank = true
				}
				field.blank = [][2]uintptr{{0, field.size}}
			}
			layout.append(field, f.Offset)
		}
		return layout, nil
	case reflect.Func:
		// RegisterFunc would create a callback for every call, and callbacks are never freed.
		return abiLayout{}, fmt.Errorf("objc: a %s argument to a block is not supported; create the callback once with purego.NewCallback and pass the uintptr", typ)
	}
	if enc, err := encodeType(typ, false); err == nil {
		return encodingLayout(enc)
	}
	return abiLayout{}, fmt.Errorf("objc: unsupported block argument or result type %s", typ)
}

// matches reports whether the Go layout g can be passed as the C layout c. A _ field
// may cover integer members, but floats must match even when blank because they
// decide which registers a struct is passed in.
func (c abiLayout) matches(g abiLayout) bool {
	if c.size != g.size {
		return false
	}
	same := func(s abiScalar) func(abiScalar) bool {
		return func(t abiScalar) bool { return t.offset == s.offset && t.kind == s.kind }
	}
	inBlank := func(s abiScalar) bool {
		return slices.ContainsFunc(g.blank, func(r [2]uintptr) bool {
			return s.offset >= r[0] && s.offset+s.size() <= r[0]+r[1]
		})
	}
	for _, s := range c.scalars {
		if !slices.ContainsFunc(g.scalars, same(s)) && (s.isFloat() || !inBlank(s)) {
			return false
		}
	}
	for _, s := range g.scalars {
		if (!s.blank || s.isFloat()) && !slices.ContainsFunc(c.scalars, same(s)) {
			return false
		}
	}
	return true
}
