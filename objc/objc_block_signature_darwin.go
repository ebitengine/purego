// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"errors"
	"fmt"
	"reflect"
	stdstrings "strings"
	"unsafe"

	"github.com/ebitengine/purego/internal/strings"
)

// encQualifiers are the method type qualifiers (const, in, inout, out, bycopy, byref, oneway)
// that may prefix a type encoding. They do not affect the calling convention.
const encQualifiers = "rnNoORV"

// abiVoid is the ABI form of no value. See encodingABI.
const abiVoid = "v"

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

// encodingABI reduces a type encoding to what the calling convention sees, so that it
// can be compared with goABI: each scalar becomes its size in bytes, or "f" or "d" for
// floating point, and a struct becomes its scalars in order between braces, with nested
// structs and arrays flattened. For example, {CGRect={CGPoint=dd}{CGSize=dd}} is "{dddd}".
func encodingABI(enc string) (string, error) {
	enc = stdstrings.TrimLeft(enc, encQualifiers)
	if enc == "" {
		return "", errors.New("missing type encoding")
	}
	switch enc[0] {
	case 'v':
		return abiVoid, nil
	case 'c', 'C', 'B':
		return "1", nil
	case 's', 'S':
		return "2", nil
	case 'i', 'I', 'l', 'L': // long is encoded as a 32-bit quantity
		return "4", nil
	case 'q', 'Q', '^', '*', '@', '#', ':':
		return "8", nil
	case 'f', 'd':
		return enc[:1], nil
	case '[':
		count := len(enc) - 1 - len(stdstrings.TrimLeft(enc[1:], "0123456789"))
		n := 0
		for _, c := range enc[1 : 1+count] {
			n = n*10 + int(c-'0')
		}
		elem, err := encodingABI(enc[1+count : len(enc)-1])
		if err != nil {
			return "", err
		}
		return stdstrings.Repeat(stdstrings.Trim(elem, "{}"), n), nil
	case '{':
		_, fields, ok := stdstrings.Cut(enc[1:len(enc)-1], "=")
		if !ok {
			return "", fmt.Errorf("struct %s has no fields", enc)
		}
		var abi stdstrings.Builder
		abi.WriteByte('{')
		for fields != "" {
			if fields[0] == '"' {
				// a field name
				n, err := quotedLen(fields)
				if err != nil {
					return "", err
				}
				fields = fields[n:]
				continue
			}
			n, err := encodingLen(fields)
			if err != nil {
				return "", err
			}
			field, err := encodingABI(fields[:n])
			if err != nil {
				return "", err
			}
			abi.WriteString(stdstrings.Trim(field, "{}"))
			fields = fields[n:]
		}
		abi.WriteByte('}')
		return abi.String(), nil
	}
	return "", fmt.Errorf("unsupported type encoding %s", enc)
}

// goABI is encodingABI for a Go type.
func goABI(typ reflect.Type) (string, error) {
	switch typ.Kind() {
	case reflect.Bool, reflect.Int8, reflect.Uint8:
		return "1", nil
	case reflect.Int16, reflect.Uint16:
		return "2", nil
	case reflect.Int32, reflect.Uint32:
		return "4", nil
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Uintptr,
		reflect.Pointer, reflect.UnsafePointer, reflect.String:
		return "8", nil
	case reflect.Float32:
		return "f", nil
	case reflect.Float64:
		return "d", nil
	case reflect.Array:
		elem, err := goABI(typ.Elem())
		if err != nil {
			return "", err
		}
		return stdstrings.Repeat(stdstrings.Trim(elem, "{}"), typ.Len()), nil
	case reflect.Struct:
		var abi stdstrings.Builder
		abi.WriteByte('{')
		for i := range typ.NumField() {
			f := typ.Field(i)
			if f.Type.Size() == 0 {
				// structs.HostLayout and other zero-sized fields have no counterpart in C.
				continue
			}
			field, err := goABI(f.Type)
			if err != nil {
				return "", err
			}
			abi.WriteString(stdstrings.Trim(field, "{}"))
		}
		abi.WriteByte('}')
		return abi.String(), nil
	case reflect.Func:
		// RegisterFunc would create a callback for every call, and callbacks are never freed.
		return "", fmt.Errorf("objc: a %s argument to a block is not supported; create the callback once with purego.NewCallback and pass the uintptr", typ)
	}
	return "", fmt.Errorf("objc: unsupported block argument or result type %s", typ)
}
