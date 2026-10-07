// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	stdstrings "strings"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// encQualifiers are the method type qualifiers (const, in, inout, out, bycopy, byref, oneway)
// that may prefix a type encoding. They do not affect the calling convention.
const encQualifiers = "rnNoORV"

// signature returns the type encoding of a block, or false if the block does not export one.
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
	sig := strings.GoString(*(*uintptr)(unsafe.Add(unsafe.Pointer(layout.descriptor), offset)))
	return sig, sig != ""
}

// splitSignature splits a method or block type encoding, such as "v24@?0q8d16",
// into the result type followed by the argument types, without qualifiers or frame offsets.
func splitSignature(sig string) ([]string, error) {
	var types []string
	for sig != "" {
		sig = stdstrings.TrimLeft(sig, encQualifiers)
		n, err := encodingLen(sig)
		if err != nil {
			return nil, err
		}
		types = append(types, sig[:n])
		sig = stdstrings.TrimLeft(sig[n:], "0123456789")
	}
	return types, nil
}

// encodingLen returns the length of the single type encoding at the start of s.
func encodingLen(s string) (int, error) {
	if s == "" {
		return 0, errors.New("missing type encoding")
	}
	switch s[0] {
	case '{', '(', '[':
		return bracketLen(s)
	case '^':
		rest := stdstrings.TrimLeft(s[1:], encQualifiers)
		n, err := encodingLen(rest)
		return len(s) - len(rest) + n, err
	case 'b':
		return 1 + len(s[1:]) - len(stdstrings.TrimLeft(s[1:], "0123456789")), nil
	case '@':
		switch {
		case len(s) > 1 && s[1] == '"':
			// an object with its class name: @"NSString"
			n, err := quotedLen(s[1:])
			return 1 + n, err
		case len(s) > 2 && s[1] == '?' && s[2] == '<':
			// a block with its signature: @?<v@?q>
			n, err := bracketLen(s[2:])
			return 2 + n, err
		case len(s) > 1 && s[1] == '?':
			return 2, nil
		}
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
	if end := stdstrings.IndexByte(s[1:], '"'); end >= 0 {
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

	// blank and blankRegions describe the blank (_) fields of a Go struct.
	// They are padding as far as the Go type is concerned, so the type encoding
	// may have nothing there, or a member that the Go type does not name.
	blank        []abiScalar
	blankRegions [][2]uintptr // offset and size
}

// append adds the members of l at offset.
func (a *abiLayout) append(l abiLayout, offset uintptr) {
	for _, s := range l.scalars {
		a.scalars = append(a.scalars, abiScalar{offset + s.offset, s.kind})
	}
	for _, s := range l.blank {
		a.blank = append(a.blank, abiScalar{offset + s.offset, s.kind})
	}
	for _, r := range l.blankRegions {
		a.blankRegions = append(a.blankRegions, [2]uintptr{offset + r[0], r[1]})
	}
}

func scalarLayout(kind byte) abiLayout {
	s := abiScalar{kind: kind}
	return abiLayout{size: s.size(), align: s.size(), scalars: []abiScalar{s}}
}

func alignUp(n, align uintptr) uintptr {
	return (n + align - 1) / align * align
}

// encodingLayout returns the layout of a type encoding, with members at their natural alignment.
func encodingLayout(enc string) (abiLayout, error) {
	enc = stdstrings.TrimLeft(enc, encQualifiers)
	if enc == "" {
		return abiLayout{}, errors.New("missing type encoding")
	}
	switch enc[0] {
	case 'v':
		return abiLayout{}, nil
	case 'c', 'C', 'B':
		return scalarLayout('1'), nil
	case 's', 'S':
		return scalarLayout('2'), nil
	case 'i', 'I', 'l', 'L': // long is encoded as a 32-bit quantity
		return scalarLayout('4'), nil
	case 'q', 'Q', '^', '*', '@', '#', ':':
		return scalarLayout('8'), nil
	case 'f', 'd':
		return scalarLayout(enc[0]), nil
	case '[':
		digits := len(enc) - 1 - len(stdstrings.TrimLeft(enc[1:], "0123456789"))
		var count uintptr
		for _, c := range enc[1 : 1+digits] {
			count = count*10 + uintptr(c-'0')
		}
		elem, err := encodingLayout(enc[1+digits : len(enc)-1])
		if err != nil {
			return abiLayout{}, err
		}
		layout := abiLayout{size: count * elem.size, align: elem.align}
		for i := range count {
			layout.append(elem, i*elem.size)
		}
		return layout, nil
	case '{':
		_, fields, ok := stdstrings.Cut(enc[1:len(enc)-1], "=")
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
	}
	return abiLayout{}, fmt.Errorf("unsupported type encoding %s", enc)
}

// goLayout is encodingLayout for a Go type.
func goLayout(typ reflect.Type) (abiLayout, error) {
	switch typ.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Uint8:
		return scalarLayout('1'), nil
	case reflect.Int16, reflect.Uint16:
		return scalarLayout('2'), nil
	case reflect.Int32, reflect.Uint32:
		return scalarLayout('4'), nil
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Uintptr,
		reflect.Pointer, reflect.UnsafePointer, reflect.String:
		return scalarLayout('8'), nil
	case reflect.Float32:
		return scalarLayout('f'), nil
	case reflect.Float64:
		return scalarLayout('d'), nil
	case reflect.Array:
		elem, err := goLayout(typ.Elem())
		if err != nil {
			return abiLayout{}, err
		}
		layout := abiLayout{size: typ.Size(), align: uintptr(typ.Align())}
		for i := range uintptr(typ.Len()) {
			layout.append(elem, i*elem.size)
		}
		return layout, nil
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
				// everything in a blank field is blank.
				field.blank = append(field.blank, field.scalars...)
				field.scalars = nil
				field.blankRegions = [][2]uintptr{{0, field.size}}
			}
			layout.append(field, f.Offset)
		}
		return layout, nil
	case reflect.Func:
		// RegisterFunc would create a callback for every call, and callbacks are never freed.
		return abiLayout{}, fmt.Errorf("objc: a %s argument to a block is not supported; create the callback once with purego.NewCallback and pass the uintptr", typ)
	}
	return abiLayout{}, fmt.Errorf("objc: unsupported block argument or result type %s", typ)
}

// matches reports whether a value of the Go type laid out as g can be passed as the C type laid out as c.
//
// Every member must be at the same offset with the same kind on both sides, except for
// the blank (_) fields of the Go type. Callers use those for padding, which has no
// counterpart in a type encoding, and for members they have no use for. A blank field
// may therefore cover any integer or pointer members, or none. Floating point members
// decide which registers a struct is passed in, so they must match even when blank.
func (c abiLayout) matches(g abiLayout) bool {
	if c.size != g.size {
		return false
	}
	for _, s := range c.scalars {
		if slices.Contains(g.scalars, s) {
			continue
		}
		if s.isFloat() {
			if !slices.Contains(g.blank, s) {
				return false
			}
			continue
		}
		if !slices.ContainsFunc(g.blankRegions, func(r [2]uintptr) bool {
			return s.offset >= r[0] && s.offset+s.size() <= r[0]+r[1]
		}) {
			return false
		}
	}
	for _, s := range g.scalars {
		if !slices.Contains(c.scalars, s) {
			return false
		}
	}
	for _, s := range g.blank {
		if s.isFloat() && !slices.Contains(c.scalars, s) {
			return false
		}
	}
	return true
}
