// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc

import (
	"reflect"
	"slices"
	"structs"
	"testing"
)

func TestSplitSignature(t *testing.T) {
	tests := []struct {
		sig  string
		want []string
	}{
		{"v24@?0q8d16", []string{"v", "@?", "q", "d"}},
		{"q28@?0i8c12s16q20", []string{"q", "@?", "i", "c", "s", "q"}},
		{`v32@?0@"NSString"8@"NSError"16q24`, []string{"v", "@?", `@"NSString"`, `@"NSError"`, "q"}},
		{"v24@?0@?<v@?q>8q16", []string{"v", "@?", "@?<v@?q>", "q"}},
		{"v16@?0@?<v@?@?<v@?>>8", []string{"v", "@?", "@?<v@?@?<v@?>>"}},
		{`v16@?0@"<NSCopying>"8`, []string{"v", "@?", `@"<NSCopying>"`}},
		{"{Big=qqqq}16@?0q8", []string{"{Big=qqqq}", "@?", "q"}},
		{"v16@?0^{S=[2{T=i}]}8", []string{"v", "@?", "^{S=[2{T=i}]}"}},
		{"v24@?0r*8Vv16", []string{"v", "@?", "*", "v"}},
		{"v16@?0^r^v8", []string{"v", "@?", "^r^v"}},
		{`{P="x"d"y"d}8@?0`, []string{`{P="x"d"y"d}`, "@?"}},
	}
	for _, tt := range tests {
		got, err := splitSignature(tt.sig)
		if err != nil {
			t.Errorf("splitSignature(%q): %v", tt.sig, err)
			continue
		}
		if !slices.Equal(got, tt.want) {
			t.Errorf("splitSignature(%q) = %q, want %q", tt.sig, got, tt.want)
		}
	}

	for _, sig := range []string{"v8@?0{S=i", `v8@?0@"NSString`, "v8@?0@?<v"} {
		if got, err := splitSignature(sig); err == nil {
			t.Errorf("splitSignature(%q) = %q, want an error", sig, got)
		}
	}
}

func TestEncodingABI(t *testing.T) {
	tests := []struct {
		enc  string
		want string
	}{
		{"v", "v"},
		{"B", "1"},
		{"c", "1"},
		{"S", "2"},
		{"i", "4"},
		{"L", "4"},
		{"q", "8"},
		{"^v", "8"},
		{"@?<v@?q>", "8"},
		{`@"NSString"`, "8"},
		{"r*", "8"},
		{"f", "f"},
		{"d", "d"},
		{"{Big=qqqq}", "{8888}"},
		{"{CGRect={CGPoint=dd}{CGSize=dd}}", "{dddd}"},
		{`{P="x"d"y"i}`, "{d4}"},
		{"{S=[3f]c}", "{fff1}"},
		{"{S=[2{T=id}]}", "{4d4d}"},
	}
	for _, tt := range tests {
		got, err := encodingABI(tt.enc)
		if err != nil {
			t.Errorf("encodingABI(%q): %v", tt.enc, err)
			continue
		}
		if got != tt.want {
			t.Errorf("encodingABI(%q) = %q, want %q", tt.enc, got, tt.want)
		}
	}

	for _, enc := range []string{"(U=id)", "{S=b3}", "D", "{Opaque}"} {
		if got, err := encodingABI(enc); err == nil {
			t.Errorf("encodingABI(%q) = %q, want an error", enc, got)
		}
	}
}

func TestGoABI(t *testing.T) {
	type point struct {
		_    structs.HostLayout
		X, Y float64
	}
	type rect struct {
		_      structs.HostLayout
		Origin point
		Size   point
	}
	type mixed struct {
		_ structs.HostLayout
		A [3]float32
		B bool
		C *int
	}
	tests := []struct {
		typ  reflect.Type
		want string
	}{
		{reflect.TypeFor[bool](), "1"},
		{reflect.TypeFor[int32](), "4"},
		{reflect.TypeFor[int](), "8"},
		{reflect.TypeFor[ID](), "8"},
		{reflect.TypeFor[Block](), "8"},
		{reflect.TypeFor[float32](), "f"},
		{reflect.TypeFor[rect](), "{dddd}"},
		{reflect.TypeFor[mixed](), "{fff18}"},
	}
	for _, tt := range tests {
		got, err := goABI(tt.typ)
		if err != nil {
			t.Errorf("goABI(%v): %v", tt.typ, err)
			continue
		}
		if got != tt.want {
			t.Errorf("goABI(%v) = %q, want %q", tt.typ, got, tt.want)
		}
	}

	for _, typ := range []reflect.Type{reflect.TypeFor[func()](), reflect.TypeFor[any](), reflect.TypeFor[[]int](), reflect.TypeFor[map[int]int]()} {
		if got, err := goABI(typ); err == nil {
			t.Errorf("goABI(%v) = %q, want an error", typ, got)
		}
	}
}
