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

// encQualifiers are the method type qualifiers (const, in, inout, out, bycopy, byref, oneway)
// that may prefix a type encoding. They do not affect the calling convention.
const encQualifiers = "rnNoORV"

// signature returns the type encoding of a block, or false if the block does not export one.
// The string refers to the block's descriptor; it must be cloned to be kept beyond the call.
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

// splitSignature splits a method or block type encoding, such as "v24@?0q8d16",
// into the result type followed by the argument types, without qualifiers or frame offsets.
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

// cutDigits splits s after its leading decimal digits.
func cutDigits(s string) (digits, rest string) {
	rest = strings.TrimLeft(s, "0123456789")
	return s[:len(s)-len(rest)], rest
}

// encodingLen returns the length of the single type encoding at the start of s.
func encodingLen(s string) (int, error) {
	if s == "" {
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
		// an object with its class name: @"NSString"
		n, err := quotedLen(s[1:])
		return 1 + n, err
	case strings.HasPrefix(s, "@?<"):
		// a block with its signature: @?<v@?q>
		n, err := bracketLen(s[2:])
		return 2 + n, err
	case strings.HasPrefix(s, "@?"):
		return 2, nil
	}
	return 1, nil
}

// bracketLen returns the length of the bracketed encoding at the start of s,
// such as {name=type...}, (name=type...), [count type] or <signature>.
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
			// field and class names may contain brackets, as in @"<NSCopying>".
			n, err := quotedLen(s[i:])
			if err != nil {
				return 0, err
			}
			i += n - 1
		}
	}
	return 0, fmt.Errorf("unterminated type encoding %q", s)
}

// quotedLen returns the length of the quoted name at the start of s, including the quotes.
func quotedLen(s string) (int, error) {
	if end := strings.IndexByte(s[1:], '"'); end >= 0 {
		return end + 2, nil
	}
	return 0, fmt.Errorf("unterminated name in type encoding %q", s)
}

// abiScalar is one scalar member of a type, as the calling convention sees it.
type abiScalar struct {
	offset uintptr
	// kind is the size in bytes of an integer or pointer ('1', '2', '4' or '8'),
	// or 'f' or 'd' for floating point.
	kind byte
	// blank reports whether the member is in a blank (_) field of a Go struct.
	blank bool
}

// size returns the size of the scalar in bytes.
func (s abiScalar) size() uintptr {
	switch s.kind {
	case 'f':
		return 4
	case 'd':
		return 8
	}
	return uintptr(s.kind - '0')
}

// isFloat reports whether the scalar is passed in a floating point register.
func (s abiScalar) isFloat() bool {
	return s.kind == 'f' || s.kind == 'd'
}

// abiLayout is the layout of a type with nested structs and arrays flattened into scalars,
// so that a type encoding can be compared with a Go type. A void result has no size.
type abiLayout struct {
	size, align uintptr
	scalars     []abiScalar
	// blank holds the offset and size of each blank (_) field of a Go struct.
	// Callers use those for padding, which has no counterpart in a type encoding,
	// and for members they have no use for.
	blank [][2]uintptr
}

// append adds the members of l at offset.
func (a *abiLayout) append(l abiLayout, offset uintptr) {
	for _, s := range l.scalars {
		s.offset += offset
		a.scalars = append(a.scalars, s)
	}
	for _, r := range l.blank {
		a.blank = append(a.blank, [2]uintptr{offset + r[0], r[1]})
	}
}

// repeat returns the layout of an array of count elements laid out as l.
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

// encodingLayout returns the layout of a type encoding without qualifiers,
// with members at their natural alignment.
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
				// a field name
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

// goLayout is encodingLayout for a Go type.
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

// matches reports whether a value of the Go type laid out as g can be passed as the C type laid out as c.
//
// Every member must be at the same offset with the same kind on both sides, except that
// a blank (_) field of the Go type may cover any integer or pointer members, or none.
// Floating point members decide which registers a struct is passed in, so they must
// match even when blank.
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
